package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/inventory/model"
	"gorm.io/gorm"
)

// InventoryRepository 库存仓库接口
type InventoryRepository interface {
	// GetBySkuID 根据SKU ID获取库存
	GetBySkuID(ctx context.Context, skuID uint64) (*model.Inventory, error)
	// BatchGetBySkuIDs 批量获取库存
	BatchGetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*model.Inventory, error)
	// Create 创建库存记录
	Create(ctx context.Context, inventory *model.Inventory) error
	// Update 更新库存
	Update(ctx context.Context, inventory *model.Inventory) error
	// LockStock 锁定库存（使用乐观锁）
	LockStock(ctx context.Context, skuID uint64, quantity int) error
	// DeductStock 扣减库存
	DeductStock(ctx context.Context, skuID uint64, quantity int) error
	// UnlockStock 解锁库存
	UnlockStock(ctx context.Context, skuID uint64, quantity int) error
	// RollbackStock 回退库存
	RollbackStock(ctx context.Context, skuID uint64, quantity int) error
	// StockIn 入库
	StockIn(ctx context.Context, skuID uint64, quantity int) error
}

type inventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository 创建库存仓库
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

// GetBySkuID 根据SKU ID获取库存
func (r *inventoryRepository) GetBySkuID(ctx context.Context, skuID uint64) (*model.Inventory, error) {
	var inventory model.Inventory
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// BatchGetBySkuIDs 批量获取库存
func (r *inventoryRepository) BatchGetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*model.Inventory, error) {
	var inventories []*model.Inventory
	err := r.db.WithContext(ctx).Where("sku_id IN ?", skuIDs).Find(&inventories).Error
	return inventories, err
}

// Create 创建库存记录
func (r *inventoryRepository) Create(ctx context.Context, inventory *model.Inventory) error {
	return r.db.WithContext(ctx).Create(inventory).Error
}

// Update 更新库存
func (r *inventoryRepository) Update(ctx context.Context, inventory *model.Inventory) error {
	return r.db.WithContext(ctx).Save(inventory).Error
}

// LockStock 锁定库存（使用乐观锁）
func (r *inventoryRepository) LockStock(ctx context.Context, skuID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("sku_id = ? AND available_stock >= ?", skuID, quantity).
		Updates(map[string]interface{}{
			"available_stock": gorm.Expr("available_stock - ?", quantity),
			"locked_stock":    gorm.Expr("locked_stock + ?", quantity),
		}).Error
}

// DeductStock 扣减库存
func (r *inventoryRepository) DeductStock(ctx context.Context, skuID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("sku_id = ?", skuID).
		Updates(map[string]interface{}{
			"locked_stock": gorm.Expr("locked_stock - ?", quantity),
			"sold_stock":   gorm.Expr("sold_stock + ?", quantity),
		}).Error
}

// UnlockStock 解锁库存
func (r *inventoryRepository) UnlockStock(ctx context.Context, skuID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("sku_id = ?", skuID).
		Updates(map[string]interface{}{
			"locked_stock":    gorm.Expr("locked_stock - ?", quantity),
			"available_stock": gorm.Expr("available_stock + ?", quantity),
		}).Error
}

// RollbackStock 回退库存
func (r *inventoryRepository) RollbackStock(ctx context.Context, skuID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("sku_id = ?", skuID).
		Updates(map[string]interface{}{
			"sold_stock":      gorm.Expr("sold_stock - ?", quantity),
			"available_stock": gorm.Expr("available_stock + ?", quantity),
		}).Error
}

// StockIn 入库
func (r *inventoryRepository) StockIn(ctx context.Context, skuID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("sku_id = ?", skuID).
		Updates(map[string]interface{}{
			"total_stock":     gorm.Expr("total_stock + ?", quantity),
			"available_stock": gorm.Expr("available_stock + ?", quantity),
		}).Error
}
