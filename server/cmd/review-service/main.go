package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	reviewpb "github.com/zhian9/GoForge/server/api/review/v1"
	"github.com/zhian9/GoForge/server/internal/service/review"
)

var configFile = flag.String("f", "../configs/dev/review-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c review.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := review.NewServiceContext(c)
	reviewSvc := review.NewReviewService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		reviewpb.RegisterReviewServiceServer(grpcServer, reviewSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("评价服务启动在 %s\n", c.ListenOn)
	s.Start()
}
