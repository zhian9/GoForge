package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
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

// GatewayRateLimitMiddleware 网关级限流中间件（按客户端 IP 的令牌桶）。
//
// 为什么单独写一个：上面三个中间件的类型是 func(http.Handler) http.Handler（标准库风格），
// 而 go-zero 网关的 Use() 要求 func(http.HandlerFunc) http.HandlerFunc —— 类型对不上，
// 所以它们虽然写好了却一直没法挂到网关上（全项目 0 引用）。
// 这个适配器让限流真正生效。
//
// 维度选择：只按客户端 IP 限流。按用户限流需要解析 token 拿到 user_id，
// 那是鉴权中间件的职责；网关这一层用 IP 做粗粒度兜底即可，
// 更细的维度（用户、接口）交给各服务自己控制。
//
// 注意：经过 Nginx / Cloudflare 时 RemoteAddr 是代理 IP，
// 所以优先取 X-Forwarded-For / X-Real-IP（Nginx 已配置真实 IP 还原）。
func GatewayRateLimitMiddleware(qps, burst int) func(http.HandlerFunc) http.HandlerFunc {
	limiter := NewRateLimiter()

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !limiter.GetLimiter(ClientIP(r), qps, burst).Allow() {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Retry-After", strconv.Itoa(1))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":429,"message":"请求过于频繁，请稍后再试"}`))
				return
			}
			next(w, r)
		}
	}
}

// ClientIP 从请求中还原客户端 IP，优先信任代理透传的头。
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能是 "client, proxy1, proxy2"，取第一个
		if idx := strings.IndexByte(xff, ','); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
