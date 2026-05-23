package middleware

import (
	"net/http"

	"github.com/zhian9/GoForge/server/internal/handler"
)

// FileUploadMiddleware 文件上传中间件
// 拦截文件上传请求，使用专门的 handler 处理
func FileUploadMiddleware(fileUploadHandler *handler.FileUploadHandler) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查是否是文件上传请求（必须精确匹配路径）
			path := r.URL.Path

			// 先设置 CORS 头（必须在任何响应之前）
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Content-Length")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// 处理 OPTIONS 预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if fileUploadHandler != nil &&
				(path == "/api/v1/files/upload" || path == "/api/v1/files/batch-upload") {

				// 直接处理，不继续传递
				if path == "/api/v1/files/upload" {
					fileUploadHandler.HandleUpload(w, r)
					return
				} else if path == "/api/v1/files/batch-upload" {
					fileUploadHandler.HandleBatchUpload(w, r)
					return
				}
			}

			// 其他请求继续传递给 Gateway
			next(w, r)
		}
	}
}
