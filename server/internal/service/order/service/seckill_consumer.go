package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/order/model"
	"github.com/zhian9/GoForge/server/internal/service/order/repository"
)

// seckillActivitySnapshot 用于读取秒杀活动 + SKU 的快照信息
type seckillActivitySnapshot struct {
	SeckillPrice  float64 `gorm:"column:seckill_price"`
	OriginalPrice float64 `gorm:"column:original_price"`
	SkuName       string  `gorm:"column:sku_name"`
	SkuImage      string  `gorm:"column:sku_image"`
	ProductID     uint64  `gorm:"column:product_id"`
}

// SeckillMessage 秒杀消息结构
type SeckillMessage struct {
	UserID    int64 `json:"user_id"`
	SkuID     int64 `json:"sku_id"`
	Quantity  int   `json:"quantity"`
	Timestamp int64 `json:"timestamp"`
}

// SeckillConsumer 秒杀订单消费者
type SeckillConsumer struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	db            *gorm.DB
	cache         *cache.CacheOperations
}

// NewSeckillConsumer 创建秒杀消费者
func NewSeckillConsumer(orderRepo repository.OrderRepository, orderItemRepo repository.OrderItemRepository, db *gorm.DB, cacheOps *cache.CacheOperations) *SeckillConsumer {
	return &SeckillConsumer{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		db:            db,
		cache:         cacheOps,
	}
}

// Consume 消费秒杀消息
func (c *SeckillConsumer) Consume(ctx context.Context, message *mq.Message) error {
	if c.db == nil {
		return fmt.Errorf("数据库未初始化，无法消费秒杀订单消息")
	}

	// 解析消息数据
	var seckillMsg SeckillMessage
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("序列化消息数据失败: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &seckillMsg); err != nil {
		return fmt.Errorf("解析秒杀消息失败: %w", err)
	}

	logx.Infof("收到秒杀消息: user_id=%d, sku_id=%d, quantity=%d",
		seckillMsg.UserID, seckillMsg.SkuID, seckillMsg.Quantity)

	// 幂等性校验：检查是否已经创建过订单
	exists, err := c.checkOrderExists(ctx, seckillMsg.UserID, seckillMsg.SkuID)
	if err != nil {
		logx.Errorf("检查订单是否存在失败: %v", err)
		return err
	}
	if exists {
		logx.Infof("订单已存在，跳过处理: user_id=%d, sku_id=%d", seckillMsg.UserID, seckillMsg.SkuID)
		return nil // 幂等，直接返回成功
	}

	// 查询秒杀活动价格与商品信息快照
	activitySnapshot, err := c.getSeckillSnapshot(ctx, seckillMsg.SkuID)
	if err != nil {
		logx.Errorf("查询秒杀活动价格失败: %v", err)
	}

	price := 0.0
	if activitySnapshot != nil {
		price = activitySnapshot.SeckillPrice
	}
	totalAmount := price * float64(seckillMsg.Quantity)

	// 创建订单
	orderNo := c.generateOrderNo(ctx)
	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          uint64(seckillMsg.UserID),
		OrderType:       2,                        // 秒杀订单
		Status:          model.OrderStatusPending, // 待支付
		TotalAmount:     totalAmount,
		PayAmount:       totalAmount,
		DiscountAmount:  0,
		FreightAmount:   0,
		ReceiverName:    "", // 需要从地址服务获取
		ReceiverPhone:   "",
		ReceiverAddress: "",
	}

	// 开启事务
	tx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建订单
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建订单失败: %w", err)
	}

	// 使用秒杀活动快照填充订单项
	orderItem := &model.OrderItem{
		OrderID:     order.ID,
		OrderNo:     orderNo,
		ProductID:   0, // 可选：需要从商品服务获取
		ProductName: fmt.Sprintf("秒杀商品-%d", seckillMsg.SkuID),
		SkuID:       uint64(seckillMsg.SkuID),
		SkuCode:     "",
		SkuName:     fmt.Sprintf("秒杀SKU-%d", seckillMsg.SkuID),
		Price:       price,
		Quantity:    seckillMsg.Quantity,
		TotalAmount: totalAmount,
	}

	if activitySnapshot != nil {
		if activitySnapshot.ProductID != 0 {
			orderItem.ProductID = activitySnapshot.ProductID
		}
		if activitySnapshot.SkuName != "" {
			orderItem.SkuName = activitySnapshot.SkuName
			orderItem.ProductName = activitySnapshot.SkuName
		}
		if activitySnapshot.SkuImage != "" {
			orderItem.SkuImage = &activitySnapshot.SkuImage
		}
	}

	if err := tx.Create(orderItem).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建订单项失败: %w", err)
	}

	// 扣减真源库存：必须与订单写在同一个事务里。
	//
	// 这里原先的实现是「先提交订单 → 再扣 sku.stock → 失败只打日志」，后果是：
	// 只要 Redis 闸门放行的数量超过 sku.stock（例如秒杀活动配的库存大于实际库存），
	// 那句 UPDATE ... WHERE stock >= ? 就会影响 0 行，而订单已经提交，
	// 于是产生「有订单、没扣库存」的脏数据，真源与订单脱节。
	//
	// 现在把扣减放进同一事务：库存不足 → 整个订单回滚，不会留下无库存支撑的订单。
	// cacheOps 传 nil，缓存失效等到事务提交后再做（回滚时不必做无意义的缓存失效）。
	if err := deductStock(ctx, tx, nil, uint64(seckillMsg.SkuID), seckillMsg.Quantity); err != nil {
		tx.Rollback()
		return fmt.Errorf("秒杀订单扣减真源库存失败，订单已回滚: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务提交成功后再失效商品缓存（列表 / 详情 / SKU 信息）
	invalidateProductCache(ctx, c.db, c.cache, uint64(seckillMsg.SkuID))

	logx.Infof("秒杀订单创建成功: order_no=%s, user_id=%d, sku_id=%d",
		orderNo, seckillMsg.UserID, seckillMsg.SkuID)

	return nil
}

// checkOrderExists 检查订单是否存在（幂等性校验）
func (c *SeckillConsumer) checkOrderExists(ctx context.Context, userID, skuID int64) (bool, error) {
	if c.db == nil {
		return false, fmt.Errorf("数据库未初始化")
	}

	// 查询该用户是否已经为该SKU创建过秒杀订单
	var count int64
	err := c.db.WithContext(ctx).
		Model(&model.Order{}).
		Joins("JOIN order_item ON orders.id = order_item.order_id").
		Where("orders.user_id = ? AND order_item.sku_id = ? AND orders.order_type = ?",
			userID, skuID, 2). // order_type = 2 表示秒杀订单
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// generateOrderNo 生成订单号
func (c *SeckillConsumer) generateOrderNo(ctx context.Context) string {
	// 秒杀订单号格式：SECKILL + 时间戳 + 随机数
	return fmt.Sprintf("SECKILL%d%06d", time.Now().Unix(), time.Now().UnixNano()%1000000)
}

// getSeckillSnapshot 获取秒杀活动的价格和商品信息快照
func (c *SeckillConsumer) getSeckillSnapshot(ctx context.Context, skuID int64) (*seckillActivitySnapshot, error) {
	var snap seckillActivitySnapshot
	err := c.db.WithContext(ctx).
		Table("seckill_activity sa").
		Joins("LEFT JOIN sku s ON sa.sku_id = s.id").
		// gorm.Select 需要单个字符串（多个字符串会被当成占位参数，导致字段取不到）
		Select("sa.seckill_price AS seckill_price, s.original_price AS original_price, s.name AS sku_name, s.image AS sku_image, s.product_id AS product_id").
		Where("sa.sku_id = ?", skuID).
		Where("sa.status = 1"). // 已启用
		Order("sa.id DESC").
		Limit(1).
		Scan(&snap).Error
	if err != nil {
		return nil, err
	}
	// 如果所有字段都是零值，认为未查到
	if snap.SeckillPrice == 0 && snap.OriginalPrice == 0 && snap.SkuName == "" && snap.SkuImage == "" {
		return nil, nil
	}
	return &snap, nil
}
