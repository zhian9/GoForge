package main

import (
	"os"

	"flag"
	"fmt"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	paymentpb "github.com/zhian9/GoForge/server/api/payment/v1"
	"github.com/zhian9/GoForge/server/internal/service/payment"
)

var configFile = flag.String("f", "../configs/dev/payment-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c payment.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := payment.NewServiceContext(c)
	paymentSvc := payment.NewPaymentService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		paymentpb.RegisterPaymentServiceServer(grpcServer, paymentSvc)

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

	fmt.Printf("支付服务启动在 %s\n", c.ListenOn)
	s.Start()
}
