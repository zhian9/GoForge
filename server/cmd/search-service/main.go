package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	searchpb "github.com/zhian9/GoForge/server/api/search/v1"
	"github.com/zhian9/GoForge/server/internal/service/search"
)

var configFile = flag.String("f", "../configs/dev/search-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c search.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := search.NewServiceContext(c)
	searchSvc := search.NewSearchService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		searchpb.RegisterSearchServiceServer(grpcServer, searchSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("搜索服务启动在 %s\n", c.ListenOn)
	s.Start()
}
