package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/inventory/model"
	"gorm.io/gorm"
)

// InventoryLogRepository 库存流水仓库接口
type InventoryLogRepository interface {
	// Create 创建库存流水
	Create(ctx context.Context, log *model.InventoryLog) error
	// GetBySkuID 根据SKU ID获取库存流水
	GetBySkuID(ctx context.Context, skuID uint64, page, pageSize int) ([]*model.InventoryLog, int64, error)
}

type inventoryLogRepository struct {
	db *gorm.DB
}

// NewInventoryLogRepository 创建库存流水仓库
func NewInventoryLogRepository(db *gorm.DB) InventoryLogRepository {
	return &inventoryLogRepository{db: db}
}

// Create 创建库存流水
func (r *inventoryLogRepository) Create(ctx context.Context, log *model.InventoryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetBySkuID 根据SKU ID获取库存流水
func (r *inventoryLogRepository) GetBySkuID(ctx context.Context, skuID uint64, page, pageSize int) ([]*model.InventoryLog, int64, error) {
	var logs []*model.InventoryLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.InventoryLog{}).Where("sku_id = ?", skuID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
