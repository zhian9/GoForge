package channel

import (
	"context"
	"fmt"
	"time"
)

// MockChannel 模拟支付渠道，用于开发/联调环境跑通完整支付流程。
// 不调用任何真实第三方，支付结果由 PaymentCallback 显式驱动。
type MockChannel struct{}

// NewMockChannel 创建模拟支付渠道
func NewMockChannel() *MockChannel {
	return &MockChannel{}
}

// CreatePayment 模拟预下单：返回一个本地 mock 支付链接。
func (m *MockChannel) CreatePayment(_ context.Context, req *CreateRequest) (*CreateResponse, error) {
	return &CreateResponse{
		PayStr:       fmt.Sprintf("/mock-pay?payment_no=%s", req.PaymentNo),
		ThirdPartyNo: "",
	}, nil
}

// VerifyCallback 模拟回调验签：不验签，直接通过。
func (m *MockChannel) VerifyCallback(_ context.Context, _ string) error {
	return nil
}

// Refund 模拟退款：返回一个 mock 退款单号。
func (m *MockChannel) Refund(_ context.Context, req *RefundRequest) (*RefundResponse, error) {
	return &RefundResponse{
		ThirdPartyRefundNo: fmt.Sprintf("MOCK-REFUND-%d-%s", time.Now().Unix(), req.PaymentNo),
	}, nil
}
