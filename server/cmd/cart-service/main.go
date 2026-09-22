package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	cartpb "github.com/zhian9/GoForge/server/api/cart/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"
	"github.com/zhian9/GoForge/server/internal/service/cart"
)

var configFile = flag.String("f", "../configs/dev/cart-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c cart.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := cart.NewServiceContext(c)
	cartSvc := cart.NewCartService(svcCtx)

	// 获取 JWT Secret（从配置中获取，如果没有则使用默认值）
	jwtSecret := c.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-key" // 开发环境默认值
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		cartpb.RegisterCartServiceServer(grpcServer, cartSvc)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 添加认证拦截器
	s.AddUnaryInterceptors(interceptor.AuthInterceptor(jwtSecret))
	defer s.Stop()

	fmt.Printf("购物车服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
