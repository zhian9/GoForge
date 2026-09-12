package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/payment/model"
	"gorm.io/gorm"
)

// PaymentRepository 支付单仓库接口
type PaymentRepository interface {
	// Create 创建支付单
	Create(ctx context.Context, payment *model.Payment) error
	// GetByPaymentNo 根据支付单号获取
	GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error)
	// GetByOrderID 根据订单ID获取
	GetByOrderID(ctx context.Context, orderID uint64) (*model.Payment, error)
	// Update 更新支付单
	Update(ctx context.Context, payment *model.Payment) error
	// UpdateStatus 更新支付状态
	UpdateStatus(ctx context.Context, paymentNo string, status int8) error
	// UpdateStatusAtomic 原子更新状态（带前置状态条件，用于幂等），返回是否命中并更新
	UpdateStatusAtomic(ctx context.Context, paymentNo string, fromStatus int8, updates map[string]interface{}) (bool, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付单仓库
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create 创建支付单
func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// GetByPaymentNo 根据支付单号获取
func (r *paymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByOrderID 根据订单ID获取
func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID uint64) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Update 更新支付单
func (r *paymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// UpdateStatus 更新支付状态
func (r *paymentRepository) UpdateStatus(ctx context.Context, paymentNo string, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("payment_no = ?", paymentNo).
		Update("status", status).Error
}

// UpdateStatusAtomic 原子更新状态（带前置状态条件，用于幂等），返回是否命中并更新
func (r *paymentRepository) UpdateStatusAtomic(ctx context.Context, paymentNo string, fromStatus int8, updates map[string]interface{}) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("payment_no = ? AND status = ?", paymentNo, fromStatus).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
