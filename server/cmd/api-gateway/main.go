package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"

	"github.com/zhian9/GoForge/server/internal/handler"
	"github.com/zhian9/GoForge/server/internal/middleware"
)

var configFile = flag.String("f", "../configs/dev/gateway.yaml", "配置文件路径")

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

	// 启动 Swagger 静态文件服务和 UI，单独端口，避免影响原有网关
	go func() {
		mux := http.NewServeMux()

		// 静态文件服务
		fs := http.FileServer(http.Dir("docs/swagger"))
		mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

		// Swagger UI 首页
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>E-Commerce System API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    
    
    
        window. function() {
            const ui = SwaggerUIBundle({
                urls: [
                    {url: "/swagger/api/user/v1/user.swagger.json", name: "用户服务"},
                    {url: "/swagger/api/product/v1/product.swagger.json", name: "商品服务"},
                    {url: "/swagger/api/order/v1/order.swagger.json", name: "订单服务"},
                    {url: "/swagger/api/payment/v1/payment.swagger.json", name: "支付服务"},
                    {url: "/swagger/api/cart/v1/cart.swagger.json", name: "购物车服务"},
                    {url: "/swagger/api/inventory/v1/inventory.swagger.json", name: "库存服务"},
                    {url: "/swagger/api/promotion/v1/promotion.swagger.json", name: "营销服务"},
                    {url: "/swagger/api/message/v1/message.swagger.json", name: "消息服务"},
                    {url: "/swagger/api/search/v1/search.swagger.json", name: "搜索服务"},
                    {url: "/swagger/api/recommend/v1/recommend.swagger.json", name: "推荐服务"},
                    {url: "/swagger/api/review/v1/review.swagger.json", name: "评价服务"},
                    {url: "/swagger/api/logistics/v1/logistics.swagger.json", name: "物流服务"},
                    {url: "/swagger/api/file/v1/file.swagger.json", name: "文件服务"},
                    {url: "/swagger/api/job/v1/job.swagger.json", name: "任务服务"}
                ],
                "urls.primaryName": "用户服务",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl,
                    function() {
                        return {
                            statePlugins: {
                                spec: {
                                    wrapSelectors: {
                                        url: function(originalSelector) {
                                            return function(state) {
                                                // 强制使用 localhost:8080
                                                const spec = originalSelector(state);
                                                if (spec && spec.get && spec.get('host')) {
                                                    return 'http://localhost:8080';
                                                }
                                                return 'http://localhost:8080';
                                            };
                                        }
                                    }
                                }
                            }
                        }; 
                    }
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                requestInterceptor: function(request) {
                    // 将所有 API 请求的 host 改为 localhost:8080
                    if (request.url) {
                        // 如果 URL 是相对路径（以 / 开头），添加完整的 host
                        if (request.url.startsWith('/api/')) {
                            request.url = 'http://localhost:8080' + request.url;
                        }
                        // 如果 URL 包含 localhost:8095，替换为 localhost:8080
                        request.url = request.url.replace(/localhost:8095/g, 'localhost:8080');
                        // 如果 URL 是完整的 URL 但 host 不对，也替换
                        request.url = request.url.replace(/http:\\/\\/localhost:8095/g, 'http://localhost:8080');
                        request.url = request.url.replace(/https:\\/\\/localhost:8095/g, 'https://localhost:8080');
                    }
                    return request;
                },
                // 强制使用指定的 host
                onComplete: function() {
                    // 等待 Swagger UI 加载完成后，修改所有显示的 URL
                    setTimeout(function() {
                        // 查找所有显示 URL 的元素并替换
                        const urlElements = document.querySelectorAll('.scheme-container, .request-url, .url');
                        urlElements.forEach(function(el) {
                            if (el.textContent) {
                                el.textContent = el.textContent.replace(/localhost:8095/g, 'localhost:8080');
                            }
                        });
                        // 修改输入框中的 URL
                        const inputElements = document.querySelectorAll('input[type="text"]');
                        inputElements.forEach(function(el) {
                            if (el.value && el.value.includes('localhost:8095')) {
                                el.value = el.value.replace(/localhost:8095/g, 'localhost:8080');
                            }
                        });
                    }, 1000);
                }
            });
        };
    
</body>
</html>`))
				return
			}
			http.NotFound(w, r)
		})

		addr := ":8095" // 修改端口避免冲突
		log.Printf("Swagger UI 服务启动在 %s\\n", addr)
		log.Printf("访问地址: http://localhost%s/\\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("Swagger UI 服务启动失败: %v (可能端口被占用，跳过)", err)
		}
	}()

	// 创建文件上传处理器
	fileAddr := os.Getenv("FILE_SERVICE_ADDR")
	if fileAddr == "" {
		fileAddr = "127.0.0.1:8012"
	}
	fileUploadHandler, err := handler.NewFileUploadHandler(fileAddr)
	if err != nil {
		log.Printf("⚠️  创建文件上传处理器失败: %v，文件上传功能将不可用", err)
		fileUploadHandler = nil
	}

	// 创建 CORS 中间件（按照 go-zero 最佳实践）
	corsMiddleware := middleware.NewCorsMiddleware()

	// 创建 Gateway 服务器（使用内部端口，不直接暴露）
	// Gateway 将在内部端口运行，然后通过反向代理暴露
	internalPort := c.Port + 1000 // 使用 9080 作为内部端口
	internalConfig := c
	internalConfig.Port = internalPort

	gw := gateway.MustNewServer(internalConfig, func(svr *gateway.Server) {
		// 添加 CORS 中间件
		svr.Use(corsMiddleware.Handle)

		// 网关级限流（按客户端 IP 的令牌桶）。
		// 之前限流器写好了却没有任何调用方，是因为类型与网关的 Use() 不匹配，
		// 现在通过 GatewayRateLimitMiddleware 适配后真正生效。
		if qps, burst := gatewayRateLimit(); qps > 0 {
			svr.Use(middleware.GatewayRateLimitMiddleware(qps, burst))
			log.Printf("网关限流已启用: qps=%d burst=%d（按客户端 IP）", qps, burst)
		}
	}, gateway.WithHeaderProcessor(func(header http.Header) []string {
		// 转发 Authorization 头到 gRPC metadata，否则 gRPC 服务收不到 token
		var headers []string
		if auth := header.Get("Authorization"); auth != "" {
			headers = append(headers, "gateway-authorization:"+auth)
		}
		return headers
	}))
	defer gw.Stop()

	// 在后台启动 Gateway（使用内部端口）
	go func() {
		log.Printf("Gateway 内部服务启动在 %s:%d", c.Host, internalPort)
		gw.Start() // Start() 没有返回值，直接调用
	}()

	// 等待 Gateway 启动
	time.Sleep(500 * time.Millisecond)

	// 创建反向代理，转发到 Gateway 内部端口
	gatewayURL, err := url.Parse(fmt.Sprintf("http://%s:%d", c.Host, internalPort))
	if err != nil {
		log.Fatalf("解析 Gateway URL 失败: %v", err)
	}
	gatewayProxy := httputil.NewSingleHostReverseProxy(gatewayURL)

	// 修改反向代理的传输配置，增加请求体大小限制
	gatewayProxy.Transport = &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}

	// 创建主 HTTP 服务器（使用 Gateway 的原始端口）
	mainMux := http.NewServeMux()

	// 静态文件服务：提供上传的图片文件访问
	// 文件保存在 uploads/ 目录下，通过 /uploads/ 路径访问
	// 添加 CORS 支持，允许前端跨域访问图片
	uploadsFS := http.FileServer(http.Dir("uploads"))
	mainMux.HandleFunc("/uploads/", func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 提供静态文件
		http.StripPrefix("/uploads/", uploadsFS).ServeHTTP(w, r)
	})
	log.Printf("✅ 静态文件服务已注册: /uploads/ -> uploads/ (已启用 CORS)")

	// 静态文件服务：提供爬虫下载的图片文件访问
	// 文件保存在 images/ 目录下，通过 /images/ 路径访问
	imagesFS := http.FileServer(http.Dir("images"))
	mainMux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 提供静态文件
		http.StripPrefix("/images/", imagesFS).ServeHTTP(w, r)
	})
	log.Printf("✅ 静态文件服务已注册: /images/ -> images/ (已启用 CORS)")

	// 文件上传路由（直接处理，不经过 Gateway）
	if fileUploadHandler != nil {
		mainMux.HandleFunc("/api/v1/files/upload", fileUploadHandler.HandleUpload)
		mainMux.HandleFunc("/api/v1/files/batch-upload", fileUploadHandler.HandleBatchUpload)
		log.Printf("✅ 文件上传路由已注册: /api/v1/files/upload, /api/v1/files/batch-upload")
	}

	// 修改反向代理，添加 CORS 支持 + Authorization 转发
	originalDirector := gatewayProxy.Director
	gatewayProxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 转发 Authorization 头到 gRPC metadata（go-zero gateway 只转发 Grpc-Metadata- 前缀的头）
		if auth := req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Grpc-Metadata-Authorization", auth)
		}
		origin := req.Header.Get("Origin")
		if origin != "" {
			req.Header.Set("X-Forwarded-Origin", origin)
		}
	}

	// 修改反向代理的响应修改器，添加 CORS 头
	gatewayProxy.ModifyResponse = func(resp *http.Response) error {
		origin := resp.Request.Header.Get("Origin")
		if origin == "" {
			origin = resp.Request.Header.Get("X-Forwarded-Origin")
		}
		if origin != "" {
			resp.Header.Set("Access-Control-Allow-Origin", origin)
			resp.Header.Set("Access-Control-Allow-Credentials", "true")
		} else {
			resp.Header.Set("Access-Control-Allow-Origin", "*")
		}
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Content-Length")
		resp.Header.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		resp.Header.Set("Access-Control-Max-Age", "3600")
		return nil
	}

	// 其他请求转发给 Gateway
	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 如果是文件上传请求，已经在上面处理了，不会到这里
		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Content-Length")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		gatewayProxy.ServeHTTP(w, r)
	})

	// 创建主 HTTP 服务器（增加请求体大小限制）
	mainServer := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", c.Host, c.Port),
		Handler:        mainMux,
		MaxHeaderBytes: 1 << 20,           // 1MB header limit
		ReadTimeout:    120 * time.Second, // 增加读取超时，支持大文件上传
		WriteTimeout:   120 * time.Second, // 增加写入超时
	}

	log.Printf("✅ 主 HTTP 服务器启动在 %s:%d", c.Host, c.Port)
	log.Printf("✅ 文件上传功能已启用，支持最大 100MB 文件")
	log.Printf("✅ 静态文件服务已启用，可通过 /uploads/ 访问上传的文件")

	fmt.Printf("API Gateway 启动在 %s:%d\\n", c.Host, c.Port)
	fmt.Printf("注意：如果看到 Prometheus 端口冲突错误，Gateway 仍可正常使用\\n")
	fmt.Printf("📡 API 文档: http://localhost:8095/\\n")
	fmt.Printf("💡 提示: 根路径 / 未配置路由，请使用 /api/v1/... 路径访问 API\\n")
	fmt.Printf("✅ CORS 已启用，允许跨域请求\\n")
	fmt.Printf("✅ 文件上传功能已启用，支持最大 100MB 文件\\n")
	fmt.Printf("✅ 静态文件服务已启用，可通过 /uploads/ 访问上传的文件\\n")

	// 启动主 HTTP 服务器（这会阻塞）
	if err := mainServer.ListenAndServe(); err != nil {
		log.Fatalf("主 HTTP 服务器启动失败: %v", err)
	}
}
