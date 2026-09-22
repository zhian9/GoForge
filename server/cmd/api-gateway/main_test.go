package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/search"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/gateway"
)

// TestStaticRoutePathsMatchUploadLayout 校验静态文件路由能匹配真实的产物路径。
//
// go-zero 的路由匹配用的是 core/search.Tree（字面量分段 + :param 单段），
// 这里用同一棵树验证：/uploads/image/<file> 这类真实路径必须命中，否则图片会 404。
func TestStaticRoutePathsMatchUploadLayout(t *testing.T) {
	tree := search.NewTree()
	for _, p := range staticRoutePaths("/uploads") {
		if err := tree.Add(p, p); err != nil {
			t.Fatalf("注册路由 %s 失败: %v", p, err)
		}
	}

	for _, url := range []string{
		"/uploads/image/0b0f44133ea10b011553f3b5d9de3dad.jpg",  // 上传的商品图
		"/uploads/avatar/3920ce5bec54683c2fa24d4551b0b884.png", // 上传的头像
		"/uploads/token.png", // 单段（GetFileURL 的形态）
	} {
		if _, ok := tree.Search(url); !ok {
			t.Errorf("静态路由匹配不到 %s，该图片会 404", url)
		}
	}

	// 已知限制：层级超过两段的路径匹配不到（go-zero 路由没有 catch-all）。
	// 如果这条断言失败，说明 go-zero 支持了通配符，staticRoutePaths 可以简化。
	if _, ok := tree.Search("/uploads/a/b/c.png"); ok {
		t.Log("go-zero 路由已支持更深层级，可以简化 staticRoutePaths 为单条规则")
	}
}

// TestGatewayConfigHasNoUploadMappings 上传接口必须只由网关进程内的 handler 处理。
//
// 如果网关配置里也把 /api/v1/files/upload 映射成 gRPC 调用，启动时 AddRoute
// 会因为路由重复而 panic（multipart 请求体也不是 gRPC 能解析的）。
func TestGatewayConfigHasNoUploadMappings(t *testing.T) {
	configs := []string{
		"../../../configs/dev/gateway.yaml",
		"../../../configs/docker/gateway.yaml",
		"../../../configs/k8s/gateway.yaml",
	}
	blocked := []string{"/api/v1/files/upload", "/api/v1/files/batch-upload"}

	for _, cfg := range configs {
		data, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatalf("读取配置 %s 失败: %v", cfg, err)
		}
		for _, path := range blocked {
			if bytes.Contains(data, []byte("Path: "+path)) {
				t.Errorf("%s 中出现了 %s 的路由映射，会与网关注册的上传路由冲突", cfg, path)
			}
		}
	}
}

// TestStaticRoutesServeThroughGateway 用真实端口跑一遍注册后的路由，
// 确认静态文件由网关自身（同进程、同端口）返回，不再依赖那层反向代理。
func TestStaticRoutesServeThroughGateway(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "uploads")
	if err := os.MkdirAll(filepath.Join(dir, "image"), 0o755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	content := []byte("fake-image-bytes")
	if err := os.WriteFile(filepath.Join(dir, "image", "abc.png"), content, 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	port := freePort(t)
	var c gateway.GatewayConf
	c.Name = "api-gateway-test"
	c.Host = "127.0.0.1"
	c.Port = port
	c.Mode = service.DevMode
	c.Timeout = 3000
	c.MaxBytes = 1 << 20
	c.Middlewares.MaxBytes = true
	c.Middlewares.Log = false
	c.Middlewares.Timeout = false

	// 没有 upstream：本用例只验证网关自身的静态路由，不需要 etcd 与下游服务
	gw := gateway.MustNewServer(c)
	registerStaticRoutes(gw, dir, "/uploads")
	go gw.Start()
	defer gw.Stop()

	url := fmt.Sprintf("http://127.0.0.1:%d/uploads/image/abc.png", port)

	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		res, err := http.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}

		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200（响应体: %s）", url, res.StatusCode, body)
		}
		if string(body) != string(content) {
			t.Fatalf("响应体 = %q, want %q", body, content)
		}
		return
	}

	t.Fatalf("网关未在 10s 内就绪: %v", lastErr)
}

// freePort 取一个空闲端口；当前环境不允许监听时跳过用例（例如受限沙箱）
func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("当前环境无法监听端口，跳过集成用例: %v", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}
