package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/job/repository"
)

// JobLogic 定时任务业务逻辑
type JobLogic struct {
	orderRepo  repository.OrderRepository
	couponRepo repository.CouponRepository
	redis      *redis.Client // 用于释放秒杀闸门配额（可能为 nil）
	batchLimit int           // 单次最多处理多少笔超时订单
	// 可后续扩展统计专用仓库
	// statsRepo  repository.StatisticsRepository
}

// NewJobLogic 创建定时任务业务逻辑
func NewJobLogic(
	orderRepo repository.OrderRepository,
	couponRepo repository.CouponRepository,
	redisClient *redis.Client,
	batchLimit int,
) *JobLogic {
	if batchLimit <= 0 {
		batchLimit = 200
	}
	return &JobLogic{
		orderRepo:  orderRepo,
		couponRepo: couponRepo,
		redis:      redisClient,
		batchLimit: batchLimit,
	}
}

// CancelExpiredOrdersRequest 订单超时取消请求
type CancelExpiredOrdersRequest struct {
	TimeoutMinutes int
}

// CancelExpiredOrdersResponse 订单超时取消响应
type CancelExpiredOrdersResponse struct {
	CancelledCount int64
}

// CancelExpiredOrders 订单超时取消
func (l *JobLogic) CancelExpiredOrders(ctx context.Context, req *CancelExpiredOrdersRequest) (*CancelExpiredOrdersResponse, error) {
	cancelled, err := l.orderRepo.CancelExpiredOrders(ctx, req.TimeoutMinutes, l.batchLimit)
	if err != nil {
		return nil, apperrors.NewInternalError("取消超时订单失败")
	}

	// 秒杀订单占用的库存有两个地方：MySQL 的 sku.stock（真源，上面已回补）
	// 和 Redis 的 seckill:stock:{skuId}（闸门）。只回补真源不释放闸门，
	// 秒杀配额就永久损失了，用户也会被防重 key 锁住不能重抢。
	for _, o := range cancelled {
		if o.OrderType != repository.OrderTypeSeckill {
			continue
		}
		for _, item := range o.Items {
			l.releaseSeckillQuota(ctx, item.SkuID, o.UserID, item.Quantity)
		}
	}

	return &CancelExpiredOrdersResponse{
		CancelledCount: int64(len(cancelled)),
	}, nil
}

// releaseSeckillQuota 释放一次秒杀预扣：闸门库存回补 + 删除用户防重标记。
//
// 复用秒杀服务里同一个 Lua 脚本（LuaScriptSeckillRollback），保证「退库存」和
// 「删防重标记」原子完成——分两步做一旦中途失败，会留下"库存退了但用户还被锁着"
// 这类更难查的状态。
//
// Redis 不可用时只记日志：真源库存已经回补，这里失败不会造成资损，
// 属于可降级的补偿动作。
func (l *JobLogic) releaseSeckillQuota(ctx context.Context, skuID, userID uint64, quantity int) {
	if l.redis == nil || quantity <= 0 {
		return
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", skuID)
	userKey := fmt.Sprintf("seckill:user:%d:%d", skuID, userID)

	if _, err := cache.ExecuteLuaScript(ctx, l.redis, cache.LuaScriptSeckillRollback,
		[]string{stockKey, userKey}, quantity); err != nil {
		logx.Errorf("释放秒杀配额失败（真源库存已回补，仅闸门配额未释放）: sku_id=%d, user_id=%d, quantity=%d, err=%v",
			skuID, userID, quantity, err)
		return
	}

	logx.Infof("超时取消已释放秒杀配额: sku_id=%d, user_id=%d, quantity=%d", skuID, userID, quantity)
}

// ProcessExpiredCouponsRequest 优惠券过期处理请求
type ProcessExpiredCouponsRequest struct{}

// ProcessExpiredCouponsResponse 优惠券过期处理响应
type ProcessExpiredCouponsResponse struct {
	ExpiredCount int64
}

// ProcessExpiredCoupons 优惠券过期处理
func (l *JobLogic) ProcessExpiredCoupons(ctx context.Context, req *ProcessExpiredCouponsRequest) (*ProcessExpiredCouponsResponse, error) {
	count, err := l.couponRepo.ProcessExpiredCoupons(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("处理过期优惠券失败")
	}

	return &ProcessExpiredCouponsResponse{
		ExpiredCount: count,
	}, nil
}

// GenerateStatisticsRequest 数据统计请求
type GenerateStatisticsRequest struct {
	Date string `json:"date"` // YYYY-MM-DD 格式
}

// GenerateStatisticsResponse 数据统计响应（可选返回统计结果）
type GenerateStatisticsResponse struct {
	Date         string `json:"date"`
	OrderCount   int64  `json:"order_count"`
	OrderAmount  int64  `json:"order_amount"`
	ProductCount int64  `json:"product_count"`
	SalesAmount  int64  `json:"sales_amount"`
	UserCount    int64  `json:"user_count"`
	ActiveUsers  int64  `json:"active_users"`
	CouponUsed   int64  `json:"coupon_used"`
	CouponAmount int64  `json:"coupon_amount"`
}

// GenerateStatistics 数据统计（核心实现）
func (l *JobLogic) GenerateStatistics(ctx context.Context, req *GenerateStatisticsRequest) error {
	if req == nil || req.Date == "" {
		return apperrors.NewInvalidParamError("统计日期不能为空")
	}

	// 验证日期格式 YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return apperrors.NewInvalidParamError("日期格式错误，应为 YYYY-MM-DD")
	}

	//// 1. 订单统计
	//orderStats, err := l.orderRepo.GenerateOrderStatistics(ctx, req.Date)
	//if err != nil {
	//	return apperrors.NewInternalError(fmt.Sprintf("订单统计失败: %v", err))
	//}
	//
	//// 2. 优惠券统计
	//couponStats, err := l.couponRepo.GenerateCouponStatistics(ctx, req.Date)
	//if err != nil {
	//	return apperrors.NewInternalError(fmt.Sprintf("优惠券统计失败: %v", err))
	//}

	// 3. 商品统计 & 用户统计（后续扩展）
	// productStats, err := l.productRepo.GenerateProductStatistics(ctx, req.Date)
	// userStats, err := l.userRepo.GenerateUserStatistics(ctx, req.Date)

	// 4. 保存统计结果到数据库或 Redis（推荐）
	// err = l.statsRepo.SaveDailyStatistics(ctx, req.Date, combineStats(orderStats, couponStats, ...))
	// if err != nil { ... }

	// 可打印日志或记录执行成功
	//fmt.Printf("[Job] GenerateStatistics completed for date: %s | Orders: %d | Sales: %d | Coupons: %d\n",
	//	req.Date, orderStats.OrderCount, orderStats.SalesAmount, couponStats.UsedCount)

	return nil
}
