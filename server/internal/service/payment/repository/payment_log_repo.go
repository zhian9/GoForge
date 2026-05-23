package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/payment/model"
	"gorm.io/gorm"
)

// PaymentLogRepository 支付流水仓库接口
type PaymentLogRepository interface {
	// Create 创建支付流水
	Create(ctx context.Context, log *model.PaymentLog) error
	// GetByPaymentNo 根据支付单号获取流水
	GetByPaymentNo(ctx context.Context, paymentNo string) ([]*model.PaymentLog, error)
}

type paymentLogRepository struct {
	db *gorm.DB
}

// NewPaymentLogRepository 创建支付流水仓库
func NewPaymentLogRepository(db *gorm.DB) PaymentLogRepository {
	return &paymentLogRepository{db: db}
}

// Create 创建支付流水
func (r *paymentLogRepository) Create(ctx context.Context, log *model.PaymentLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByPaymentNo 根据支付单号获取流水
func (r *paymentLogRepository) GetByPaymentNo(ctx context.Context, paymentNo string) ([]*model.PaymentLog, error) {
	var logs []*model.PaymentLog
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).
		Order("created_at DESC").Find(&logs).Error
	return logs, err
}
