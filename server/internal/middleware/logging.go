package middleware

import (
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"time"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 包装ResponseWriter以捕获状态码
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logx.Infow("HTTP请求",
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("status", rw.statusCode),
				logx.Field("duration", duration.Milliseconds()),
				logx.Field("ip", r.RemoteAddr),
				logx.Field("user_agent", r.UserAgent()),
			)
		})
	}
}

// responseWriter 包装ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
