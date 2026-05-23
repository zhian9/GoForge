package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/payment/model"
	"github.com/zhian9/GoForge/server/internal/service/payment/repository"
	"gorm.io/gorm"
)

// PaymentLogic 支付业务逻辑
type PaymentLogic struct {
	paymentRepo    repository.PaymentRepository
	paymentLogRepo repository.PaymentLogRepository
}

// NewPaymentLogic 创建支付业务逻辑
func NewPaymentLogic(
	paymentRepo repository.PaymentRepository,
	paymentLogRepo repository.PaymentLogRepository,
) *PaymentLogic {
	return &PaymentLogic{
		paymentRepo:    paymentRepo,
		paymentLogRepo: paymentLogRepo,
	}
}

// CreatePaymentRequest 创建支付单请求
type CreatePaymentRequest struct {
	OrderID       uint64
	OrderNo       string
	UserID        uint64
	Amount        float64
	PaymentMethod int8
}

// CreatePaymentResponse 创建支付单响应
type CreatePaymentResponse struct {
	Payment *model.Payment
	PayURL  string
}

// CreatePayment 创建支付单
func (l *PaymentLogic) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 生成支付单号
	paymentNo := fmt.Sprintf("P%d%d", time.Now().Unix(), req.UserID)

	// 设置过期时间（30分钟）
	expireAt := time.Now().Add(30 * time.Minute)

	payment := &model.Payment{
		PaymentNo:     paymentNo,
		OrderID:       req.OrderID,
		OrderNo:       req.OrderNo,
		UserID:        req.UserID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        0, // 待支付
		ExpireAt:      &expireAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := l.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, apperrors.NewInternalError("创建支付单失败")
	}

	// 记录支付流水
	log := &model.PaymentLog{
		PaymentID:   payment.ID,
		PaymentNo:   paymentNo,
		Action:      "create",
		Amount:      req.Amount,
		AfterStatus: &payment.Status,
		CreatedAt:   time.Now(),
	}
	_ = l.paymentLogRepo.Create(ctx, log)

	// 生成支付链接（这里简化处理，实际应该调用第三方支付API）
	payURL := fmt.Sprintf("/payment-service/pay?payment_no=%s", paymentNo)

	return &CreatePaymentResponse{
		Payment: payment,
		PayURL:  payURL,
	}, nil
}

// GetPaymentRequest 获取支付单请求
type GetPaymentRequest struct {
	PaymentNo string
}

// GetPaymentResponse 获取支付单响应
type GetPaymentResponse struct {
	Payment *model.Payment
}

// GetPayment 获取支付单
func (l *PaymentLogic) GetPayment(ctx context.Context, req *GetPaymentRequest) (*GetPaymentResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	return &GetPaymentResponse{
		Payment: payment,
	}, nil
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PaymentNo    string
	ThirdPartyNo string
	Status       int8
	CallbackData string
}

// PaymentCallback 支付回调处理
func (l *PaymentLogic) PaymentCallback(ctx context.Context, req *PaymentCallbackRequest) error {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("支付单不存在")
		}
		return apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return apperrors.NewNotFoundError("支付单不存在")
	}

	// 更新支付状态
	now := time.Now()
	payment.Status = req.Status
	payment.ThirdPartyNo = &req.ThirdPartyNo
	if req.Status == 1 {
		payment.PaidAt = &now
	}

	err = l.paymentRepo.Update(ctx, payment)
	if err != nil {
		return apperrors.NewInternalError("更新支付状态失败")
	}

	// 记录支付流水
	log := &model.PaymentLog{
		PaymentID:    payment.ID,
		PaymentNo:    req.PaymentNo,
		Action:       "pay",
		Amount:       payment.Amount,
		BeforeStatus: &[]int8{0}[0],
		AfterStatus:  &req.Status,
		CreatedAt:    time.Now(),
	}
	_ = l.paymentLogRepo.Create(ctx, log)

	return nil
}

// RefundRequest 退款请求
type RefundRequest struct {
	PaymentNo    string
	RefundAmount float64
	Reason       string
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo string
}

// Refund 申请退款
func (l *PaymentLogic) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	if payment.Status != 1 {
		return nil, apperrors.NewError(6002, "支付单状态不允许退款")
	}

	// 生成退款单号
	refundNo := fmt.Sprintf("R%d%d", time.Now().Unix(), payment.UserID)

	// 更新支付状态为已退款
	payment.Status = 3
	err = l.paymentRepo.Update(ctx, payment)
	if err != nil {
		return nil, apperrors.NewInternalError("退款失败")
	}

	// 记录退款流水
	log := &model.PaymentLog{
		PaymentID:    payment.ID,
		PaymentNo:    req.PaymentNo,
		Action:       "refund",
		Amount:       req.RefundAmount,
		BeforeStatus: &[]int8{1}[0],
		AfterStatus:  &[]int8{3}[0],
		Remark:       req.Reason,
		CreatedAt:    time.Now(),
	}
	_ = l.paymentLogRepo.Create(ctx, log)

	return &RefundResponse{
		RefundNo: refundNo,
	}, nil
}

// QueryPaymentStatusRequest 查询支付状态请求
type QueryPaymentStatusRequest struct {
	PaymentNo string
}

// QueryPaymentStatusResponse 查询支付状态响应
type QueryPaymentStatusResponse struct {
	Status int8
}

// QueryPaymentStatus 查询支付状态
func (l *PaymentLogic) QueryPaymentStatus(ctx context.Context, req *QueryPaymentStatusRequest) (*QueryPaymentStatusResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	return &QueryPaymentStatusResponse{
		Status: payment.Status,
	}, nil
}
