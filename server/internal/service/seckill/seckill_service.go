package seckill

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	v1 "github.com/zhian9/GoForge/server/api/seckill/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/constants"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/seckill/model"
	"github.com/zhian9/GoForge/server/internal/service/seckill/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"strconv"
	"time"
)

// Seckill 秒杀下单
func (s *SeckillService) Seckill(ctx context.Context, req *v1.SeckillRequest) (*v1.SeckillResponse, error) {
	// 关键安全修复：抢购人只能是当前登录用户（由 AuthInterceptor 从 token 注入），
	// 不能相信请求体里的 user_id —— 否则任何人不带 token、随便填一个 user_id
	// 就能替别人抢购，还会占掉对方的「一人一单」名额。
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}

	//1.参数验证
	if req.SkuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "SKU ID 不能为空")
	}
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	//校验
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化(数据库未连接)")
	}
	now := time.Now().Unix()
	activity, err := s.svcCtx.SeckillActivityRepo.GetActiveBySkuID(ctx, uint64(req.SkuId), now)
	if err != nil {
		//没有活动 /未开始 /已结束
		return &v1.SeckillResponse{
			Code:    1,
			Message: "不在活动时间内",
			Data: &v1.SeckillData{
				Success: false,
				Message: "该商品当前不在秒杀活动时间内",
			},
		}, nil
	}

	// Redis 里的库存闸门只由创建/编辑活动时预热，没有兜底。若 Redis 重启或 key 被清掉，
	// 下面的 Lua 脚本会读不到 key 而恒返回「已抢光」——有库存却卖不出去。
	// 这里在下单前兜底重建一次（重建口径见函数注释）。
	s.ensureSeckillStockGate(ctx, activity)

	//Redis Lua 脚本执行：防超卖 + 防重复
	stockKey := fmt.Sprintf("seckill:stock:%d", req.SkuId)
	userKey := fmt.Sprintf("seckill:user:%d:%d", req.SkuId, userID)

	result, err := cache.ExecuteLuaScript(ctx, s.svcCtx.Redis, cache.LuaScriptSeckill, []string{stockKey, userKey}, quantity)
	if err != nil {
		logx.Errorf("执行秒杀Lua脚本失败: %v", err)
		return nil, status.Error(codes.Internal, "秒杀失败，请稍后重试")
	}

	//解析结果
	code, ok := result.(int64)
	if !ok {
		logx.Errorf("Lua脚本返回结果类型错误: %v", result)
		return nil, status.Error(codes.Internal, "秒杀失败，请稍后重试")
	}

	//处理结果
	switch {
	case code == -1:
		//库存不足
		return &v1.SeckillResponse{
			Code:    1,
			Message: "已抢光",
			Data: &v1.SeckillData{
				Success: false,
				Message: "商品已抢光，请关注下次活动",
			},
		}, nil
	case code == -2:
		//重复抢购
		return &v1.SeckillResponse{
			Code:    1,
			Message: "不可重复抢购",
			Data: &v1.SeckillData{
				Success: false,
				Message: "您已参加过本次秒杀活动",
			},
		}, nil
	case code >= 0:
		//成功 :发送kafka消息
		// 注意 quantity 必须用兜底后的值：Lua 脚本里 quantity<=0 时会按 1 扣减，
		// 如果这里回传原始的 req.Quantity(可能是 0)，就会出现「Redis 扣 1 件、订单记 0 件」的账目不一致。
		seckillMsg := map[string]interface{}{
			"user_id":   userID,
			"sku_id":    req.SkuId,
			"quantity":  quantity,
			"timestamp": time.Now().Unix(),
		}

		message := mq.NewMessage("seckill.order", seckillMsg)

		//使用 sku_id 作为分区key 保证同一SKU的消息有序
		partitionKey := strconv.FormatInt(req.SkuId, 10)
		if err := s.svcCtx.MQProducer.PublishWithKey(ctx, mq.TopicSeckillOrder, partitionKey, message); err != nil {
			logx.Errorf("发送秒杀消息到kafka失败: %v", err)
			// 关键补偿：走到这里说明 Redis 已经扣了库存、也写了防重标记。
			// 如果直接返回失败，用户会被防重 key 锁死（既没抢到、24 小时内也不能再抢），
			// 库存也会凭空少掉且无人认领。因此必须把预扣的这两步一起回滚。
			s.rollbackSeckill(ctx, req.SkuId, int64(userID), quantity)
			return nil, status.Error(codes.Internal, "秒杀失败, 请稍后重试")
		}

		logx.Infof("秒杀成功： user_id=%d, sku_id=%d, quantity=%d", userID, req.SkuId, quantity)

		return &v1.SeckillResponse{
			Code:    0,
			Message: "抢购成功",
			Data: &v1.SeckillData{
				Success: true,
				Message: "抢购成功，订单正在处理中...",
			},
		}, nil
	default:
		logx.Errorf("未知的Lua脚本返回码: %v", code)
		return nil, status.Error(codes.Internal, "秒杀失败，请稍后重试")
	}
}

// ListSeckillActivities 获取秒杀活动列表
func (s *SeckillService) ListSeckillActivities(ctx context.Context, req *v1.ListSeckillActivitiesRequest) (*v1.ListSeckillActivitiesResponse, error) {
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化（数据库未连接）")
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	now := time.Now().Unix()

	rows, total, err := s.svcCtx.SeckillActivityRepo.List(ctx, &repository.ListSeckillActivitiesRequest{
		Page:            page,
		PageSize:        pageSize,
		Status:          req.Status,
		Now:             now,
		IncludeDisabled: req.IncludeDisabled,
	})
	if err != nil {
		logx.Errorf("查询秒杀活动列表失败: %v", err)
		return nil, status.Error(codes.Internal, "查询秒杀活动失败")
	}

	list := make([]*v1.SeckillActivity, 0, len(rows))
	for _, r := range rows {
		actStatus := calcActivityStatus(r.StartTime, r.EndTime, now)

		// 读取 Redis 库存（可选）：用于展示 sold
		//原库存
		stockInit := int64(r.Stock)
		//当前库存
		currentStock := stockInit
		if s.svcCtx.Redis != nil && stockInit > 0 {
			key := fmt.Sprintf("seckill:stock:%d", r.SkuID)
			if v, e := s.svcCtx.Redis.Get(ctx, key).Int64(); e == nil {
				currentStock = v
			}
		}
		//售出
		sold := stockInit - currentStock
		if sold < 0 {
			sold = 0
		}
		if sold > stockInit {
			sold = stockInit
		}

		//原价
		original := r.SkuPrice
		// 价格字符串对齐 proto
		list = append(list, &v1.SeckillActivity{
			Id:            int64(r.ID),
			Name:          r.Name,
			SkuId:         int64(r.SkuID),
			SkuName:       r.SkuName,
			SkuImage:      r.SkuImage,
			SeckillPrice:  fmt.Sprintf("%.2f", r.SeckillPrice),
			OriginalPrice: fmt.Sprintf("%.2f", original),
			Stock:         int32(r.Stock),
			Sold:          int32(sold),
			StartTime:     r.StartTime,
			EndTime:       r.EndTime,
			Status:        actStatus,
			EnableStatus:  int32(r.Status),
		})
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &v1.ListSeckillActivitiesResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.SeckillActivityListData{
			List:       list,
			Page:       int32(page),
			PageSize:   int32(pageSize),
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetSeckillActivity 获取秒杀活动详情
func (s *SeckillService) GetSeckillActivity(ctx context.Context, req *v1.GetSeckillActivityRequest) (*v1.GetSeckillActivityResponse, error) {
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化（数据库未连接）")
	}
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "活动ID不能为空")
	}

	row, err := s.svcCtx.SeckillActivityRepo.GetByID(ctx, uint64(req.Id))
	if err != nil {
		return &v1.GetSeckillActivityResponse{
			Code:    0,
			Message: "成功",
			Data:    nil,
		}, nil
	}
	now := time.Now().Unix()
	//计算活动状态
	actStatus := calcActivityStatus(row.StartTime, row.EndTime, now)

	stockInit := int64(row.Stock)
	currentStock := stockInit
	if s.svcCtx.Redis != nil && stockInit > 0 {
		key := fmt.Sprintf("seckill:stock:%d", row.SkuID)
		if v, e := s.svcCtx.Redis.Get(ctx, key).Int64(); e == nil {
			currentStock = v
		}
	}
	sold := stockInit - currentStock
	if sold < 0 {
		sold = 0
	}
	if sold > stockInit {
		sold = stockInit
	}

	return &v1.GetSeckillActivityResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.SeckillActivity{
			Id:            int64(row.ID),
			Name:          row.Name,
			SkuId:         int64(row.SkuID),
			SkuName:       row.SkuName,
			SkuImage:      row.SkuImage,
			SeckillPrice:  fmt.Sprintf("%.2f", row.SeckillPrice),
			OriginalPrice: fmt.Sprintf("%.2f", row.SkuPrice),
			Stock:         int32(row.Stock),
			Sold:          int32(sold),
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			Status:        actStatus,
			EnableStatus:  int32(row.Status),
		},
	}, nil
}

// CreateSeckillActivity 创建秒杀活动（管理后台）
func (s *SeckillService) CreateSeckillActivity(ctx context.Context, req *v1.CreateSeckillActivityRequest) (*v1.CreateSeckillActivityResponse, error) {
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化（数据库未连接）")
	}
	if req.SkuId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "SKU ID不能为空")
	}
	if req.SeckillPrice == "" {
		return nil, status.Error(codes.InvalidArgument, "秒杀价不能为空")
	}
	if req.StartTime <= 0 || req.EndTime <= 0 || req.EndTime <= req.StartTime {
		return nil, status.Error(codes.InvalidArgument, "活动时间不合法")
	}

	// 活动库存不能超过商品真源库存，否则 Redis 闸门会超发
	if err := s.validateActivityStock(ctx, req.SkuId, req.Stock); err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(req.SeckillPrice, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "秒杀价格式不合法")
	}

	enable := int8(req.EnableStatus)
	if enable != 0 && enable != 1 {
		enable = 1
	}

	act := &model.SeckillActivity{
		Name:         req.Name,
		SkuID:        uint64(req.SkuId),
		SeckillPrice: price,
		Stock:        int(req.Stock),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Status:       enable,
	}
	if err := s.svcCtx.SeckillActivityRepo.Create(ctx, act); err != nil {
		logx.Errorf("创建秒杀活动失败: %v", err)
		return nil, status.Error(codes.Internal, "创建秒杀活动失败")
	}

	// 初始化 Redis 秒杀库存
	if s.svcCtx.Redis != nil && act.Stock > 0 {
		key := fmt.Sprintf("seckill:stock:%d", act.SkuID)
		_ = s.svcCtx.Redis.Set(ctx, key, act.Stock, 0).Err()
	}

	row, _ := s.svcCtx.SeckillActivityRepo.GetByID(ctx, act.ID)
	if row == nil {
		// 兜底返回
		return &v1.CreateSeckillActivityResponse{Code: 0, Message: "成功"}, nil
	}
	now := time.Now().Unix()
	actStatus := calcActivityStatus(row.StartTime, row.EndTime, now)
	return &v1.CreateSeckillActivityResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.SeckillActivity{
			Id:            int64(row.ID),
			Name:          row.Name,
			SkuId:         int64(row.SkuID),
			SkuName:       row.SkuName,
			SkuImage:      row.SkuImage,
			SeckillPrice:  fmt.Sprintf("%.2f", row.SeckillPrice),
			OriginalPrice: fmt.Sprintf("%.2f", row.SkuPrice),
			Stock:         int32(row.Stock),
			Sold:          0,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			Status:        actStatus,
			EnableStatus:  int32(row.Status),
		},
	}, nil
}

// UpdateSeckillActivity 更新秒杀活动（管理后台）
func (s *SeckillService) UpdateSeckillActivity(ctx context.Context, req *v1.UpdateSeckillActivityRequest) (*v1.UpdateSeckillActivityResponse, error) {
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化（数据库未连接）")
	}
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "活动ID不能为空")
	}
	if req.SkuId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "SKU ID不能为空")
	}
	if req.SeckillPrice == "" {
		return nil, status.Error(codes.InvalidArgument, "秒杀价不能为空")
	}
	if req.StartTime <= 0 || req.EndTime <= 0 || req.EndTime <= req.StartTime {
		return nil, status.Error(codes.InvalidArgument, "活动时间不合法")
	}

	// 活动库存不能超过商品真源库存，否则 Redis 闸门会超发
	if err := s.validateActivityStock(ctx, req.SkuId, req.Stock); err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(req.SeckillPrice, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "秒杀价格式不合法")
	}

	enable := int8(req.EnableStatus)
	if enable != 0 && enable != 1 {
		enable = 1
	}

	updates := map[string]any{
		"name":          req.Name,
		"sku_id":        req.SkuId,
		"seckill_price": price,
		"stock":         req.Stock,
		"start_time":    req.StartTime,
		"end_time":      req.EndTime,
		"status":        enable,
	}
	if err := s.svcCtx.SeckillActivityRepo.Update(ctx, uint64(req.Id), updates); err != nil {
		logx.Errorf("更新秒杀活动失败: %v", err)
		return nil, status.Error(codes.Internal, "更新秒杀活动失败")
	}

	//（可选）重置 Redis 秒杀库存为配置库存
	if s.svcCtx.Redis != nil && req.Stock > 0 {
		key := fmt.Sprintf("seckill:stock:%d", req.SkuId)
		_ = s.svcCtx.Redis.Set(ctx, key, req.Stock, 0).Err()
	}

	row, _ := s.svcCtx.SeckillActivityRepo.GetByID(ctx, uint64(req.Id))
	if row == nil {
		return &v1.UpdateSeckillActivityResponse{Code: 0, Message: "成功"}, nil
	}
	now := time.Now().Unix()
	actStatus := calcActivityStatus(row.StartTime, row.EndTime, now)
	return &v1.UpdateSeckillActivityResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.SeckillActivity{
			Id:            int64(row.ID),
			Name:          row.Name,
			SkuId:         int64(row.SkuID),
			SkuName:       row.SkuName,
			SkuImage:      row.SkuImage,
			SeckillPrice:  fmt.Sprintf("%.2f", row.SeckillPrice),
			OriginalPrice: fmt.Sprintf("%.2f", row.SkuPrice),
			Stock:         int32(row.Stock),
			Sold:          0,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			Status:        actStatus,
			EnableStatus:  int32(row.Status),
		},
	}, nil
}

// DeleteSeckillActivity 删除秒杀活动（管理后台）
func (s *SeckillService) DeleteSeckillActivity(ctx context.Context, req *v1.DeleteSeckillActivityRequest) (*v1.DeleteSeckillActivityResponse, error) {
	if s.svcCtx.SeckillActivityRepo == nil {
		return nil, status.Error(codes.FailedPrecondition, "秒杀活动未初始化（数据库未连接）")
	}
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "活动ID不能为空")
	}
	if err := s.svcCtx.SeckillActivityRepo.Delete(ctx, uint64(req.Id)); err != nil {
		logx.Errorf("删除秒杀活动失败: %v", err)
		return nil, status.Error(codes.Internal, "删除秒杀活动失败")
	}
	return &v1.DeleteSeckillActivityResponse{Code: 0, Message: "成功"}, nil
}

// calcActivityStatus 计算活动状态
func calcActivityStatus(start, end, now int64) int32 {
	if start > 0 && now < start {
		return 0
	}
	if end > 0 && now > end {
		return 2
	}
	return 1
}

// rollbackSeckill 回滚一次秒杀预扣（库存 + 防重标记）。
//
// 只在"Redis 扣减已经成功、但后续步骤失败"的分支调用，
// 例如 Kafka 发送失败。回滚本身失败时只记录日志——此时库存与用户标记可能不一致，
// 需要靠后续的对账补偿任务兜底，不能在这里再抛错掩盖原始失败原因。
func (s *SeckillService) rollbackSeckill(ctx context.Context, skuID, userID int64, quantity int32) {
	if s.svcCtx.Redis == nil {
		return
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", skuID)
	userKey := fmt.Sprintf("seckill:user:%d:%d", skuID, userID)

	if _, err := cache.ExecuteLuaScript(ctx, s.svcCtx.Redis, cache.LuaScriptSeckillRollback,
		[]string{stockKey, userKey}, quantity); err != nil {
		logx.Errorf("秒杀预扣回滚失败，需要靠对账补偿: sku_id=%d, user_id=%d, quantity=%d, err=%v",
			skuID, userID, quantity, err)
		return
	}

	logx.Infof("秒杀预扣已回滚: sku_id=%d, user_id=%d, quantity=%d", skuID, userID, quantity)
}

// seckillSoldRow 回源统计已售数量时的行结构
type seckillSoldRow struct {
	Quantity  int       `gorm:"column:quantity"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// ensureSeckillStockGate 确保 Redis 里的秒杀库存闸门存在，缺失时从数据库回源重建。
//
// 重建口径（宁可少卖，不可超卖）：
//
//		剩余 = min(活动库存 - 活动窗口内已售数量, sku 真源库存)
//
//	  - 已售数量按「同一 SKU、秒杀订单、未取消、下单时间不早于活动开始时间」统计，
//	    与消费端扣减真源库存、超时取消后释放配额的口径保持一致；
//	  - 再与 sku.stock 取小：即使统计有偏差，也不会卖出比真源更多的库存；
//	  - 用 SetNX 写入，多个请求同时发现 key 缺失时只有一个生效，避免重复重建把库存放大。
//
// 回源失败时不写 key，保持原有行为（Lua 读不到 key 会按「已抢光」拒绝），不会超卖。
func (s *SeckillService) ensureSeckillStockGate(ctx context.Context, act *model.SeckillActivity) {
	if s.svcCtx.Redis == nil || act == nil || act.Stock <= 0 {
		return
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", act.SkuID)
	exists, err := s.svcCtx.Redis.Exists(ctx, stockKey).Result()
	if err != nil {
		logx.Errorf("检查秒杀库存闸门失败: sku_id=%d, err=%v", act.SkuID, err)
		return
	}
	if exists > 0 {
		return
	}

	remaining, err := s.rebuildSeckillStock(ctx, act)
	if err != nil {
		logx.Errorf("秒杀库存闸门缺失且回源重建失败，本次按已抢光处理: sku_id=%d, activity_id=%d, err=%v",
			act.SkuID, act.ID, err)
		return
	}

	written, err := s.svcCtx.Redis.SetNX(ctx, stockKey, remaining, 0).Result()
	if err != nil {
		logx.Errorf("重建秒杀库存闸门失败: sku_id=%d, err=%v", act.SkuID, err)
		return
	}

	logx.Infof("秒杀库存闸门缺失，已从数据库重建: sku_id=%d, activity_id=%d, remaining=%d, written=%v",
		act.SkuID, act.ID, remaining, written)
}

// rebuildSeckillStock 计算闸门的重建值。
func (s *SeckillService) rebuildSeckillStock(ctx context.Context, act *model.SeckillActivity) (int64, error) {
	if s.svcCtx.DB == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	sold, err := s.seckillSoldQuantity(ctx, act)
	if err != nil {
		return 0, err
	}

	remaining := int64(act.Stock) - sold

	// 与真源库存取小：真源是最终约束，闸门再大也不能卖出超过真源的量。
	// 注意 SKU 不存在时这里查到 0，等价于该活动已无法继续售卖，属于保守结果。
	var realStock int64
	if err := s.svcCtx.DB.WithContext(ctx).Table("sku").
		Select("COALESCE(stock, 0)").Where("id = ?", act.SkuID).Scan(&realStock).Error; err != nil {
		return 0, err
	}
	if realStock < remaining {
		remaining = realStock
	}

	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// seckillSoldQuantity 统计本次活动窗口内已售出的秒杀数量。
//
// 时间比较放在 Go 里而不是 SQL 里用 FROM_UNIXTIME：UNIX_TIMESTAMP 的换算依赖 MySQL
// 会话时区（容器里通常是 UTC），而 created_at 是应用按本地时区写入的 DATETIME，
// 两者混用在时区不一致时会把活动窗口算错 8 小时。这里按驱动解析出的时间去比较，
// 与写入时的时区语义完全一致。
func (s *SeckillService) seckillSoldQuantity(ctx context.Context, act *model.SeckillActivity) (int64, error) {
	var rows []seckillSoldRow
	err := s.svcCtx.DB.WithContext(ctx).Raw(`
		SELECT oi.quantity AS quantity, o.created_at AS created_at
		FROM order_item oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.sku_id = ?
		  AND o.order_type = ?
		  AND o.status <> ?
	`, act.SkuID, constants.OrderTypeSeckill, constants.OrderStatusCanceled).Scan(&rows).Error
	if err != nil {
		return 0, err
	}

	var sold int64
	for _, row := range rows {
		// 只统计本次活动的订单：更早活动卖掉的量不该再从本次活动配额里扣
		if row.CreatedAt.Unix() < act.StartTime {
			continue
		}
		sold += int64(row.Quantity)
	}
	return sold, nil
}

// validateActivityStock 校验秒杀活动配置的库存不超过商品的真源库存。
//
// 为什么必须校验：Redis 里的秒杀库存（闸门）和 MySQL 的 sku.stock（真源）是两个独立数字。
// 如果活动库存配得比真源还大，Redis 会照常放行，但消费端扣真源库存时
// UPDATE ... WHERE stock >= ? 会影响 0 行 —— 即使消费端做了回滚，用户也已经被扣了防重标记、
// 白抢一场；更糟的是如果消费端没做回滚，就会留下「有订单、没扣库存」的脏数据。
// 所以从源头堵住：活动库存必须在真源库存范围内。
func (s *SeckillService) validateActivityStock(ctx context.Context, skuID int64, stock int32) error {
	if s.svcCtx.DB == nil {
		// 数据库不可用时不阻断主流程（此时活动本身也无法持久化）
		return nil
	}

	var realStock int64
	if err := s.svcCtx.DB.WithContext(ctx).Table("sku").
		Select("stock").Where("id = ?", skuID).Scan(&realStock).Error; err != nil {
		logx.Errorf("查询商品真源库存失败: sku_id=%d, err=%v", skuID, err)
		return status.Error(codes.Internal, "查询商品库存失败")
	}

	if int64(stock) > realStock {
		return status.Errorf(codes.InvalidArgument,
			"活动库存(%d)不能超过商品实际库存(%d)", stock, realStock)
	}
	return nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}
