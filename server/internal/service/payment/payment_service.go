package payment

import (
	"context"
	"strconv"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/payment/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/payment/model"
	"github.com/zhian9/GoForge/server/internal/service/payment/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentService 实现 gRPC 服务接口
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	svcCtx *ServiceContext
	logic  *service.PaymentLogic
}

// NewPaymentService 创建支付服务
func NewPaymentService(svcCtx *ServiceContext) *PaymentService {
	logic := service.NewPaymentLogic(
		svcCtx.PaymentRepo,
		svcCtx.PaymentLogRepo,
		svcCtx.Cache,
		svcCtx.MQProducer,
		svcCtx.DB,
	)

	return &PaymentService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// CreatePayment 创建支付单
func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.CreatePaymentResponse, error) {
	amount, _ := strconv.ParseFloat(req.Amount, 64)

	createReq := &service.CreatePaymentRequest{
		OrderID:       uint64(req.OrderId),
		OrderNo:       req.OrderNo,
		UserID:        uint64(req.UserId),
		Amount:        amount,
		PaymentMethod: int8(req.PaymentMethod),
	}

	resp, err := s.logic.CreatePayment(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CreatePaymentResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertPaymentToProto(resp.Payment),
		PayUrl:  resp.PayURL,
	}, nil
}

// GetPayment 获取支付单
func (s *PaymentService) GetPayment(ctx context.Context, req *v1.GetPaymentRequest) (*v1.GetPaymentResponse, error) {
	getReq := &service.GetPaymentRequest{
		PaymentNo: req.PaymentNo,
	}

	resp, err := s.logic.GetPayment(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetPaymentResponse{
		Code:    0,
		Message: "成功",
		Data:    convertPaymentToProto(resp.Payment),
	}, nil
}

// PaymentCallback 支付回调处理
func (s *PaymentService) PaymentCallback(ctx context.Context, req *v1.PaymentCallbackRequest) (*v1.PaymentCallbackResponse, error) {
	callbackReq := &service.PaymentCallbackRequest{
		PaymentNo:    req.PaymentNo,
		ThirdPartyNo: req.ThirdPartyNo,
		Status:       int8(req.Status),
		CallbackData: req.CallbackData,
	}

	err := s.logic.PaymentCallback(ctx, callbackReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.PaymentCallbackResponse{
		Code:    0,
		Message: "处理成功",
	}, nil
}

// Refund 申请退款
func (s *PaymentService) Refund(ctx context.Context, req *v1.RefundRequest) (*v1.RefundResponse, error) {
	refundAmount, _ := strconv.ParseFloat(req.RefundAmount, 64)

	refundReq := &service.RefundRequest{
		PaymentNo:    req.PaymentNo,
		RefundAmount: refundAmount,
		Reason:       req.Reason,
	}

	resp, err := s.logic.Refund(ctx, refundReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.RefundResponse{
		Code:     0,
		Message:  "退款成功",
		RefundNo: resp.RefundNo,
	}, nil
}

// QueryPaymentStatus 查询支付状态
func (s *PaymentService) QueryPaymentStatus(ctx context.Context, req *v1.QueryPaymentStatusRequest) (*v1.QueryPaymentStatusResponse, error) {
	queryReq := &service.QueryPaymentStatusRequest{
		PaymentNo: req.PaymentNo,
	}

	resp, err := s.logic.QueryPaymentStatus(ctx, queryReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.QueryPaymentStatusResponse{
		Code:    0,
		Message: "成功",
		Status:  int32(resp.Status),
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}

	// 检查是否是 BusinessError
	if bizErr, ok := err.(*apperrors.BusinessError); ok {
		var grpcCode codes.Code
		switch bizErr.Code {
		case apperrors.CodeNotFound:
			grpcCode = codes.NotFound
		case apperrors.CodeInvalidParam:
			grpcCode = codes.InvalidArgument
		case apperrors.CodeUnauthorized:
			grpcCode = codes.Unauthenticated
		case apperrors.CodeForbidden:
			grpcCode = codes.PermissionDenied
		default:
			grpcCode = codes.Internal
		}
		return status.Error(grpcCode, bizErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// convertPaymentToProto 转换支付单模型为 Protobuf 消息
func convertPaymentToProto(payment *model.Payment) *v1.Payment {
	if payment == nil {
		return nil
	}

	var thirdPartyNo string
	if payment.ThirdPartyNo != nil {
		thirdPartyNo = *payment.ThirdPartyNo
	}

	return &v1.Payment{
		Id:            int64(payment.ID),
		PaymentNo:     payment.PaymentNo,
		OrderId:       int64(payment.OrderID),
		OrderNo:       payment.OrderNo,
		UserId:        int64(payment.UserID),
		Amount:        strconv.FormatFloat(payment.Amount, 'f', 2, 64),
		PaymentMethod: int32(payment.PaymentMethod),
		Status:        int32(payment.Status),
		ThirdPartyNo:  thirdPartyNo,
		PaidAt:        formatTime(payment.PaidAt),
		ExpireAt:      formatTime(payment.ExpireAt),
		CreatedAt:     formatTime(&payment.CreatedAt),
		UpdatedAt:     formatTime(&payment.UpdatedAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
