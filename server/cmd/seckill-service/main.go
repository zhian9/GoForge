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

	seckillpb "github.com/zhian9/GoForge/server/api/seckill/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"
	"github.com/zhian9/GoForge/server/internal/service/seckill"
)

var configFile = flag.String("f", "../configs/dev/seckill-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	var c seckill.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 创建服务上下文
	svcCtx := seckill.NewServiceContext(c)

	// 检查关键依赖
	if svcCtx.Redis == nil {
		fmt.Fprintf(os.Stderr, "错误: Redis 连接失败，请确保 Redis 已启动\\n")
		fmt.Fprintf(os.Stderr, "启动命令: make start-infra\\n")
		os.Exit(1)
	}

	if svcCtx.MQProducer == nil {
		fmt.Fprintf(os.Stderr, "警告: Kafka 生产者初始化失败，秒杀功能可能无法正常工作\\n")
		fmt.Fprintf(os.Stderr, "请确保 Kafka 已启动: make start-infra\\n")
		// 不退出，允许服务启动，但功能会受限
	}

	// 创建秒杀服务
	seckillSvc := seckill.NewSeckillService(svcCtx)

	// 创建 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		seckillpb.RegisterSeckillServiceServer(grpcServer, seckillSvc)

		// 开发/测试环境开启 gRPC 反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 添加认证拦截器：抢购必须用 token 里的用户身份，
	// 否则任何人不带 token、随便填一个 user_id 就能替别人抢购。
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "goforge-jwt-secret"
	}
	s.AddUnaryInterceptors(interceptor.AuthInterceptor(jwtSecret))

	defer s.Stop()

	fmt.Printf("秒杀服务启动在 %s\\n", c.ListenOn)
	s.Start()
}
