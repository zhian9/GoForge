package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/order/model"
)

// OrderRepository 订单数据访问接口
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uint64) (*model.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	List(ctx context.Context, req *ListOrdersRequest) ([]*model.Order, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status int8, cancelReason *string) error
	// MarkPaid 标记订单已支付（待支付 -> 待发货），写入支付方式与支付时间
	MarkPaid(ctx context.Context, id uint64, paymentMethod int8) error
	// MarkRefunded 标记订单已退款（仅当订单非已取消/已退款时更新，幂等），返回是否命中
	MarkRefunded(ctx context.Context, id uint64) (bool, error)
}

// ListOrdersRequest 订单列表查询请求
type ListOrdersRequest struct {
	UserID   uint64
	Status   int8 // -1表示全部
	Page     int
	PageSize int
}

// orderRepository 订单数据访问实现
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// Create 创建订单
func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID 根据ID获取订单
func (r *orderRepository) GetByID(ctx context.Context, id uint64) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByOrderNo 根据订单号获取订单
func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// Update 更新订单
func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// List 获取订单列表
func (r *orderRepository) List(ctx context.Context, req *ListOrdersRequest) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Order{})
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 状态过滤：-1表示全部，>=0表示指定状态
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}
	// req.Status = -1 时表示查询全部状态，不应用过滤

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询（包含订单项）
	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").
		Preload("Items"). // 预加载订单项
		Offset(offset).
		Limit(req.PageSize).
		Find(&orders).Error

	return orders, total, err
}

// UpdateStatus 更新订单状态
func (r *orderRepository) UpdateStatus(ctx context.Context, id uint64, status int8, cancelReason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case model.OrderStatusCancelled:
		updates["cancel_time"] = now
		if cancelReason != nil {
			updates["cancel_reason"] = *cancelReason
		}
	case model.OrderStatusPaid:
		updates["payment_time"] = now
	case model.OrderStatusShipped:
		updates["delivery_time"] = now
	case model.OrderStatusCompleted:
		updates["receive_time"] = now
	}

	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// MarkPaid 标记订单已支付（仅当订单处于待支付状态时更新，保证幂等）
func (r *orderRepository) MarkPaid(ctx context.Context, id uint64, paymentMethod int8) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status = ?", id, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":         model.OrderStatusPaid,
			"payment_method": paymentMethod,
			"payment_time":   now,
		}).Error
}

// MarkRefunded 标记订单已退款（仅当订单非已取消/已退款时更新，保证幂等）
func (r *orderRepository) MarkRefunded(ctx context.Context, id uint64) (bool, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status NOT IN (?, ?)", id, model.OrderStatusCancelled, model.OrderStatusRefunded).
		Updates(map[string]interface{}{
			"status":     model.OrderStatusRefunded,
			"updated_at": now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
