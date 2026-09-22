package main

import (
	"context"
	"errors"
	"os"

	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	messagepb "github.com/zhian9/GoForge/server/api/message/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/interceptor"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/message"
	messageService "github.com/zhian9/GoForge/server/internal/service/message/service"
)

var configFile = flag.String("f", "../configs/dev/message-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c message.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := message.NewServiceContext(c)
	messageSvc := message.NewMessageService(svcCtx)

	// 事件驱动：订阅订单/支付事件，自动写入站内信。
	// Kafka 不可用时只记录日志，不影响 gRPC 接口（消息仍然可以由后台手动发送）。
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()
	if len(c.Kafka.Brokers) > 0 && svcCtx.MessageRepo != nil {
		consumerGroup := c.Kafka.ConsumerGroup
		if consumerGroup == "" {
			consumerGroup = "message-service-consumer"
		}
		eventConsumer, err := message.NewEventConsumer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ConsumerGroup: consumerGroup,
			OffsetInitial: c.Kafka.OffsetInitial,
		}, messageService.NewMessageLogic(svcCtx.MessageRepo))
		if err != nil {
			logx.Errorf("初始化事件消费者失败（站内信不会自动生成）: %v", err)
		} else {
			defer eventConsumer.Close()
			go func() {
				if err := eventConsumer.Start(consumerCtx); err != nil && !errors.Is(err, context.Canceled) {
					logx.Errorf("事件消费者退出: %v", err)
				}
			}()
			fmt.Println("已订阅订单/支付事件，将自动生成站内信")
		}
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		messagepb.RegisterMessageServiceServer(grpcServer, messageSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
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

	fmt.Printf("消息服务启动在 %s\n", c.ListenOn)
	s.Start()
}
