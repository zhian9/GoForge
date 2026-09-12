package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/order"
	orderService "github.com/zhian9/GoForge/server/internal/service/order/service"
)

var configFile = flag.String("f", "../configs/dev/order-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	var c order.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 创建服务上下文
	svcCtx := order.NewServiceContext(c)

	// DB 必须可用，否则消费时会在 gorm 调用处 panic
	if svcCtx.DB == nil {
		fmt.Fprintf(os.Stderr, "错误: order-service-consumer 数据库连接失败，请检查 MySQL 配置（host/user/password/db）\\n")
		fmt.Fprintf(os.Stderr, "提示: 当前使用配置文件 %s\\n", *configFile)
		os.Exit(1)
	}

	// 创建秒杀消费者
	seckillConsumer := orderService.NewSeckillConsumer(
		svcCtx.OrderRepo,
		svcCtx.OrderItemRepo,
		svcCtx.DB,
		svcCtx.Cache,
	)

	// 创建支付结果消费者
	paymentConsumer := orderService.NewPaymentConsumer(svcCtx.OrderRepo, svcCtx.DB, svcCtx.Cache)

	// 创建Kafka消费者（重试机制）
	consumerConfig := &mq.Config{
		Brokers:       c.Kafka.Brokers,
		Version:       c.Kafka.Version,
		ConsumerGroup: "order-service-seckill-consumer",
	}

	var consumer *mq.Consumer
	var err error
	maxRetries := 5
	retryDelay := 3 // 秒

	for i := 0; i < maxRetries; i++ {
		consumer, err = mq.NewConsumer(consumerConfig)
		if err == nil {
			break
		}
		logx.Errorf("创建Kafka消费者失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			logx.Infof("等待 %d 秒后重试...", retryDelay)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}

	if err != nil {
		logx.Errorf("创建Kafka消费者失败，已重试 %d 次: %v", maxRetries, err)
		logx.Errorf("请确保 Kafka 已启动: make start-infra")
		os.Exit(1)
	}
	defer consumer.Close()

	// 注册消息处理器
	consumer.RegisterHandler(mq.TopicSeckillOrder, func(ctx context.Context, message *mq.Message) error {
		return seckillConsumer.Consume(ctx, message)
	})
	consumer.RegisterHandler(mq.TopicPaymentSuccess, func(ctx context.Context, message *mq.Message) error {
		return paymentConsumer.ConsumePaymentSuccess(ctx, message)
	})
	consumer.RegisterHandler(mq.TopicPaymentRefunded, func(ctx context.Context, message *mq.Message) error {
		return paymentConsumer.ConsumePaymentRefund(ctx, message)
	})

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动消费者（在goroutine中）
	go func() {
		if err := consumer.Start(ctx, []string{mq.TopicSeckillOrder, mq.TopicPaymentSuccess, mq.TopicPaymentRefunded}); err != nil {
			logx.Errorf("启动Kafka消费者失败: %v", err)
			cancel()
		}
	}()

	fmt.Println("订单服务Kafka消费者已启动，等待消息...")
	fmt.Println("按 Ctrl+C 退出")

	// 等待信号
	<-sigChan
	fmt.Println("\\n正在关闭消费者...")
}
