package monitoring

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics Prometheus指标
var (
	// HTTP请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"service", "method", "path", "status"},
	)

	// HTTP请求延迟
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求延迟（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)

	// gRPC请求总数
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "gRPC请求总数",
		},
		[]string{"service", "method", "status"},
	)

	// gRPC请求延迟
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC请求延迟（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// 业务指标：订单创建
	OrdersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "订单创建总数",
		},
		[]string{"service"},
	)

	// 业务指标：支付成功
	PaymentsSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_success_total",
			Help: "支付成功总数",
		},
		[]string{"service", "payment_type"},
	)

	// 业务指标：库存扣减
	InventoryDeductedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_deducted_total",
			Help: "库存扣减总数",
		},
		[]string{"service"},
	)

	// 数据库连接数
	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "数据库连接数",
		},
		[]string{"service", "state"}, // state: idle, in_use, max
	)

	// Redis连接数
	RedisConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Redis连接数",
		},
		[]string{"service", "state"},
	)

	// Kafka消息生产
	KafkaMessagesProduced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "Kafka消息生产总数",
		},
		[]string{"service", "topic"},
	)

	// Kafka消息消费
	KafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Kafka消息消费总数",
		},
		[]string{"service", "topic"},
	)
)

// RecordHTTPRequest 记录HTTP请求指标
func RecordHTTPRequest(service, method, path string, status int, duration float64) {
	HTTPRequestsTotal.WithLabelValues(service, method, path, fmt.Sprintf("%d", status)).Inc()
	HTTPRequestDuration.WithLabelValues(service, method, path).Observe(duration)
}

// RecordGRPCRequest 记录gRPC请求指标
func RecordGRPCRequest(service, method string, status string, duration float64) {
	GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
	GRPCRequestDuration.WithLabelValues(service, method).Observe(duration)
}
