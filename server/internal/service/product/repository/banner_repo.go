package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/product/model"
)

// BannerRepository Banner数据访问接口
type BannerRepository interface {
	Create(ctx context.Context, banner *model.Banner) error
	GetByID(ctx context.Context, id uint64) (*model.Banner, error)
	GetAll(ctx context.Context, status int8, limit int) ([]*model.Banner, error)
	Update(ctx context.Context, banner *model.Banner) error
	Delete(ctx context.Context, id uint64) error
}

// bannerRepository Banner数据访问实现
type bannerRepository struct {
	db *gorm.DB
}

// NewBannerRepository 创建Banner数据访问实例
func NewBannerRepository(db *gorm.DB) BannerRepository {
	return &bannerRepository{db: db}
}

// Create 创建Banner
func (r *bannerRepository) Create(ctx context.Context, banner *model.Banner) error {
	return r.db.WithContext(ctx).Create(banner).Error
}

// GetByID 根据ID获取Banner
func (r *bannerRepository) GetByID(ctx context.Context, id uint64) (*model.Banner, error) {
	var banner model.Banner
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&banner).Error
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// GetAll 获取所有Banner
func (r *bannerRepository) GetAll(ctx context.Context, status int8, limit int) ([]*model.Banner, error) {
	var banners []*model.Banner
	query := r.db.WithContext(ctx)

	if status > 0 {
		query = query.Where("status = ?", status)
	}

	// 按排序值降序，ID升序
	query = query.Order("sort DESC, id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&banners).Error
	if err != nil {
		return nil, err
	}
	return banners, nil
}

// Update 更新Banner
func (r *bannerRepository) Update(ctx context.Context, banner *model.Banner) error {
	return r.db.WithContext(ctx).Save(banner).Error
}

// Delete 删除Banner
func (r *bannerRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Banner{}, id).Error
}
