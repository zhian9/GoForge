package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// ProductSnapshotRepository 从 MySQL 读取商品/SKU 快照，组装 ES 文档
type ProductSnapshotRepository interface {
	BuildProductDocument(ctx context.Context, productID uint64) (map[string]interface{}, error)
}

type productSnapshotRepo struct {
	db *gorm.DB
}

func NewProductSnapshotRepository(db *gorm.DB) ProductSnapshotRepository {
	return &productSnapshotRepo{db: db}
}

type productRow struct {
	ID             uint64     `gorm:"column:id"`
	Name           string     `gorm:"column:name"`
	Subtitle       string     `gorm:"column:subtitle"`
	CategoryID     uint64     `gorm:"column:category_id"`
	BrandID        *uint64    `gorm:"column:brand_id"`
	MainImage      string     `gorm:"column:main_image"`
	LocalMainImage string     `gorm:"column:local_main_image"`
	Detail         string     `gorm:"column:detail"`
	Price          float64    `gorm:"column:price"`
	Sales          int        `gorm:"column:sales"`
	Status         int8       `gorm:"column:status"`
	IsHot          int8       `gorm:"column:is_hot"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type skuRow struct {
	ID        uint64     `gorm:"column:id"`
	ProductID uint64     `gorm:"column:product_id"`
	Name      string     `gorm:"column:name"`
	Specs     string     `gorm:"column:specs"`
	Price     float64    `gorm:"column:price"`
	Stock     int        `gorm:"column:stock"`
	Status    int8       `gorm:"column:status"`
	Image     string     `gorm:"column:image"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (r *productSnapshotRepo) BuildProductDocument(ctx context.Context, productID uint64) (map[string]interface{}, error) {
	if productID == 0 {
		return nil, fmt.Errorf("product_id 不能为空")
	}
	if r.db == nil {
		return nil, fmt.Errorf("db 未初始化")
	}

	var p productRow
	if err := r.db.WithContext(ctx).Table("product").Where("id = ? AND deleted_at IS NULL", productID).First(&p).Error; err != nil {
		return nil, err
	}

	var skus []skuRow
	_ = r.db.WithContext(ctx).
		Table("sku").
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Order("id ASC").
		Find(&skus).Error

	priceMin := math.MaxFloat64
	priceMax := 0.0
	esSkus := make([]map[string]interface{}, 0, len(skus))
	for _, s := range skus {
		// 只索引上架 SKU（避免搜索到无法购买的规格）
		if s.Status != 1 {
			continue
		}
		if s.Price > 0 && s.Price < priceMin {
			priceMin = s.Price
		}
		if s.Price > priceMax {
			priceMax = s.Price
		}
		specsObj := map[string]interface{}{}
		if s.Specs != "" {
			_ = json.Unmarshal([]byte(s.Specs), &specsObj)
		}
		esSkus = append(esSkus, map[string]interface{}{
			"sku_id":   s.ID,
			"sku_name": s.Name,
			"price":    s.Price,
			"stock":    s.Stock,
			"status":   s.Status,
			"image":    s.Image,
			"specs":    specsObj,
		})
	}

	// 如果没有上架 SKU，则兜底用 SPU 价格
	if priceMin == math.MaxFloat64 {
		priceMin = p.Price
		priceMax = p.Price
	}

	// 选择主图：优先本地，其次远程
	mainImage := p.LocalMainImage
	if mainImage == "" {
		mainImage = p.MainImage
	}

	brandID := uint64(0)
	if p.BrandID != nil {
		brandID = *p.BrandID
	}

	doc := map[string]interface{}{
		"product_id":  p.ID,
		"name":        p.Name,
		"subtitle":    p.Subtitle,
		"detail":      p.Detail,
		"category_id": p.CategoryID,
		"brand_id":    brandID,
		"status":      int(p.Status),
		"is_hot":      int(p.IsHot),
		"sales":       p.Sales,
		"main_image":  mainImage,
		"price":       priceMin,
		"price_min":   priceMin,
		"price_max":   priceMax,
		"skus":        esSkus,
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
	return doc, nil
}
