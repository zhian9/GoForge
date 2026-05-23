package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"gorm.io/gorm"
)

// UserCouponRepository 用户优惠券仓库接口
type UserCouponRepository interface {
	// GetByUserID 获取用户优惠券列表
	GetByUserID(ctx context.Context, userID uint64, status int8) ([]*model.UserCoupon, error)
	// Create 创建用户优惠券
	Create(ctx context.Context, userCoupon *model.UserCoupon) error
	// Update 更新用户优惠券
	Update(ctx context.Context, userCoupon *model.UserCoupon) error
	// CountByUserAndCoupon 统计用户已领取的优惠券数量
	CountByUserAndCoupon(ctx context.Context, userID, couponID uint64) (int64, error)
}

type userCouponRepository struct {
	db *gorm.DB
}

// NewUserCouponRepository 创建用户优惠券仓库
func NewUserCouponRepository(db *gorm.DB) UserCouponRepository {
	return &userCouponRepository{db: db}
}

// GetByUserID 获取用户优惠券列表
func (r *userCouponRepository) GetByUserID(ctx context.Context, userID uint64, status int8) ([]*model.UserCoupon, error) {
	var userCoupons []*model.UserCoupon
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&userCoupons).Error
	return userCoupons, err
}

// Create 创建用户优惠券
func (r *userCouponRepository) Create(ctx context.Context, userCoupon *model.UserCoupon) error {
	return r.db.WithContext(ctx).Create(userCoupon).Error
}

// Update 更新用户优惠券
func (r *userCouponRepository) Update(ctx context.Context, userCoupon *model.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

// CountByUserAndCoupon 统计用户已领取的优惠券数量
func (r *userCouponRepository) CountByUserAndCoupon(ctx context.Context, userID, couponID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}
