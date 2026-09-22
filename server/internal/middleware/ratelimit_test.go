package middleware

import (
	"net/http"
	"testing"
	"time"
)

func newTestRequest(remoteAddr string, headers map[string]string) *http.Request {
	r := &http.Request{RemoteAddr: remoteAddr, Header: http.Header{}}
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	return r
}

// 直连对端是公网地址时，请求头里的转发信息都不能信，
// 否则伪造一个 X-Forwarded-For 就能让每个请求都换一个限流 key。
func TestClientIPIgnoresSpoofedHeadersFromUntrustedPeer(t *testing.T) {
	r := newTestRequest("203.0.113.7:5555", map[string]string{
		"X-Forwarded-For": "1.2.3.4",
		"X-Real-IP":       "5.6.7.8",
	})

	if got := ClientIP(r); got != "203.0.113.7" {
		t.Fatalf("ClientIP = %q, want %q", got, "203.0.113.7")
	}
}

// 来自 Nginx（可信代理）时，使用 X-Real-IP：该头由 Nginx 用 $remote_addr 覆盖写入。
func TestClientIPUsesRealIPFromTrustedProxy(t *testing.T) {
	r := newTestRequest("127.0.0.1:5555", map[string]string{
		"X-Real-IP": "198.51.100.9",
	})

	if got := ClientIP(r); got != "198.51.100.9" {
		t.Fatalf("ClientIP = %q, want %q", got, "198.51.100.9")
	}
}

// 没有 X-Real-IP 时退回 X-Forwarded-For：最左边的值可能是客户端自己塞进来的，
// 必须从右往左跳过可信代理，取第一个非可信地址。
func TestClientIPSkipsClientSuppliedForwardedFor(t *testing.T) {
	r := newTestRequest("127.0.0.1:5555", map[string]string{
		"X-Forwarded-For": "1.2.3.4, 198.51.100.9",
	})

	if got := ClientIP(r); got != "198.51.100.9" {
		t.Fatalf("ClientIP = %q, want %q", got, "198.51.100.9")
	}
}

// 只有可信代理时才看转发头；全部是内网地址时回落到 RemoteAddr。
func TestClientIPFallsBackToPeerWhenForwardedForIsAllTrusted(t *testing.T) {
	r := newTestRequest("192.168.1.10:5555", map[string]string{
		"X-Forwarded-For": "10.0.0.5, 172.16.0.9",
	})

	if got := ClientIP(r); got != "192.168.1.10" {
		t.Fatalf("ClientIP = %q, want %q", got, "192.168.1.10")
	}
}

// 空闲的 key 要被回收，避免 map 只增不减；期间有请求的 key 要保留。
func TestRateLimiterEvictsIdleEntries(t *testing.T) {
	rl := NewRateLimiterWithTTL(30 * time.Millisecond)
	defer rl.Close()

	first := rl.GetLimiter("203.0.113.7", 10, 10)
	rl.GetLimiter("198.51.100.9", 10, 10)
	if rl.Len() != 2 {
		t.Fatalf("Len = %d, want 2", rl.Len())
	}

	// 203.0.113.7 一直有请求，不应被回收；198.51.100.9 空闲超过 ttl，应被回收
	time.Sleep(50 * time.Millisecond)
	rl.GetLimiter("203.0.113.7", 10, 10)

	if removed := rl.evictExpired(); removed != 1 {
		t.Fatalf("evictExpired removed = %d, want 1", removed)
	}
	if rl.Len() != 1 {
		t.Fatalf("Len = %d, want 1", rl.Len())
	}

	// 同一个 key 在回收前拿到的是同一个限流器实例（令牌桶状态不会丢）
	if got := rl.GetLimiter("203.0.113.7", 10, 10); got != first {
		t.Fatal("活跃 key 的限流器被重建了，令牌桶状态丢失")
	}

	// 被回收的 key 再次访问会重新创建
	third := rl.GetLimiter("198.51.100.9", 10, 10)
	if third == nil || rl.Len() != 2 {
		t.Fatalf("被回收的 key 未能重建，Len = %d", rl.Len())
	}
}
