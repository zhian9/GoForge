package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	orderpb "github.com/zhian9/GoForge/server/api/order/v1"
	"github.com/zhian9/GoForge/server/internal/service/order"
)

var configFile = flag.String("f", "../configs/dev/order-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置（包含 RpcServerConf 和自定义配置）
	var c order.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 创建服务上下文
	svcCtx := order.NewServiceContext(c)

	// 关键依赖检查：DB 初始化失败会导致 s.logic 为 nil，进而触发空指针
	if svcCtx.DB == nil || svcCtx.OrderRepo == nil || svcCtx.OrderItemRepo == nil || svcCtx.OrderLogRepo == nil {
		fmt.Fprintf(os.Stderr, "错误: order-service 数据库连接失败，请检查 MySQL 配置（host/user/password/db）\\n")
		fmt.Fprintf(os.Stderr, "提示: 当前使用配置文件 %s\\n", *configFile)
		os.Exit(1)
	}

	// 创建订单服务
	orderSvc := order.NewOrderService(svcCtx)

	// 创建 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		orderpb.RegisterOrderServiceServer(grpcServer, orderSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("订单服务启动在 %s\n", c.ListenOn)
	s.Start()
}
