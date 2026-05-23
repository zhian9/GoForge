package middleware

import (
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"strings"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(jwtSecret string, whitelist []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查白名单
			path := r.URL.Path
			for _, whitePath := range whitelist {
				if strings.HasPrefix(path, whitePath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			//从Header获取Token
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "未授权", http.StatusUnauthorized)
				return
			}

			//移除Bearer前缀
			if strings.HasPrefix(token, "Bearer") {
				token = strings.TrimPrefix(token, "Bearer")
			}

			// 验证Token
			claims, err := utils.ParseToken(token, jwtSecret)
			if err != nil {
				logx.Errorf("Token验证失败： %v", err)
				http.Error(w, "Token无效", http.StatusUnauthorized)
				return
			}

			//将用户信息存储到Context
			ctx := r.Context()
			ctx = utils.WithUserID(ctx, claims.UserID)
			ctx = utils.WithUsername(ctx, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
