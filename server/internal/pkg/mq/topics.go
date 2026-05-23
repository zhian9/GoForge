package mq

// Kafka Topic定义
const (
	// 订单相关
	TopicOrderCreated   = "order.created"
	TopicOrderPaid      = "order.paid"
	TopicOrderCancelled = "order.cancelled"
	TopicOrderCompleted = "order.completed"

	// 库存相关
	TopicInventoryDeducted = "inventory.deducted"
	TopicInventoryAlert    = "inventory.alert"

	// 用户行为
	TopicUserView     = "user.view"
	TopicUserPurchase = "user.purchase"
	TopicUserFavorite = "user.favorite"

	// 商品相关
	TopicProductOnline       = "product.online"
	TopicProductOffline      = "product.offline"
	TopicProductPriceChanged = "product.price.changed"

	// 支付相关
	TopicPaymentSuccess  = "payment-service.success"
	TopicPaymentFailed   = "payment-service.failed"
	TopicPaymentRefunded = "payment-service.refunded"

	// 营销相关
	TopicCouponIssued   = "coupon.issued"
	TopicCouponUsed     = "coupon.used"
	TopicSeckillStarted = "seckill.started"

	// 秒杀相关
	TopicSeckillOrder = "seckill.order" // 秒杀订单消息

	// 物流相关
	TopicLogisticsUpdated   = "logistics.updated"
	TopicLogisticsDelivered = "logistics.delivered"

	// 系统消息
	TopicSystemNotification = "system.notification"
	TopicDataSync           = "data.sync"
)
