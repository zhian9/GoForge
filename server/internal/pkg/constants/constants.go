package constants

// 用户相关常量
const (
	// 用户状态
	UserStatusNormal = 1

	// 会员等级
	MemberLevelNormal = 0
)

// 订单相关常量
const (
	// 订单类型
	OrderTypeNormal  = 1
	OrderTypeSeckill = 2

	// 订单状态
	OrderStatusCanceled  = 0
	OrderStatusPending   = 1
	OrderStatusPaid      = 2
	OrderStatusShipped   = 3
	OrderStatusCompleted = 5
	OrderStatusRefunded  = 6

	// 支付方式
	PaymentMethodWeChat   = 1
	PaymentMethodAlipay   = 2
	PaymentMethodUnionPay = 3
)

// 支付相关常量
const (
	// 支付状态
	PaymentStatusPending  = 0
	PaymentStatusSuccess  = 1
	PaymentStatusFailed   = 2
	PaymentStatusRefunded = 3
)
