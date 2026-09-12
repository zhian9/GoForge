package channel

import "context"

// PaymentChannel 支付渠道统一接口。
// 各支付方式（微信/支付宝/银联）实现该接口，MockChannel 用于开发联调。
type PaymentChannel interface {
	// CreatePayment 预下单，返回支付串（二维码 / H5 跳转链接等）。
	CreatePayment(ctx context.Context, req *CreateRequest) (*CreateResponse, error)

	// VerifyCallback 校验第三方异步回调签名并解析回调内容。
	// Mock 渠道直接返回 nil（不验签），真实渠道在此完成验签。
	VerifyCallback(ctx context.Context, data string) error

	// Refund 发起退款。
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
}

// CreateRequest 预下单请求
type CreateRequest struct {
	PaymentNo string  // 支付单号
	OrderNo   string  // 订单号
	Amount    float64 // 支付金额（元）
	Subject   string  // 商品描述
	Method    int8    // 支付方式
}

// CreateResponse 预下单响应
type CreateResponse struct {
	PayStr       string // 支付串 / 跳转链接
	ThirdPartyNo string // 第三方预下单单号（可为空）
}

// RefundRequest 退款请求
type RefundRequest struct {
	PaymentNo string  // 支付单号
	RefundNo  string  // 退款单号（业务侧生成）
	Amount    float64 // 退款金额（元）
	Reason    string  // 退款原因
}

// RefundResponse 退款响应
type RefundResponse struct {
	ThirdPartyRefundNo string // 第三方退款单号
}
