package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/logistics/model"
	"gorm.io/gorm"
)

// LogisticsRepository 物流仓库接口
type LogisticsRepository interface {
	// Create 创建物流单
	Create(ctx context.Context, logistics *model.Logistics) error
	// GetByOrderID 根据订单ID获取
	GetByOrderID(ctx context.Context, orderID uint64) (*model.Logistics, error)
	// GetByLogisticsNo 根据物流单号获取
	GetByLogisticsNo(ctx context.Context, logisticsNo string) (*model.Logistics, error)
	// Update 更新物流信息
	Update(ctx context.Context, logistics *model.Logistics) error
}

type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建物流仓库
func NewLogisticsRepository(db *gorm.DB) LogisticsRepository {
	return &logisticsRepository{db: db}
}

// Create 创建物流单
func (r *logisticsRepository) Create(ctx context.Context, logistics *model.Logistics) error {
	return r.db.WithContext(ctx).Create(logistics).Error
}

// GetByOrderID 根据订单ID获取
func (r *logisticsRepository) GetByOrderID(ctx context.Context, orderID uint64) (*model.Logistics, error) {
	var logistics model.Logistics
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&logistics).Error
	if err != nil {
		return nil, err
	}
	return &logistics, nil
}

// GetByLogisticsNo 根据物流单号获取
func (r *logisticsRepository) GetByLogisticsNo(ctx context.Context, logisticsNo string) (*model.Logistics, error) {
	var logistics model.Logistics
	err := r.db.WithContext(ctx).Where("logistics_no = ?", logisticsNo).First(&logistics).Error
	if err != nil {
		return nil, err
	}
	return &logistics, nil
}

// Update 更新物流信息
func (r *logisticsRepository) Update(ctx context.Context, logistics *model.Logistics) error {
	return r.db.WithContext(ctx).Save(logistics).Error
}
