package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	jobpb "github.com/zhian9/GoForge/server/api/job/v1"
	"github.com/zhian9/GoForge/server/internal/service/job"
)

var configFile = flag.String("f", "../configs/dev/job-config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c job.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := job.NewServiceContext(c)
	jobSvc := job.NewJobService(svcCtx)

	// 启动「超时未支付订单自动取消」的定时调度。
	//
	// 修复前这个能力只有一个 gRPC 接口、全项目没有任何调度器调用它，
	// 导致超时订单永远不会被取消：库存被永久占用、秒杀闸门配额也不会释放。
	go runCancelExpiredOrdersLoop(svcCtx, jobSvc)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		jobpb.RegisterJobServiceServer(grpcServer, jobSvc)

		// 开发/测试环境开启 gRPC 反射（用于调试工具如 grpcurl 和 Gateway）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("定时任务服务启动在 %s\n", c.ListenOn)
	s.Start()
}

// runCancelExpiredOrdersLoop 周期性取消超时未支付的订单。
//
// 复用 gRPC 服务的同一个入口（JobService.CancelExpiredOrders），
// 保证定时执行和手工调用走完全相同的业务逻辑（含库存回补与秒杀配额释放）。
func runCancelExpiredOrdersLoop(svcCtx *job.ServiceContext, jobSvc *job.JobService) {
	cfg := svcCtx.Config.CancelExpiredOrder
	if !cfg.Enable {
		logx.Info("超时订单自动取消未启用（CancelExpiredOrder.Enable=false）")
		return
	}

	interval := time.Duration(cfg.IntervalSecond) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}
	timeoutMinutes := cfg.TimeoutMinutes
	if timeoutMinutes <= 0 {
		timeoutMinutes = 30
	}
	batchLimit := cfg.BatchLimit
	if batchLimit <= 0 {
		batchLimit = 200
	}

	logx.Infof("超时订单自动取消已启用: 扫描间隔=%s, 超时阈值=%d 分钟, 单批上限=%d",
		interval, timeoutMinutes, batchLimit)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// 单次执行给 60 秒预算，避免任务卡住后下一轮叠加
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		resp, err := jobSvc.CancelExpiredOrders(ctx, &jobpb.CancelExpiredOrdersRequest{
			TimeoutMinutes: int32(timeoutMinutes),
		})
		cancel()

		if err != nil {
			logx.Errorf("超时订单自动取消失败: %v", err)
			continue
		}
		if resp.GetCancelledCount() > 0 {
			logx.Infof("超时订单自动取消完成: 本轮取消 %d 笔（含库存回补与秒杀配额释放）",
				resp.GetCancelledCount())
		}
	}
}
