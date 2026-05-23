package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	inventorypb "github.com/zhian9/GoForge/server/api/inventory/v1"
	"github.com/zhian9/GoForge/server/internal/service/inventory"
)

var configFile = flag.String("f", "../configs/dev/inventory-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	var c inventory.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 创建服务上下文
	svcCtx := inventory.NewServiceContext(c)

	// 创建库存服务
	inventorySvc := inventory.NewInventoryService(svcCtx)

	// 创建 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		inventorypb.RegisterInventoryServiceServer(grpcServer, inventorySvc)

		// 开发/测试环境开启 gRPC 反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("库存服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
