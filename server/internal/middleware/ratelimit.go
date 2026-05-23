package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter 获取或创建限流器
func (rl *RateLimiter) GetLimiter(key string, qps int, burst int) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter, exists = rl.limiters[key]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(qps), burst)
			rl.limiters[key] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter *RateLimiter, qps int, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 使用IP作为限流key
			clientIP := r.RemoteAddr
			l := limiter.GetLimiter(clientIP, qps, burst)

			if !l.Allow() {
				http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GlobalRateLimitMiddleware 全局限流中间件
func GlobalRateLimitMiddleware(qps int, burst int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "系统繁忙，请稍后再试", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PerEndpointRateLimitMiddleware 接口级限流中间件
func PerEndpointRateLimitMiddleware(limiter *RateLimiter, endpointLimits map[string]struct {
	QPS   int
	Burst int
}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if limit, ok := endpointLimits[path]; ok {
				l := limiter.GetLimiter(path, limit.QPS, limit.Burst)
				if !l.Allow() {
					http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
