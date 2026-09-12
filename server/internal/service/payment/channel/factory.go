package channel

// PaymentMethod 支付方式
const (
	MethodWeChat   int8 = 1 // 微信支付
	MethodAlipay   int8 = 2 // 支付宝
	MethodUnionPay int8 = 3 // 银联
)

// GetChannel 根据支付方式返回对应渠道实现。
// 当前统一返回 MockChannel（真实微信/支付宝/银联 SDK 待接入时在此扩展）。
func GetChannel(method int8) PaymentChannel {
	switch method {
	case MethodWeChat, MethodAlipay, MethodUnionPay:
		// TODO: 接入真实渠道时，按 method 返回 WechatChannel / AlipayChannel / UnionPayChannel
		return NewMockChannel()
	default:
		return NewMockChannel()
	}
}
