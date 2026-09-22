package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userpb "github.com/zhian9/GoForge/server/api/user/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"
	"github.com/zhian9/GoForge/server/internal/service/user"
)

var configFile = flag.String("f", "../configs/dev/user-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置（包含 RpcServerConf 和自定义配置）
	var c user.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 创建服务上下文
	svcCtx := user.NewServiceContext(c)

	// 创建用户服务
	userSvc := user.NewUserService(svcCtx)

	// 获取 JWT Secret（从配置中获取，如果没有则使用默认值）
	jwtSecret := c.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-key" // 开发环境默认值
	}

	// 创建 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		userpb.RegisterUserServiceServer(grpcServer, userSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// 添加认证拦截器：从 metadata.authorization 解析 JWT，把 user_id 写进 ctx
	s.AddUnaryInterceptors(interceptor.AuthInterceptor(jwtSecret))
	defer s.Stop()

	fmt.Printf("用户服务启动在 %s\n", c.ListenOn)
	s.Start()
}
