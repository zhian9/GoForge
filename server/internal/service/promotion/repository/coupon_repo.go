package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"gorm.io/gorm"
)

// CouponRepository 优惠券仓库接口
type CouponRepository interface {
	// GetList 获取优惠券列表
	GetList(ctx context.Context, page, pageSize int) ([]*model.Coupon, int64, error)
	// GetByID 根据ID获取
	GetByID(ctx context.Context, id uint64) (*model.Coupon, error)
	// Create 创建优惠券
	Create(ctx context.Context, coupon *model.Coupon) error
	// Update 更新优惠券
	Update(ctx context.Context, coupon *model.Coupon) error
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓库
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

// GetList 获取优惠券列表
func (r *couponRepository) GetList(ctx context.Context, page, pageSize int) ([]*model.Coupon, int64, error) {
	var coupons []*model.Coupon
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Coupon{}).Where("status = ?", 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&coupons).Error
	return coupons, total, err
}

// GetByID 根据ID获取
func (r *couponRepository) GetByID(ctx context.Context, id uint64) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// Create 创建优惠券
func (r *couponRepository) Create(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

// Update 更新优惠券
func (r *couponRepository) Update(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}
