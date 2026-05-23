package constants

// 用户相关常量
const (
	// 用户状态
	UserStatusDisabled = 0
	UserStatusNormal   = 1

	// 会员等级
	MemberLevelNormal = 0
	MemberLevelVIP1   = 1
	MemberLevelVIP2   = 2
	MemberLevelVIP3   = 3

	// 性别
	GenderUnknown = 0
	GenderMale    = 1
	GenderFemale  = 2
)

// 商品相关常量
const (
	// 商品状态
	ProductStatusOffline = 0
	ProductStatusOnline  = 1
	ProductStatusPending = 2

	// SKU状态
	SkuStatusOffline = 0
	SkuStatusOnline  = 1
)

// 订单相关常量
const (
	// 订单类型
	OrderTypeNormal  = 1
	OrderTypeSeckill = 2
	OrderTypeGroup   = 3

	// 订单状态
	OrderStatusCanceled  = 0
	OrderStatusPending   = 1
	OrderStatusPaid      = 2
	OrderStatusShipped   = 3
	OrderStatusReceived  = 4
	OrderStatusCompleted = 5
	OrderStatusRefunded  = 6

	// 支付方式
	PaymentMethodWeChat   = 1
	PaymentMethodAlipay   = 2
	PaymentMethodUnionPay = 3
)

// 库存相关常量
const (
	// 库存操作类型
	InventoryTypeIn       = 1 // 入库
	InventoryTypeOut      = 2 // 出库
	InventoryTypeLock     = 3 // 锁定
	InventoryTypeUnlock   = 4 // 解锁
	InventoryTypeDeduct   = 5 // 扣减
	InventoryTypeRollback = 6 // 回退
)

// 支付相关常量
const (
	// 支付状态
	PaymentStatusPending  = 0
	PaymentStatusSuccess  = 1
	PaymentStatusFailed   = 2
	PaymentStatusRefunded = 3
)

// 优惠券相关常量
const (
	// 优惠券类型
	CouponTypeFullReduction = 1 // 满减券
	CouponTypeDiscount      = 2 // 折扣券
	CouponTypeFreeShipping  = 3 // 免运费券

	// 用户优惠券状态
	UserCouponStatusUnused  = 0
	UserCouponStatusUsed    = 1
	UserCouponStatusExpired = 2
)

// 评价相关常量
const (
	// 评价状态
	ReviewStatusHidden  = 0
	ReviewStatusVisible = 1
)

// 物流相关常量
const (
	// 物流状态
	LogisticsStatusPending   = 0
	LogisticsStatusShipped   = 1
	LogisticsStatusInTransit = 2
	LogisticsStatusDelivered = 3
	LogisticsStatusException = 4
)

// 消息相关常量
const (
	// 消息类型
	MessageTypeSystem    = 1
	MessageTypeOrder     = 2
	MessageTypePromotion = 3
	MessageTypeLogistics = 4
)

// 分页相关常量
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)
