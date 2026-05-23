package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/order/model"
)

// OrderLogRepository 订单日志数据访问接口
type OrderLogRepository interface {
	Create(ctx context.Context, log *model.OrderLog) error
	GetByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderLog, error)
	GetByOrderNo(ctx context.Context, orderNo string) ([]*model.OrderLog, error)
}

// orderLogRepository 订单日志数据访问实现
type orderLogRepository struct {
	db *gorm.DB
}

// NewOrderLogRepository 创建订单日志仓储
func NewOrderLogRepository(db *gorm.DB) OrderLogRepository {
	return &orderLogRepository{
		db: db,
	}
}

// Create 创建订单日志
func (r *orderLogRepository) Create(ctx context.Context, log *model.OrderLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByOrderID 根据订单ID获取订单日志
func (r *orderLogRepository) GetByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderLog, error) {
	var logs []*model.OrderLog
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetByOrderNo 根据订单号获取订单日志
func (r *orderLogRepository) GetByOrderNo(ctx context.Context, orderNo string) ([]*model.OrderLog, error) {
	var logs []*model.OrderLog
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}
