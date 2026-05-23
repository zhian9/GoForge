package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	logisticspb "github.com/zhian9/GoForge/server/api/logistics/v1"
	"github.com/zhian9/GoForge/server/internal/service/logistics"
)

var configFile = flag.String("f", "../configs/dev/logistics-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c logistics.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := logistics.NewServiceContext(c)
	logisticsSvc := logistics.NewLogisticsService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		logisticspb.RegisterLogisticsServiceServer(grpcServer, logisticsSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("物流服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
