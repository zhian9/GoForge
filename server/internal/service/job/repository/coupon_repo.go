package repository

import (
	"context"
	"gorm.io/gorm"
)

// CouponRepository 优惠券仓库接口（用于定时任务）
type CouponRepository interface {
	// ProcessExpiredCoupons 处理过期优惠券
	ProcessExpiredCoupons(ctx context.Context) (int64, error)
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓库
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

// ProcessExpiredCoupons 处理过期优惠券
func (r *couponRepository) ProcessExpiredCoupons(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Exec(`
		UPDATE user_coupon 
		SET status = 2 
		WHERE status = 0 
		AND expire_at < NOW()
	`).Count(&count).Error
	return count, err
}
