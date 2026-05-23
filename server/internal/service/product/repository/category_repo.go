package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/product/model"
)

// CategoryRepository 类目数据访问接口
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	GetByID(ctx context.Context, id uint64) (*model.Category, error)
	GetByParentID(ctx context.Context, parentID uint64) ([]*model.Category, error)
	GetAll(ctx context.Context, status int8) ([]*model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uint64) error
}

// categoryRepository 类目数据访问实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建类目数据访问实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create 创建类目
func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID 根据ID获取类目
func (r *categoryRepository) GetByID(ctx context.Context, id uint64) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByParentID 根据父ID获取子类目列表
func (r *categoryRepository) GetByParentID(ctx context.Context, parentID uint64) ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort DESC, id ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// GetAll 获取所有类目
func (r *categoryRepository) GetAll(ctx context.Context, status int8) ([]*model.Category, error) {
	var categories []*model.Category
	query := r.db.WithContext(ctx)
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Order("level ASC, sort DESC, id ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// Update 更新类目
func (r *categoryRepository) Update(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete 删除类目
func (r *categoryRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}
