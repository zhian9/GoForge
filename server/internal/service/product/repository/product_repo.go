package repository

import (
	"context"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/product/model"
)

// ProductRepository 商品数据访问接口
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uint64) (*model.Product, error)
	GetBySpuCode(ctx context.Context, spuCode string) (*model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *ListProductsRequest) ([]*model.Product, int64, error)
}

// ListProductsRequest 商品列表查询请求
type ListProductsRequest struct {
	CategoryID uint64
	// CategoryIDs: 当需要"点击主分类展示其所有子分类商品"时使用（category_id IN (...)）
	// 若该字段非空，则优先使用该字段过滤，并忽略 CategoryID。
	CategoryIDs []uint64
	BrandID     uint64
	Keyword     string
	Status      int8
	IsHot       int8 // -1-全部, 0-否, 1-是
	Page        int
	PageSize    int
	Sort        string // price_asc, price_desc, sales_desc, created_desc
}

// productRepository 商品数据访问实现
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品数据访问实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// Create 创建商品
func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// GetByID 根据ID获取商品
func (r *productRepository) GetByID(ctx context.Context, id uint64) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetBySpuCode 根据SPU编码获取商品
func (r *productRepository) GetBySpuCode(ctx context.Context, spuCode string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("spu_code = ?", spuCode).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Update 更新商品
func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete 删除商品（软删除）
func (r *productRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}

// List 获取商品列表
func (r *productRepository) List(ctx context.Context, req *ListProductsRequest) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Product{})

	// 条件过滤
	if len(req.CategoryIDs) > 0 {
		query = query.Where("category_id IN ?", req.CategoryIDs)
	} else if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	if req.BrandID > 0 {
		query = query.Where("brand_id = ?", req.BrandID)
	}
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR subtitle LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	// 只有当 status >= 0 时才过滤（-1 表示查询所有状态，0 表示查询 status=0 的商品）
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}
	// 只有当 is_hot >= 0 时才过滤（-1 表示查询全部，0 表示查询非热门，1 表示查询热门）
	if req.IsHot >= 0 {
		query = query.Where("is_hot = ?", req.IsHot)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	switch req.Sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	case "sales_desc":
		query = query.Order("sales DESC")
	case "created_desc":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("sort DESC, created_at DESC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
