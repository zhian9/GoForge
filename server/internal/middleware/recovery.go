package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
)

// RecoveryMiddleware 恢复中间件（捕获panic）
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logx.Errorf("Panic recovered: %v", err)
					http.Error(w, "内部服务器错误", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
