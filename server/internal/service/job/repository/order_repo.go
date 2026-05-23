package repository

import (
	"context"
	"gorm.io/gorm"
)

// OrderRepository 订单仓库接口（用于定时任务）
type OrderRepository interface {
	// CancelExpiredOrders 取消超时订单
	CancelExpiredOrders(ctx context.Context, timeoutMinutes int) (int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CancelExpiredOrders 取消超时订单
func (r *orderRepository) CancelExpiredOrders(ctx context.Context, timeoutMinutes int) (int64, error) {
	// 实际实现应该查询超时的待支付订单并取消
	// 这里简化处理
	var count int64
	err := r.db.WithContext(ctx).Exec(`
		UPDATE orders 
		SET status = 6 
		WHERE status = 0 
		AND created_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
	`, timeoutMinutes).Count(&count).Error
	return count, err
}
