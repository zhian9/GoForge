package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/rest"

	"github.com/zhian9/GoForge/server/internal/handler"
	"github.com/zhian9/GoForge/server/internal/middleware"
)

var configFile = flag.String("f", "../configs/dev/gateway.yaml", "配置文件路径")

// swaggerUIPort Swagger UI 单独占用的端口（nginx 的 /swagger/ 直接指向它）
const swaggerUIPort = ":8095"

// 上传接口的体积上限（网关默认 MaxBytes 只有 1MB，multipart 上传必须单独放宽）
const (
	maxUploadBytes      = 100 << 20 // 100MB，与 handler 内部的限制保持一致
	maxBatchUploadBytes = 200 << 20 // 200MB
)

// gatewayRateLimit 返回网关限流的 (qps, burst)。
//
// 可用环境变量覆盖：GATEWAY_RATE_QPS / GATEWAY_RATE_BURST。
// 默认 200 QPS、突发 400 —— 对本地开发和压测足够宽松，
// 同时能挡住明显的异常流量。返回的 qps <= 0 表示关闭限流。
func gatewayRateLimit() (int, int) {
	qps := 200
	if v := os.Getenv("GATEWAY_RATE_QPS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			qps = n
		}
	}
	if qps <= 0 {
		return 0, 0
	}

	burst := qps * 2
	if v := os.Getenv("GATEWAY_RATE_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			burst = n
		}
	}
	if burst < 1 {
		burst = 1
	}
	return qps, burst
}

func main() {
	flag.Parse()

	var c gateway.GatewayConf
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// Swagger 静态文件服务和 UI，单独端口，不影响网关本身
	go startSwaggerServer()

	// 文件上传处理器：multipart/form-data 没法走 gRPC 映射，直接在网关进程内处理。
	// 注意 file-service 必须先启动，否则这里只记录日志、上传接口不注册。
	fileAddr := os.Getenv("FILE_SERVICE_ADDR")
	if fileAddr == "" {
		fileAddr = "127.0.0.1:8012"
	}
	fileUploadHandler, err := handler.NewFileUploadHandler(fileAddr)
	if err != nil {
		log.Printf("⚠️  创建文件上传处理器失败: %v，文件上传功能将不可用", err)
		fileUploadHandler = nil
	}
	if fileUploadHandler != nil {
		defer func() { _ = fileUploadHandler.Close() }()
	}

	corsMiddleware := middleware.NewCorsMiddleware()

	// 网关直接跑在对外端口上。
	//
	// 静态文件、文件上传、CORS、限流都注册在同一进程的路由表上
	// （gateway.Server 内嵌 *rest.Server，可直接 AddRoute），
	// 省掉一次环回 TCP 往返与一次完整的 HTTP 编解码。
	gw := gateway.MustNewServer(c, gateway.WithHeaderProcessor(func(header http.Header) []string {
		// 转发 Authorization 头到 gRPC metadata，否则 gRPC 服务收不到 token
		var headers []string
		if auth := header.Get("Authorization"); auth != "" {
			headers = append(headers, "gateway-authorization:"+auth)
		}
		return headers
	}))
	defer gw.Stop()

	// 添加 CORS 中间件（对所有路由生效，含下面注册的静态文件与上传路由）
	gw.Use(corsMiddleware.Handle)

	// 网关级限流（按客户端 IP 的令牌桶）
	if qps, burst := gatewayRateLimit(); qps > 0 {
		gw.Use(middleware.GatewayRateLimitMiddleware(qps, burst))
		log.Printf("网关限流已启用: qps=%d burst=%d（按客户端 IP）", qps, burst)
	}

	// 静态文件：上传产物 /uploads/{category}/{file} 与 /uploads/{file}，爬虫图片 /images/...
	registerStaticRoutes(gw, "uploads", "/uploads")
	registerStaticRoutes(gw, "images", "/images")

	// 文件上传路由（multipart，单独放宽请求体上限）
	if fileUploadHandler != nil {
		gw.AddRoute(rest.Route{
			Method:  http.MethodPost,
			Path:    "/api/v1/files/upload",
			Handler: fileUploadHandler.HandleUpload,
		}, rest.WithMaxBytes(maxUploadBytes))
		gw.AddRoute(rest.Route{
			Method:  http.MethodOptions,
			Path:    "/api/v1/files/upload",
			Handler: fileUploadHandler.HandleUpload,
		})
		gw.AddRoute(rest.Route{
			Method:  http.MethodPost,
			Path:    "/api/v1/files/batch-upload",
			Handler: fileUploadHandler.HandleBatchUpload,
		}, rest.WithMaxBytes(maxBatchUploadBytes))
		gw.AddRoute(rest.Route{
			Method:  http.MethodOptions,
			Path:    "/api/v1/files/batch-upload",
			Handler: fileUploadHandler.HandleBatchUpload,
		})
		log.Printf("✅ 文件上传路由已注册: /api/v1/files/upload, /api/v1/files/batch-upload")
	}

	log.Printf("✅ 主 HTTP 服务器启动在 %s:%d", c.Host, c.Port)
	log.Printf("✅ 静态文件服务已启用: /uploads/ -> uploads/, /images/ -> images/")
	log.Printf("✅ 文件上传功能已启用，支持最大 %dMB 文件", maxUploadBytes>>20)
	log.Printf("✅ CORS 已启用，允许跨域请求")

	fmt.Printf("API Gateway 启动在 %s:%d\n", c.Host, c.Port)
	fmt.Printf("📡 API 文档: http://localhost%s/\n", swaggerUIPort)
	fmt.Printf("💡 提示: 根路径 / 未配置路由，请使用 /api/v1/... 路径访问 API\n")

	// 启动网关（阻塞）
	gw.Start()
}

// staticRoutePaths 返回某个静态目录需要注册的路由路径。
//
// go-zero 的路由只支持字面量分段和 :param（单个分段），没有 catch-all，
// 所以按实际产物形态把 1 段、2 段两种路径都注册上：
//
//	/uploads/image/<fileID>.png、/uploads/avatar/<fileID>.jpg
//	/images/<file>
func staticRoutePaths(prefix string) []string {
	return []string{prefix + "/:file", prefix + "/:category/:file"}
}

// registerStaticRoutes 把本地目录挂到网关路由上（等价于原来的 http.FileServer + StripPrefix）
func registerStaticRoutes(svr *gateway.Server, dir, prefix string) {
	fileServer := http.StripPrefix(prefix+"/", http.FileServer(http.Dir(dir)))
	handler := func(w http.ResponseWriter, r *http.Request) {
		setStaticCORSHeaders(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		fileServer.ServeHTTP(w, r)
	}

	for _, p := range staticRoutePaths(prefix) {
		svr.AddRoute(rest.Route{Method: http.MethodGet, Path: p, Handler: handler})
		// 预检请求必须先匹配到路由，中间件才有机会执行
		svr.AddRoute(rest.Route{Method: http.MethodOptions, Path: p, Handler: handler})
	}

	log.Printf("✅ 静态文件路由已注册: %s/ -> %s/", prefix, dir)
}

// setStaticCORSHeaders 给静态资源补 CORS 头（图片可能被跨域引用）
func setStaticCORSHeaders(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// startSwaggerServer 启动 Swagger 静态文件服务与 UI（独立端口）
func startSwaggerServer() {
	mux := http.NewServeMux()

	// 静态文件服务：docs/swagger 下是各服务的 *.swagger.json
	fs := http.FileServer(http.Dir("docs/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerUIPage))
	})

	log.Printf("Swagger UI 服务启动在 %s", swaggerUIPort)
	log.Printf("访问地址: http://localhost%s/", swaggerUIPort)
	if err := http.ListenAndServe(swaggerUIPort, mux); err != nil {
		log.Printf("Swagger UI 服务启动失败: %v (可能端口被占用，跳过)", err)
	}
}

// swaggerUIPage Swagger UI 页面。
//
// 之前这段内嵌 HTML 的 JS 是坏的（window.onload 被写成 window. function，
// 且漏了 swagger-ui-bundle.js / swagger-ui-standalone-preset.js 两个 script 标签），
// 页面打开必然白屏；这里改成能正常加载的最小实现。
// 注意：docs/swagger 下的 *.swagger.json 需要由 cmd/generate-swagger 生成后才有内容。
const swaggerUIPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8" />
    <title>GoForge API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js"></script>
    <script>
        function renderSwagger(urls) {
            window.ui = SwaggerUIBundle({
                urls: urls,
                "urls.primaryName": urls.length ? urls[0].name : "",
                dom_id: "#swagger-ui",
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null
            })
        }

        window.onload = function () {
            // 文档清单由 cmd/generate-swagger 生成（docs/swagger/index.json），
            // 这样页面只会列出真实存在的文档，不会再出现点开就是 Not Found 的服务。
            fetch("/swagger/index.json")
                .then(function (res) {
                    if (!res.ok) throw new Error("index.json " + res.status)
                    return res.json()
                })
                .then(function (data) {
                    var specs = (data && data.specs) || []
                    if (!specs.length) throw new Error("清单为空")
                    renderSwagger(specs)
                })
                .catch(function (err) {
                    document.getElementById("swagger-ui").innerHTML =
                        "<p style=\"font-family:system-ui;padding:24px\">加载文档清单失败：" + err +
                        "<br/>请先在 server/ 目录执行 <code>go run ./cmd/generate-swagger</code> 生成 docs/swagger。</p>"
                })
        }
    </script>
</body>
</html>
`
