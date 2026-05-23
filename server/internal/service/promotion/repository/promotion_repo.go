package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/promotion/model"
	"gorm.io/gorm"
)

// PromotionRepository 促销活动仓库接口
type PromotionRepository interface {
	// GetList 获取促销活动列表
	GetList(ctx context.Context, productID, categoryID uint64) ([]*model.Promotion, error)
	// GetByID 根据ID获取
	GetByID(ctx context.Context, id uint64) (*model.Promotion, error)
}

type promotionRepository struct {
	db *gorm.DB
}

// NewPromotionRepository 创建促销活动仓库
func NewPromotionRepository(db *gorm.DB) PromotionRepository {
	return &promotionRepository{db: db}
}

// GetList 获取促销活动列表
func (r *promotionRepository) GetList(ctx context.Context, productID, categoryID uint64) ([]*model.Promotion, error) {
	var promotions []*model.Promotion
	query := r.db.WithContext(ctx).Where("status = ?", 1)

	// 这里简化处理，实际应该根据product_ids和category_ids进行JSON查询
	err := query.Order("created_at DESC").Find(&promotions).Error
	return promotions, err
}

// GetByID 根据ID获取
func (r *promotionRepository) GetByID(ctx context.Context, id uint64) (*model.Promotion, error) {
	var promotion model.Promotion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&promotion).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}
