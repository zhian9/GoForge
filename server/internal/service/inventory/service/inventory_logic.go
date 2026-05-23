package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/inventory/model"
	"github.com/zhian9/GoForge/server/internal/service/inventory/repository"
	"gorm.io/gorm"
)

// InventoryLogic 库存业务逻辑
type InventoryLogic struct {
	inventoryRepo    repository.InventoryRepository
	inventoryLogRepo repository.InventoryLogRepository
	cache            *cache.CacheOperations
	mqProducer       *mq.Producer
}

// NewInventoryLogic 创建库存业务逻辑
func NewInventoryLogic(
	inventoryRepo repository.InventoryRepository,
	inventoryLogRepo repository.InventoryLogRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
) *InventoryLogic {
	return &InventoryLogic{
		inventoryRepo:    inventoryRepo,
		inventoryLogRepo: inventoryLogRepo,
		cache:            cache,
		mqProducer:       mqProducer,
	}
}

// GetInventoryRequest 获取库存请求
type GetInventoryRequest struct {
	SkuID uint64
}

// GetInventoryResponse 获取库存响应
type GetInventoryResponse struct {
	Inventory *model.Inventory
}

// GetInventory 获取库存（优先从Redis读取）
func (l *InventoryLogic) GetInventory(ctx context.Context, req *GetInventoryRequest) (*GetInventoryResponse, error) {
	// 优先从Redis读取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		stockStr, err := l.cache.Get(ctx, cacheKey)
		if err == nil && stockStr != "" {
			// 从Redis读取成功，构造返回对象
			var stock int
			fmt.Sscanf(stockStr, "%d", &stock)
			inventory := &model.Inventory{
				SkuID:          req.SkuID,
				AvailableStock: stock,
			}
			return &GetInventoryResponse{Inventory: inventory}, nil
		}
	}

	// 从数据库读取
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("库存不存在")
		}
		return nil, apperrors.NewInternalError("查询库存失败: " + err.Error())
	}
	if inventory == nil {
		return nil, apperrors.NewNotFoundError("库存不存在")
	}

	// 同步到Redis
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_ = l.cache.Set(ctx, cacheKey, inventory.AvailableStock, 0) // 永不过期
	}

	return &GetInventoryResponse{
		Inventory: inventory,
	}, nil
}

// BatchGetInventoryRequest 批量获取库存请求
type BatchGetInventoryRequest struct {
	SkuIDs []uint64
}

// BatchGetInventoryResponse 批量获取库存响应
type BatchGetInventoryResponse struct {
	Inventories []*model.Inventory
}

// BatchGetInventory 批量获取库存
func (l *InventoryLogic) BatchGetInventory(ctx context.Context, req *BatchGetInventoryRequest) (*BatchGetInventoryResponse, error) {
	inventories, err := l.inventoryRepo.BatchGetBySkuIDs(ctx, req.SkuIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("批量获取库存失败")
	}

	return &BatchGetInventoryResponse{
		Inventories: inventories,
	}, nil
}

// LockStockRequest 锁定库存请求
type LockStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// LockStock 锁定库存（预占）
func (l *InventoryLogic) LockStock(ctx context.Context, req *LockStockRequest) error {
	// 获取当前库存
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		return apperrors.NewNotFoundError("库存不存在")
	}

	// 检查可用库存是否充足
	if inventory.AvailableStock < req.Quantity {
		return apperrors.NewError(5000, "库存不足")
	}

	// 锁定库存
	err = l.inventoryRepo.LockStock(ctx, req.SkuID, req.Quantity)
	if err != nil {
		return apperrors.NewInternalError("锁定库存失败")
	}

	// 记录库存流水
	log := &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        3, // 锁定
		Quantity:    req.Quantity,
		BeforeStock: inventory.AvailableStock,
		AfterStock:  inventory.AvailableStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	}
	_ = l.inventoryLogRepo.Create(ctx, log)

	return nil
}

// DeductStockRequest 扣减库存请求
type DeductStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// DeductStock 扣减库存（使用Redis保证原子性）
func (l *InventoryLogic) DeductStock(ctx context.Context, req *DeductStockRequest) error {
	if l.cache == nil {
		return apperrors.NewInternalError("Redis未初始化")
	}

	// 使用原子操作扣减Redis库存
	cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
	// 先检查库存
	currentStock, err := l.cache.Get(ctx, cacheKey)
	if err != nil {
		// Redis nil 错误表示库存不存在，尝试从数据库加载
		if err.Error() == "redis: nil" {
			// 从数据库读取库存
			inventory, dbErr := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
			if dbErr != nil || inventory == nil {
				return apperrors.NewNotFoundError("库存不存在")
			}
			// 同步到Redis
			_ = l.cache.Set(ctx, cacheKey, inventory.AvailableStock, 0)
			currentStock = fmt.Sprintf("%d", inventory.AvailableStock)
		} else {
			return apperrors.NewInternalError("获取库存失败: " + err.Error())
		}
	}
	var stock int
	fmt.Sscanf(currentStock, "%d", &stock)
	if stock < req.Quantity {
		return apperrors.NewError(5000, "库存不足")
	}
	// 扣减库存（使用DecrementBy保证原子性）
	newStock, err := l.cache.DecrementBy(ctx, cacheKey, int64(req.Quantity))
	if err != nil {
		return apperrors.NewInternalError("扣减库存失败: " + err.Error())
	}

	if newStock < 0 {
		return apperrors.NewError(5000, "库存不足")
	}

	// 发送Kafka消息，异步同步到MySQL
	if l.mqProducer != nil {
		message := mq.NewMessage(mq.TopicInventoryDeducted, map[string]interface{}{
			"sku_id":    req.SkuID,
			"quantity":  req.Quantity,
			"order_id":  req.OrderID,
			"new_stock": newStock,
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicInventoryDeducted, fmt.Sprintf("%d", req.SkuID), message)
	}

	return nil
}

// UnlockStockRequest 解锁库存请求
type UnlockStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// UnlockStock 解锁库存
func (l *InventoryLogic) UnlockStock(ctx context.Context, req *UnlockStockRequest) error {
	// 获取当前库存
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		return apperrors.NewNotFoundError("库存不存在")
	}

	// 解锁库存
	err = l.inventoryRepo.UnlockStock(ctx, req.SkuID, req.Quantity)
	if err != nil {
		return apperrors.NewInternalError("解锁库存失败")
	}

	// 记录库存流水
	log := &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        4, // 解锁
		Quantity:    req.Quantity,
		BeforeStock: inventory.LockedStock,
		AfterStock:  inventory.LockedStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	}
	_ = l.inventoryLogRepo.Create(ctx, log)

	return nil
}

// RollbackStockRequest 回退库存请求
type RollbackStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// RollbackStock 回退库存
func (l *InventoryLogic) RollbackStock(ctx context.Context, req *RollbackStockRequest) error {
	// 获取当前库存
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		return apperrors.NewNotFoundError("库存不存在")
	}

	// 回退库存
	err = l.inventoryRepo.RollbackStock(ctx, req.SkuID, req.Quantity)
	if err != nil {
		return apperrors.NewInternalError("回退库存失败")
	}

	// 记录库存流水
	log := &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        6, // 回退
		Quantity:    req.Quantity,
		BeforeStock: inventory.SoldStock,
		AfterStock:  inventory.SoldStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	}
	_ = l.inventoryLogRepo.Create(ctx, log)

	return nil
}

// StockInRequest 入库请求
type StockInRequest struct {
	SkuID    uint64
	Quantity int
	Remark   string
}

// StockIn 入库
func (l *InventoryLogic) StockIn(ctx context.Context, req *StockInRequest) error {
	// 获取当前库存
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		// 如果库存不存在，创建新记录
		inventory = &model.Inventory{
			SkuID:             req.SkuID,
			TotalStock:        req.Quantity,
			AvailableStock:    req.Quantity,
			LockedStock:       0,
			SoldStock:         0,
			LowStockThreshold: 10,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		err = l.inventoryRepo.Create(ctx, inventory)
		if err != nil {
			return apperrors.NewInternalError("创建库存记录失败")
		}
	} else {
		// 更新库存
		err = l.inventoryRepo.StockIn(ctx, req.SkuID, req.Quantity)
		if err != nil {
			return apperrors.NewInternalError("入库失败")
		}
	}

	// 记录库存流水
	log := &model.InventoryLog{
		SkuID:       req.SkuID,
		Type:        1, // 入库
		Quantity:    req.Quantity,
		BeforeStock: inventory.AvailableStock,
		AfterStock:  inventory.AvailableStock + req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	}
	_ = l.inventoryLogRepo.Create(ctx, log)

	return nil
}

// GetInventoryLogRequest 获取库存流水请求
type GetInventoryLogRequest struct {
	SkuID    uint64
	Page     int
	PageSize int
}

// GetInventoryLogResponse 获取库存流水响应
type GetInventoryLogResponse struct {
	Logs  []*model.InventoryLog
	Total int64
}

// GetInventoryLog 获取库存流水
func (l *InventoryLogic) GetInventoryLog(ctx context.Context, req *GetInventoryLogRequest) (*GetInventoryLogResponse, error) {
	logs, total, err := l.inventoryLogRepo.GetBySkuID(ctx, req.SkuID, req.Page, req.PageSize)
	if err != nil {
		return nil, apperrors.NewInternalError("获取库存流水失败")
	}

	return &GetInventoryLogResponse{
		Logs:  logs,
		Total: total,
	}, nil
}
