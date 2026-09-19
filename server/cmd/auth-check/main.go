// auth-check 是 GoForge 的鉴权回归工具。
//
// 针对一组「需要登录」的接口做两类断言：
//  1. **不带 token，但故意传入他人的 user_id** —— 必须被拒绝（401）。
//     这一条专门用来抓「服务端信任请求参数里的 user_id」这类越权漏洞：
//     修复前 /api/v1/orders?user_id=<别人> 能直接返回他人订单。
//  2. **带合法 token** —— 必须能通过鉴权（不能是 401）。
//     这一条防止「修越权修过头」，把正常用户也挡在门外。
//
// 用法（在 server/ 目录下执行）：
//
//	go run ./cmd/auth-check
//	go run ./cmd/auth-check -victim=37976967029 -base=http://localhost:8080
//
// 前置条件：docker compose 起的网关与服务在运行。
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zhian9/GoForge/server/internal/pkg/utils"
)

var (
	baseURL  = flag.String("base", "http://localhost:8080", "API 网关地址")
	victimID = flag.Int64("victim", 37976967029, "被冒充的用户 ID（用一个真实存在的用户更有说服力）")
	secret   = flag.String("secret", "", "JWT 密钥（留空则读环境变量 JWT_SECRET，再退回内置默认值）")
)

// endpoint 一个需要登录的接口
type endpoint struct {
	name   string
	method string
	// path 用 %d 占位要冒充的 user_id（query 形式）
	path string
	// body 非空时作为 POST body，同样用 %d 占位
	body string
}

var endpoints = []endpoint{
	{"购物车列表", http.MethodGet, "/api/v1/cart?user_id=%d", ""},
	{"用户信息", http.MethodGet, "/api/v1/user/info?user_id=%d", ""},
	{"订单列表", http.MethodGet, "/api/v1/orders?user_id=%d&status=-1", ""},
	{"订单统计", http.MethodGet, "/api/v1/orders/stats?user_id=%d", ""},
	{"秒杀抢购", http.MethodPost, "/api/v1/seckill", `{"user_id":%d,"sku_id":1,"quantity":1}`},
}

func main() {
	flag.Parse()

	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		jwtSecret = "goforge-jwt-secret"
	}

	client := &http.Client{Timeout: 15 * time.Second}

	fmt.Println(strings.Repeat("=", 66))
	fmt.Println("GoForge 鉴权回归（越权 + 正常放行）")
	fmt.Println(strings.Repeat("=", 66))
	fmt.Printf("[环境] 网关 %s，被冒充用户 id=%d，密钥长度 %d\n\n", *baseURL, *victimID, len(jwtSecret))

	token, err := utils.GenerateToken(uint64(*victimID), "auth-check", jwtSecret, 3600)
	if err != nil {
		fmt.Printf("[ERROR] 生成测试 token 失败: %v\n", err)
		os.Exit(1)
	}

	pass, fail := 0, 0
	for _, ep := range endpoints {
		path := ep.path
		if strings.Contains(path, "%d") {
			path = fmt.Sprintf(path, *victimID)
		}
		body := ""
		if ep.body != "" {
			body = fmt.Sprintf(ep.body, *victimID)
		}

		// 断言 1：不带 token，冒充他人 —— 必须 401
		codeNoToken, text1 := call(client, ep.method, path, body, "")
		ok1 := codeNoToken == http.StatusUnauthorized
		detail1 := fmt.Sprintf("返回 HTTP %d", codeNoToken)
		if codeNoToken < 0 {
			detail1 = "请求失败: " + text1
		} else if !ok1 {
			detail1 += "；响应：" + text1
		}
		report(ok1, fmt.Sprintf("%s｜无 token 冒充他人必须 401", ep.name), detail1)

		// 断言 2：带合法 token —— 必须不是 401
		codeWithToken, text2 := call(client, ep.method, path, body, token)
		ok2 := codeWithToken > 0 && codeWithToken != http.StatusUnauthorized
		detail2 := fmt.Sprintf("返回 HTTP %d", codeWithToken)
		if codeWithToken < 0 {
			detail2 = "请求失败: " + text2
		}
		report(ok2, fmt.Sprintf("%s｜带合法 token 必须放行", ep.name), detail2)

		if ok1 {
			pass++
		} else {
			fail++
		}
		if ok2 {
			pass++
		} else {
			fail++
		}
	}

	fmt.Println(strings.Repeat("=", 66))
	fmt.Printf("汇总：%d 项通过，%d 项失败\n", pass, fail)
	if fail > 0 {
		fmt.Println("有接口仍可被未授权访问，或正常请求被误拦 —— 需要检查该服务是否注册了 gRPC 鉴权拦截器。")
	} else {
		fmt.Println("全部通过：这些接口既不接受未授权的他人身份，也不误拦合法用户。")
	}
	fmt.Println(strings.Repeat("=", 66))

	if fail > 0 {
		os.Exit(1)
	}
}

func report(ok bool, name, detail string) {
	mark := "[PASS]"
	if !ok {
		mark = "[FAIL]"
	}
	fmt.Printf("  %s %s（%s）\n", mark, name, detail)
}

// call 发起一次请求，返回 HTTP 状态码与响应体（截断）
func call(client *http.Client, method, path, body, token string) (int, string) {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, *baseURL+path, reader)
	if err != nil {
		return -1, err.Error()
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return -1, err.Error()
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	text := strings.TrimSpace(string(data))
	if len(text) > 160 {
		text = text[:160] + "..."
	}
	return resp.StatusCode, text
}
