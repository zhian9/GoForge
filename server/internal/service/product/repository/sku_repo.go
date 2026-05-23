package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/service/product/model"
)

// ListSkusRequest SKU列表查询请求
type ListSkusRequest struct {
	ProductID uint64
	Status    int8 // -1-全部, 0-下架, 1-上架
	Page      int
	PageSize  int
}

// SkuRepository SKU数据访问接口
type SkuRepository interface {
	Create(ctx context.Context, sku *model.Sku) error
	GetByID(ctx context.Context, id uint64) (*model.Sku, error)
	GetBySkuCode(ctx context.Context, skuCode string) (*model.Sku, error)
	GetBySkuCodeUnscoped(ctx context.Context, skuCode string) (*model.Sku, error)
	GetByProductID(ctx context.Context, productID uint64) ([]*model.Sku, error)
	List(ctx context.Context, req *ListSkusRequest) ([]*model.Sku, int64, error)
	GetAggByProductIDs(ctx context.Context, productIDs []uint64, status int8) (map[uint64]SkuAgg, error)
	RestoreAndUpdateByID(ctx context.Context, id uint64, updates map[string]any) error
	Update(ctx context.Context, sku *model.Sku) error
	Delete(ctx context.Context, id uint64) error
}

// SkuAgg SKU 聚合信息（用于列表页展示最低价/总库存）
type SkuAgg struct {
	MinPrice   float64
	TotalStock int64
	// MinOriginalPrice 如果需要在商品列表展示"原价"，可用它兜底；为空表示所有 SKU 都没有原价
	MinOriginalPrice *float64
}

// skuRepository SKU数据访问实现
type skuRepository struct {
	db *gorm.DB
}

// NewSkuRepository 创建SKU数据访问实例
func NewSkuRepository(db *gorm.DB) SkuRepository {
	return &skuRepository{db: db}
}

// Create 创建SKU
func (r *skuRepository) Create(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

// GetByID 根据ID获取SKU
func (r *skuRepository) GetByID(ctx context.Context, id uint64) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// GetBySkuCode 根据SKU编码获取SKU
func (r *skuRepository) GetBySkuCode(ctx context.Context, skuCode string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Where("sku_code = ?", skuCode).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// GetBySkuCodeUnscoped 根据SKU编码获取SKU（包含软删除数据）
func (r *skuRepository) GetBySkuCodeUnscoped(ctx context.Context, skuCode string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Unscoped().Where("sku_code = ?", skuCode).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// GetByProductID 根据商品ID获取SKU列表
func (r *skuRepository) GetByProductID(ctx context.Context, productID uint64) ([]*model.Sku, error) {
	var skus []*model.Sku
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("id ASC").
		Find(&skus).Error
	if err != nil {
		return nil, err
	}
	return skus, nil
}

// Update 更新SKU
func (r *skuRepository) Update(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

// List 获取SKU列表（管理后台）
func (r *skuRepository) List(ctx context.Context, req *ListSkusRequest) ([]*model.Sku, int64, error) {
	var skus []*model.Sku
	var total int64

	// 管理后台查询，明确指定表名和软删除条件
	query := r.db.WithContext(ctx).Table("sku").Where("deleted_at IS NULL")

	// 条件过滤
	if req.ProductID > 0 {
		query = query.Where("product_id = ?", req.ProductID)
	}
	// 只有当 status >= 0 时才过滤（-1 表示查询所有状态）
	if req.Status >= 0 {
		query = query.Where("status = ?", req.Status)
	}

	// 分页参数处理
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 先统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表，使用 Find 方法映射到结构体
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&skus).Error; err != nil {
		return nil, 0, err
	}

	return skus, total, nil
}

// GetAggByProductIDs 按商品ID批量聚合 SKU：最低价 / 总库存（可选原价）
func (r *skuRepository) GetAggByProductIDs(ctx context.Context, productIDs []uint64, status int8) (map[uint64]SkuAgg, error) {
	if len(productIDs) == 0 {
		return map[uint64]SkuAgg{}, nil
	}

	type aggRow struct {
		ProductID        uint64          `gorm:"column:product_id"`
		MinPrice         float64         `gorm:"column:min_price"`
		TotalStock       int64           `gorm:"column:total_stock"`
		MinOriginalPrice sql.NullFloat64 `gorm:"column:min_original_price"`
	}

	query := r.db.WithContext(ctx).
		Table("sku").
		Select(`product_id,
			MIN(price) AS min_price,
			SUM(stock) AS total_stock,
			MIN(CASE WHEN original_price IS NULL THEN NULL ELSE original_price END) AS min_original_price`).
		Where("deleted_at IS NULL").
		Where("product_id IN ?", productIDs).
		Group("product_id")

	// status >= 0 时才过滤；-1 表示不按状态过滤
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	var rows []aggRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[uint64]SkuAgg, len(rows))
	for _, r := range rows {
		var minOriginal *float64
		if r.MinOriginalPrice.Valid {
			v := r.MinOriginalPrice.Float64
			minOriginal = &v
		}
		out[r.ProductID] = SkuAgg{
			MinPrice:         r.MinPrice,
			TotalStock:       r.TotalStock,
			MinOriginalPrice: minOriginal,
		}
	}
	return out, nil
}

// RestoreAndUpdateByID 恢复软删除记录并更新字段（解决：删了同一个 sku_code 之后无法再新增的问题）
func (r *skuRepository) RestoreAndUpdateByID(ctx context.Context, id uint64, updates map[string]any) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}
	if updates == nil {
		updates = map[string]any{}
	}
	// 关键：把 deleted_at 置 NULL 才能恢复
	updates["deleted_at"] = nil
	return r.db.WithContext(ctx).Unscoped().Model(&model.Sku{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除SKU（软删除）
func (r *skuRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Sku{}, id).Error
}
