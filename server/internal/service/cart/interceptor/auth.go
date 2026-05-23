package interceptor

import (
	"context"
	"strings"

	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor 认证拦截器，从 metadata 中提取 JWT token 并解析 user_id
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			authHeaders := md.Get("authorization")
			if len(authHeaders) == 0 {
				authHeaders = md.Get("gateway-authorization")
			}
			if len(authHeaders) > 0 {
				token := authHeaders[0]
				if strings.HasPrefix(token, "Bearer ") {
					token = strings.TrimPrefix(token, "Bearer ")
				}

				claims, err := utils.ParseToken(token, jwtSecret)
				if err == nil {
					ctx = utils.WithUserID(ctx, claims.UserID)
					ctx = utils.WithUsername(ctx, claims.Username)
				}
			}
		}

		return handler(ctx, req)
	}
}
