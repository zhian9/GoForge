package mq

// Kafka Topic定义
const (
	// 订单相关
	TopicOrderCreated   = "order.created"
	TopicOrderCancelled = "order.cancelled"

	// 库存相关
	TopicInventoryDeducted = "inventory.deducted"

	// 支付相关
	TopicPaymentSuccess  = "payment-service.success"
	TopicPaymentFailed   = "payment-service.failed"
	TopicPaymentRefunded = "payment-service.refunded"

	// 秒杀相关
	TopicSeckillOrder = "seckill.order" // 秒杀订单消息

	// 数据同步
	TopicDataSync = "data.sync"
)
