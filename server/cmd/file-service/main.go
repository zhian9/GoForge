package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	filepb "github.com/zhian9/GoForge/server/api/file/v1"
	"github.com/zhian9/GoForge/server/internal/service/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "../configs/dev/file-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c file.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := file.NewServiceContext(c)
	fileSvc := file.NewFileService(svcCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		//注册服务
		filepb.RegisterFileServiceServer(grpcServer, fileSvc)

		// 开发/测试环境开启 grpc 反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("文件服务启动在 %s\n", c.ListenOn)
	s.Start()
}
