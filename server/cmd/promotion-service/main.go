package main

import (
	"os"

	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	promotionpb "github.com/zhian9/GoForge/server/api/promotion/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"
	"github.com/zhian9/GoForge/server/internal/service/promotion"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "../configs/dev/promotion-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c promotion.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	//创建上下文
	svcCtx := promotion.NewServiceContext(c)
	promotionSvc := promotion.NewPromotionService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		promotionpb.RegisterPromotionServiceServer(grpcServer, promotionSvc)

		//
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "goforge-jwt-secret"
	}
	s.AddUnaryInterceptors(interceptor.AuthInterceptor(jwtSecret))

	defer s.Stop()

	fmt.Printf("营销服务启动在: %s\n", c.ListenOn)
	s.Start()
}
