package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/payment/channel"
	"github.com/zhian9/GoForge/server/internal/service/payment/model"
	"github.com/zhian9/GoForge/server/internal/service/payment/repository"
	"gorm.io/gorm"
)

// PaymentLogic 支付业务逻辑
type PaymentLogic struct {
	paymentRepo    repository.PaymentRepository
	paymentLogRepo repository.PaymentLogRepository
	cache          *cache.CacheOperations
	mqProducer     *mq.Producer
	db             *gorm.DB
}

// NewPaymentLogic 创建支付业务逻辑
func NewPaymentLogic(
	paymentRepo repository.PaymentRepository,
	paymentLogRepo repository.PaymentLogRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
	db *gorm.DB,
) *PaymentLogic {
	return &PaymentLogic{
		paymentRepo:    paymentRepo,
		paymentLogRepo: paymentLogRepo,
		cache:          cache,
		mqProducer:     mqProducer,
		db:             db,
	}
}

// orderStatusPending 订单待支付状态（避免跨服务 import order model）
const orderStatusPending = 1

// validateOrderForPayment 校验订单存在、状态为待支付、金额一致
func (l *PaymentLogic) validateOrderForPayment(ctx context.Context, req *CreatePaymentRequest) error {
	if l.db == nil {
		return nil
	}
	var count int64
	if err := l.db.WithContext(ctx).Table("orders").Where("id = ?", req.OrderID).Count(&count).Error; err != nil {
		return apperrors.NewInternalError("查询订单失败")
	}
	if count == 0 {
		return apperrors.NewError(apperrors.CodeOrderNotFound, "订单不存在")
	}
	var order struct {
		Status    int8    `gorm:"column:status"`
		PayAmount float64 `gorm:"column:pay_amount"`
	}
	if err := l.db.WithContext(ctx).Table("orders").Select("status, pay_amount").Where("id = ?", req.OrderID).Scan(&order).Error; err != nil {
		return apperrors.NewInternalError("查询订单失败")
	}
	if order.Status != orderStatusPending {
		return apperrors.NewError(apperrors.CodeForbidden, "订单状态不允许支付")
	}
	if math.Abs(order.PayAmount-req.Amount) > 0.01 {
		return apperrors.NewError(apperrors.CodeInvalidParam, "支付金额与订单金额不一致")
	}
	return nil
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
	// 校验订单状态与金额，防止篡改金额或对非待支付订单发起支付
	if err := l.validateOrderForPayment(ctx, req); err != nil {
		return nil, err
	}

	// 生成唯一支付单号（Redis INCR）
	paymentNo := l.generatePaymentNo(ctx)

	// 设置过期时间（30分钟）
	expireAt := time.Now().Add(30 * time.Minute)

	// 通过渠道预下单，获取支付串（Mock 渠道返回本地链接）
	ch := channel.GetChannel(req.PaymentMethod)
	channelResp, err := ch.CreatePayment(ctx, &channel.CreateRequest{
		PaymentNo: paymentNo,
		OrderNo:   req.OrderNo,
		Amount:    req.Amount,
		Subject:   fmt.Sprintf("订单-%s", req.OrderNo),
		Method:    req.PaymentMethod,
	})
	if err != nil {
		return nil, apperrors.NewInternalError("渠道预下单失败: " + err.Error())
	}

	payment := &model.Payment{
		PaymentNo:     paymentNo,
		OrderID:       req.OrderID,
		OrderNo:       req.OrderNo,
		UserID:        req.UserID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        model.PaymentStatusPending,
		ExpireAt:      &expireAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if channelResp.ThirdPartyNo != "" {
		payment.ThirdPartyNo = &channelResp.ThirdPartyNo
	}

	if err := l.paymentRepo.Create(ctx, payment); err != nil {
		return nil, apperrors.NewInternalError("创建支付单失败")
	}

	// 记录支付流水
	_ = l.writeLog(ctx, payment, "create", req.Amount, nil, &payment.Status, nil)

	return &CreatePaymentResponse{
		Payment: payment,
		PayURL:  channelResp.PayStr,
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	return &GetPaymentResponse{Payment: payment}, nil
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PaymentNo    string
	ThirdPartyNo string
	Status       int8
	CallbackData string
}

// PaymentCallback 支付回调处理（第三方异步通知入口）
func (l *PaymentLogic) PaymentCallback(ctx context.Context, req *PaymentCallbackRequest) error {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("支付单不存在")
		}
		return apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return apperrors.NewNotFoundError("支付单不存在")
	}

	// 幂等：已成功或已退款，直接返回成功，避免重复处理
	if payment.Status == model.PaymentStatusSuccess || payment.Status == model.PaymentStatusRefunded {
		return nil
	}

	// 渠道验签（Mock 不验签）
	ch := channel.GetChannel(payment.PaymentMethod)
	if err := ch.VerifyCallback(ctx, req.CallbackData); err != nil {
		return apperrors.NewError(6003, "回调验签失败: "+err.Error())
	}

	now := time.Now()
	before := payment.Status

	if req.Status == model.PaymentStatusSuccess {
		// 余额支付：先原子扣余额（防透支）
		if payment.PaymentMethod == model.PaymentMethodBalance {
			if err := l.deductBalance(ctx, payment.UserID, payment.Amount); err != nil {
				return err
			}
		}

		// 原子更新：仅在待支付状态下更新为成功
		updates := map[string]interface{}{
			"status":    model.PaymentStatusSuccess,
			"paid_at":   now,
			"updated_at": now,
		}
		if req.ThirdPartyNo != "" {
			updates["third_party_no"] = req.ThirdPartyNo
		}
		affected, err := l.paymentRepo.UpdateStatusAtomic(ctx, req.PaymentNo, model.PaymentStatusPending, updates)
		if err != nil {
			// 状态更新失败，回退已扣余额
			if payment.PaymentMethod == model.PaymentMethodBalance {
				l.refundBalance(ctx, payment.UserID, payment.Amount)
			}
			return apperrors.NewInternalError("更新支付状态失败")
		}
		if !affected {
			// 已被并发处理，回退已扣余额（幂等）
			if payment.PaymentMethod == model.PaymentMethodBalance {
				l.refundBalance(ctx, payment.UserID, payment.Amount)
			}
			return nil
		}

		payment.Status = model.PaymentStatusSuccess
		payment.PaidAt = &now
		payment.ThirdPartyNo = &req.ThirdPartyNo

		_ = l.writeLog(ctx, payment, "pay", payment.Amount, &before, &payment.Status, nil)

		// 发布支付成功事件，联动订单服务
		l.publishEvent(ctx, mq.TopicPaymentSuccess, payment.PaymentNo, map[string]interface{}{
			"order_id":       payment.OrderID,
			"order_no":       payment.OrderNo,
			"payment_no":     payment.PaymentNo,
			"payment_method": payment.PaymentMethod,
			"amount":         payment.Amount,
			"paid_at":        now.Format(time.RFC3339),
		})
		return nil
	}

	// 支付失败
	updates := map[string]interface{}{
		"status":     model.PaymentStatusFailed,
		"updated_at": now,
	}
	affected, err := l.paymentRepo.UpdateStatusAtomic(ctx, req.PaymentNo, model.PaymentStatusPending, updates)
	if err != nil {
		return apperrors.NewInternalError("更新支付状态失败")
	}
	if !affected {
		return nil
	}
	payment.Status = model.PaymentStatusFailed
	_ = l.writeLog(ctx, payment, "pay_fail", payment.Amount, &before, &payment.Status, nil)
	l.publishEvent(ctx, mq.TopicPaymentFailed, payment.PaymentNo, map[string]interface{}{
		"order_id":   payment.OrderID,
		"order_no":   payment.OrderNo,
		"payment_no": payment.PaymentNo,
	})

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	if payment.Status != model.PaymentStatusSuccess {
		return nil, apperrors.NewError(6002, "支付单状态不允许退款")
	}

	// 生成退款单号
	refundNo := fmt.Sprintf("R%d%06d", time.Now().Unix(), time.Now().UnixNano()%1000000)

	// 渠道发起退款
	ch := channel.GetChannel(payment.PaymentMethod)
	refundResp, err := ch.Refund(ctx, &channel.RefundRequest{
		PaymentNo: req.PaymentNo,
		RefundNo:  refundNo,
		Amount:    req.RefundAmount,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, apperrors.NewInternalError("渠道退款失败: " + err.Error())
	}

	// 原子更新：仅在支付成功状态下更新为已退款
	updates := map[string]interface{}{
		"status":     model.PaymentStatusRefunded,
		"updated_at": time.Now(),
	}
	affected, err := l.paymentRepo.UpdateStatusAtomic(ctx, req.PaymentNo, model.PaymentStatusSuccess, updates)
	if err != nil {
		return nil, apperrors.NewInternalError("更新退款状态失败")
	}
	if !affected {
		return nil, apperrors.NewError(6004, "退款已处理")
	}

	before := model.PaymentStatusSuccess
	after := model.PaymentStatusRefunded
	_ = l.writeLog(ctx, payment, "refund", req.RefundAmount, &before, &after, &req.Reason)

	// 余额支付退款：回退余额
	if payment.PaymentMethod == model.PaymentMethodBalance {
		l.refundBalance(ctx, payment.UserID, req.RefundAmount)
	}

	l.publishEvent(ctx, mq.TopicPaymentRefunded, payment.PaymentNo, map[string]interface{}{
		"order_id":      payment.OrderID,
		"order_no":      payment.OrderNo,
		"payment_no":    payment.PaymentNo,
		"refund_no":     refundNo,
		"refund_amount": req.RefundAmount,
		"third_refund_no": refundResp.ThirdPartyRefundNo,
	})

	return &RefundResponse{RefundNo: refundNo}, nil
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("支付单不存在")
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewNotFoundError("支付单不存在")
	}

	return &QueryPaymentStatusResponse{Status: payment.Status}, nil
}

// generatePaymentNo 生成唯一支付单号（Redis INCR 保证唯一）
func (l *PaymentLogic) generatePaymentNo(ctx context.Context) string {
	if l.cache != nil {
		today := time.Now().Format("20060102")
		seqKey := cache.BuildKey(cache.KeyPrefixPaymentSeq, today)
		seq, err := l.cache.Increment(ctx, seqKey)
		if err == nil {
			_ = l.cache.Expire(ctx, seqKey, 24*time.Hour)
			return fmt.Sprintf("PAY%s%06d", today, seq)
		}
	}
	now := time.Now()
	return fmt.Sprintf("PAY%s%06d", now.Format("20060102"), now.Nanosecond()%1000000)
}

// writeLog 记录支付流水（失败不影响主流程）
func (l *PaymentLogic) writeLog(ctx context.Context, payment *model.Payment, action string, amount float64, before, after *int8, remark *string) error {
	log := &model.PaymentLog{
		PaymentID:    payment.ID,
		PaymentNo:    payment.PaymentNo,
		Action:       action,
		Amount:       amount,
		BeforeStatus: before,
		AfterStatus:  after,
		CreatedAt:    time.Now(),
	}
	if remark != nil {
		log.Remark = *remark
	}
	return l.paymentLogRepo.Create(ctx, log)
}

// publishEvent 发布支付事件
func (l *PaymentLogic) publishEvent(ctx context.Context, topic, key string, data map[string]interface{}) {
	if l.mqProducer == nil {
		return
	}
	message := mq.NewMessage(topic, data)
	_ = l.mqProducer.PublishWithKey(ctx, topic, key, message)
}

// deductBalance 原子扣减余额（防透支）
func (l *PaymentLogic) deductBalance(ctx context.Context, userID uint64, amount float64) error {
	if l.db == nil {
		return apperrors.NewInternalError("数据库未初始化")
	}
	res := l.db.WithContext(ctx).Exec(
		"UPDATE user SET balance = balance - ? WHERE id = ? AND balance >= ?",
		amount, userID, amount,
	)
	if res.Error != nil {
		return apperrors.NewInternalError("扣减余额失败")
	}
	if res.RowsAffected == 0 {
		return apperrors.NewError(apperrors.CodeInvalidParam, "余额不足")
	}
	// 记余额流水（支付支出）
	l.writeBalanceLog(ctx, userID, 2, -amount, "余额支付")
	return nil
}

// refundBalance 回退余额
func (l *PaymentLogic) refundBalance(ctx context.Context, userID uint64, amount float64) {
	if l.db == nil {
		return
	}
	_ = l.db.WithContext(ctx).Exec("UPDATE user SET balance = balance + ? WHERE id = ?", amount, userID).Error
	// 记余额流水（退款收入）
	l.writeBalanceLog(ctx, userID, 3, amount, "退款回退")
}

// writeBalanceLog 记余额流水（amount 带符号：正增负减）
func (l *PaymentLogic) writeBalanceLog(ctx context.Context, userID uint64, logType int8, amount float64, remark string) {
	if l.db == nil {
		return
	}
	var balance float64
	if err := l.db.WithContext(ctx).Table("user").Select("balance").Where("id = ?", userID).Scan(&balance).Error; err != nil {
		return
	}
	_ = l.db.WithContext(ctx).Exec(
		"INSERT INTO balance_log (user_id, order_no, type, amount, before_balance, after_balance, remark, created_at) VALUES (?, '', ?, ?, ?, ?, ?, ?)",
		userID, logType, amount, balance-amount, balance, remark, time.Now(),
	).Error
}
