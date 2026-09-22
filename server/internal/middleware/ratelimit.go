package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

const (
	// defaultLimiterTTL 一个 key（通常是客户端 IP）多久没有新请求后回收其限流器。
	// 没有这个回收，limiters 会随着扫描/爬虫带来的新 IP 一直增长，只增不减。
	defaultLimiterTTL = 10 * time.Minute

	// defaultLimiterCleanupInterval 后台清理协程的检查间隔。
	defaultLimiterCleanupInterval = time.Minute
)

// limiterEntry 包装单个 key 的限流器，并记录最近一次使用时间（原子读写，避免每次请求都加写锁）。
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // UnixNano
}

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	ttl      time.Duration
	stop     chan struct{}
	stopOnce sync.Once
}

// NewRateLimiter 创建限流器
func NewRateLimiter() *RateLimiter {
	return NewRateLimiterWithTTL(defaultLimiterTTL)
}

// NewRateLimiterWithTTL 创建限流器并指定空闲回收时间。
// ttl <= 0 时使用默认值。清理协程随限流器一起启动，进程退出/调用 Close 后停止。
func NewRateLimiterWithTTL(ttl time.Duration) *RateLimiter {
	if ttl <= 0 {
		ttl = defaultLimiterTTL
	}
	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		ttl:      ttl,
		stop:     make(chan struct{}),
	}
	go rl.janitor(defaultLimiterCleanupInterval)
	return rl
}

// janitor 定期回收长时间没有请求的 key。
func (rl *RateLimiter) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			if n := rl.evictExpired(); n > 0 {
				logx.Infof("限流器回收空闲 key: %d 个，剩余 %d 个", n, rl.Len())
			}
		}
	}
}

// evictExpired 删除超过 ttl 未被访问的限流器，返回删除数量。
func (rl *RateLimiter) evictExpired() int {
	deadline := time.Now().Add(-rl.ttl).UnixNano()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	removed := 0
	for key, entry := range rl.limiters {
		if entry.lastSeen.Load() < deadline {
			delete(rl.limiters, key)
			removed++
		}
	}
	return removed
}

// Len 返回当前持有多少个 key 的限流器（用于观测与测试）。
func (rl *RateLimiter) Len() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// Close 停止后台清理协程。重复调用安全。
func (rl *RateLimiter) Close() {
	rl.stopOnce.Do(func() {
		close(rl.stop)
	})
}

// GetLimiter 获取或创建限流器
func (rl *RateLimiter) GetLimiter(key string, qps int, burst int) *rate.Limiter {
	now := time.Now().UnixNano()

	rl.mu.RLock()
	entry := rl.limiters[key]
	rl.mu.RUnlock()

	if entry != nil {
		entry.lastSeen.Store(now)
		return entry.limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双检：可能在拿写锁期间被其它 goroutine 创建
	if entry = rl.limiters[key]; entry != nil {
		entry.lastSeen.Store(now)
		return entry.limiter
	}

	entry = &limiterEntry{limiter: rate.NewLimiter(rate.Limit(qps), burst)}
	entry.lastSeen.Store(now)
	rl.limiters[key] = entry
	return entry.limiter
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter *RateLimiter, qps int, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 使用IP作为限流key
			clientIP := ClientIP(r)
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
// 注意：经过 Nginx 时 RemoteAddr 是代理 IP，此时才允许读取 X-Real-IP /
// X-Forwarded-For 还原真实客户端（见 ClientIP）。如果网关端口被直连，
// 请求头里的转发信息一律忽略，避免伪造 header 绕过限流。
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

// trustedProxyNets 只有直连对端属于这些网段时，才认为请求来自我们自己的
// 反向代理（Nginx / Ingress / 本机 / Docker、K8s 内网），也就是才允许相信
// X-Forwarded-For / X-Real-IP 这类客户端可自行伪造的请求头。
var trustedProxyNets = func() []*net.IPNet {
	cidrs := []string{
		"127.0.0.0/8",    // 本机 IPv4
		"::1/128",        // 本机 IPv6
		"10.0.0.0/8",     // 内网 / K8s Pod 网段
		"172.16.0.0/12",  // 内网 / Docker bridge
		"192.168.0.0/16", // 内网
		"fc00::/7",       // IPv6 唯一本地地址
	}

	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}()

// isTrustedProxy 判断直连对端是否是可信任的代理地址。
func isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, ipNet := range trustedProxyNets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// remoteIP 解析 RemoteAddr（形如 "1.2.3.4:5678"）中的 IP。
func remoteIP(remoteAddr string) net.IP {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(strings.TrimSpace(remoteAddr))
}

// parseIPHeader 解析请求头中的 IP，允许带端口（如 "1.2.3.4:5678"）的形式。
func parseIPHeader(value string) net.IP {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		return net.ParseIP(host)
	}
	return nil
}

// ClientIP 还原客户端 IP，用于限流 key。
//
// 这里的原则是「不信任客户端能自己设置的东西」：
//   - 直连对端不是我们的代理时，一律用 RemoteAddr，忽略 X-Forwarded-For / X-Real-IP。
//     否则任何人只要伪造一个 header，每个请求都会被算成不同的 IP，限流形同虚设。
//   - 直连对端是可信代理时，优先用 X-Real-IP：GoForge 的 Nginx 用 $remote_addr
//     覆盖写入这个头，客户端塞不进来。
//   - 退回解析 X-Forwarded-For 时从右往左找第一个非可信地址：最左边那个值
//     可能是客户端自己加的（$proxy_add_x_forwarded_for 会保留它），从左边取会被伪造。
func ClientIP(r *http.Request) string {
	peer := remoteIP(r.RemoteAddr)

	// 直连对端不可信：只用 RemoteAddr
	if !isTrustedProxy(peer) {
		if peer != nil {
			return peer.String()
		}
		return r.RemoteAddr
	}

	// 来自可信代理（Nginx 已用 $remote_addr 覆盖写入 X-Real-IP）
	if ip := parseIPHeader(r.Header.Get("X-Real-IP")); ip != nil {
		return ip.String()
	}

	// 退回 X-Forwarded-For：从右往左跳过可信代理，取第一个真实客户端地址
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := parseIPHeader(parts[i])
			if ip == nil {
				continue
			}
			if !isTrustedProxy(ip) {
				return ip.String()
			}
		}
	}

	if peer != nil {
		return peer.String()
	}
	return r.RemoteAddr
}
