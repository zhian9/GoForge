package governance

import (
	"github.com/zeromicro/go-zero/core/breaker"
)

// BreakerConfig 断路器配置
type BreakerConfig struct {
	Name        string  // 断路器名称
	ErrorRate   float64 // 错误率阈值（0-1）
	MinRequests int     // 最小请求数
	Timeout     int     // 超时时间（秒）
	MaxRequests int     // 半开状态最大请求数
	Interval    int     // 统计窗口时间（秒）
}

// CircuitBreaker 断路器封装
type CircuitBreaker struct {
	breaker breaker.Breaker
	config  BreakerConfig
}

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(cfg BreakerConfig) *CircuitBreaker {
	// go-zero的breaker使用简化配置
	// K: 错误率阈值（标准差倍数，通常5-10）
	// Request: 最小请求数
	// Timeout: 超时时间
	brk := breaker.NewBreaker(breaker.WithName(cfg.Name))

	return &CircuitBreaker{
		breaker: brk,
		config:  cfg,
	}
}

// Do 执行操作（带断路器保护）
func (cb *CircuitBreaker) Do(fn func() error) error {
	return breaker.Do(cb.config.Name, fn)
}

// DoWithAcceptable 执行操作（自定义可接受错误）
func (cb *CircuitBreaker) DoWithAcceptable(fn func() error, acceptable func(err error) bool) error {
	return breaker.DoWithAcceptable(cb.config.Name, fn, acceptable)
}

// DoWithFallback 执行操作（带降级处理）
func (cb *CircuitBreaker) DoWithFallback(fn func() error, fallback func(err error) error) error {
	return breaker.DoWithFallback(cb.config.Name, fn, fallback)
}

// IsOpen 检查断路器是否打开
func (cb *CircuitBreaker) IsOpen() bool {
	// go-zero的breaker使用Promise模式
	promise, err := cb.breaker.Allow()
	if err != nil {
		return true // 断路器打开
	}
	promise.Accept() // 立即接受，避免影响统计
	return false
}

// BreakerManager 断路器管理器
type BreakerManager struct {
	breakers map[string]*CircuitBreaker
}

var globalBreakerManager *BreakerManager

// InitBreakerManager 初始化断路器管理器
func InitBreakerManager() {
	globalBreakerManager = &BreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetBreaker 获取或创建断路器
func GetBreaker(name string, cfg BreakerConfig) *CircuitBreaker {
	if globalBreakerManager == nil {
		InitBreakerManager()
	}

	if breaker, ok := globalBreakerManager.breakers[name]; ok {
		return breaker
	}

	breaker := NewCircuitBreaker(cfg)
	globalBreakerManager.breakers[name] = breaker
	return breaker
}

// ServiceBreaker 服务调用断路器
type ServiceBreaker struct {
	userService      *CircuitBreaker
	productService   *CircuitBreaker
	orderService     *CircuitBreaker
	paymentService   *CircuitBreaker
	inventoryService *CircuitBreaker
}

// NewServiceBreaker 创建服务断路器
func NewServiceBreaker() *ServiceBreaker {
	return &ServiceBreaker{
		userService: GetBreaker("user-service", BreakerConfig{
			Name:        "user-service",
			ErrorRate:   0.5,
			MinRequests: 100,
			Timeout:     60,
			MaxRequests: 10,
			Interval:    60,
		}),
		productService: GetBreaker("product-service", BreakerConfig{
			Name:        "product-service",
			ErrorRate:   0.5,
			MinRequests: 100,
			Timeout:     60,
			MaxRequests: 10,
			Interval:    60,
		}),
		orderService: GetBreaker("order-service", BreakerConfig{
			Name:        "order-service",
			ErrorRate:   0.3,
			MinRequests: 50,
			Timeout:     60,
			MaxRequests: 10,
			Interval:    60,
		}),
		paymentService: GetBreaker("payment-service-service", BreakerConfig{
			Name:        "payment-service-service",
			ErrorRate:   0.2,
			MinRequests: 50,
			Timeout:     60,
			MaxRequests: 10,
			Interval:    60,
		}),
		inventoryService: GetBreaker("inventory-service", BreakerConfig{
			Name:        "inventory-service",
			ErrorRate:   0.5,
			MinRequests: 100,
			Timeout:     60,
			MaxRequests: 10,
			Interval:    60,
		}),
	}
}

// CallUserService 调用用户服务（带断路器保护）
func (sb *ServiceBreaker) CallUserService(fn func() error) error {
	return sb.userService.Do(fn)
}

// CallProductService 调用商品服务（带断路器保护）
func (sb *ServiceBreaker) CallProductService(fn func() error) error {
	return sb.productService.Do(fn)
}

// CallOrderService 调用订单服务（带断路器保护）
func (sb *ServiceBreaker) CallOrderService(fn func() error) error {
	return sb.orderService.Do(fn)
}

// CallPaymentService 调用支付服务（带断路器保护）
func (sb *ServiceBreaker) CallPaymentService(fn func() error) error {
	return sb.paymentService.Do(fn)
}

// CallInventoryService 调用库存服务（带断路器保护）
func (sb *ServiceBreaker) CallInventoryService(fn func() error) error {
	return sb.inventoryService.Do(fn)
}

// IsOpenError 检查是否为断路器打开错误
func IsOpenError(err error) bool {
	return err == breaker.ErrServiceUnavailable
}
