// Package interceptor 提供跨服务复用的 gRPC 服务端拦截器。
package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/zhian9/GoForge/server/internal/pkg/utils"
)

// AuthInterceptor 从 gRPC metadata 的 Authorization 中解析 JWT，
// 并把 user_id / username 注入 context，供业务方法取用。
//
// 为什么需要它：服务端不能信任请求参数里的 user_id —— 那是调用方随便填的。
// 正确做法是只认 token 里的身份。拦截器负责「解析身份」，
// 「这个操作要不要登录」由具体业务方法决定（见各服务的 401 判断）。
//
// token 来源有两种：
//   - authorization        —— 直接调用 gRPC
//   - gateway-authorization —— 经 go-zero 网关转发（网关把 HTTP 头转成了这个 metadata key）
//
// 解析失败时不直接拒绝，而是继续执行：这样公开接口（商品列表等）不受影响；
// 受保护接口因为在 context 里拿不到 user_id，会在业务层返回 401。
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			authHeaders = md.Get("gateway-authorization")
		}
		if len(authHeaders) == 0 {
			return handler(ctx, req)
		}

		token := strings.TrimSpace(authHeaders[0])
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
		if token == "" {
			return handler(ctx, req)
		}

		claims, err := utils.ParseToken(token, jwtSecret)
		if err != nil {
			// 无效 token 不注入身份：受保护接口会因为没有 user_id 而返回 401
			return handler(ctx, req)
		}

		ctx = utils.WithUserID(ctx, claims.UserID)
		ctx = utils.WithUsername(ctx, claims.Username)
		return handler(ctx, req)
	}
}
