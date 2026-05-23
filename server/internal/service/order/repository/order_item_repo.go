package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/order/model"
)

// OrderItemRepository 订单商品项数据访问接口
type OrderItemRepository interface {
	Create(ctx context.Context, item *model.OrderItem) error
	CreateBatch(ctx context.Context, items []*model.OrderItem) error
	GetByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderItem, error)
	GetByOrderNo(ctx context.Context, orderNo string) ([]*model.OrderItem, error)
}

// orderItemRepository 订单商品项数据访问实现
type orderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository 创建订单商品项仓储
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{
		db: db,
	}
}

// Create 创建订单商品项
func (r *orderItemRepository) Create(ctx context.Context, item *model.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// CreateBatch 批量创建订单商品项
func (r *orderItemRepository) CreateBatch(ctx context.Context, items []*model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}

// GetByOrderID 根据订单ID获取订单商品项
func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// GetByOrderNo 根据订单号获取订单商品项
func (r *orderItemRepository) GetByOrderNo(ctx context.Context, orderNo string) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).Find(&items).Error
	return items, err
}
