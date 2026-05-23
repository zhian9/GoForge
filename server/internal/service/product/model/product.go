package model

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品模型（SPU）
type Product struct {
	ID             uint64         `gorm:"primaryKey;column:id" json:"id"`
	SpuCode        string         `gorm:"column:spu_code;uniqueIndex;not null;size:50" json:"spu_code"`
	Name           string         `gorm:"column:name;not null;size:200" json:"name"`
	Subtitle       string         `gorm:"column:subtitle;size:200" json:"subtitle"`
	CategoryID     uint64         `gorm:"column:category_id;not null;index" json:"category_id"`
	BrandID        *uint64        `gorm:"column:brand_id;index" json:"brand_id"`
	MainImage      string         `gorm:"column:main_image;size:255" json:"main_image"`             // 主图URL
	LocalMainImage string         `gorm:"column:local_main_image;size:255" json:"local_main_image"` // 本地主图路径
	Images         string         `gorm:"column:images;type:json" json:"images"`                    // JSON 格式存储图片URL列表
	LocalImages    string         `gorm:"column:local_images;type:json" json:"local_images"`        // JSON 格式存储本地图片路径列表
	Detail         string         `gorm:"column:detail;type:text" json:"detail"`
	Price          float64        `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	OriginalPrice  *float64       `gorm:"column:original_price;type:decimal(10,2)" json:"original_price"`
	Stock          int            `gorm:"column:stock;default:0" json:"stock"`
	Sales          int            `gorm:"column:sales;default:0" json:"sales"`
	Status         int8           `gorm:"column:status;default:1" json:"status"`       // 0-下架, 1-上架, 2-待审核
	IsHot          int8           `gorm:"column:is_hot;default:0;index" json:"is_hot"` // 0-否, 1-是
	Sort           int            `gorm:"column:sort;default:0" json:"sort"`
	CreatedAt      time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "product"
}

// Category 商品类目模型
type Category struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	ParentID    uint64    `gorm:"column:parent_id;default:0;index" json:"parent_id"`
	Name        string    `gorm:"column:name;not null;size:100" json:"name"`
	Level       int8      `gorm:"column:level;not null" json:"level"`
	Sort        int       `gorm:"column:sort;default:0" json:"sort"`
	Icon        string    `gorm:"column:icon;size:255" json:"icon"`                // 图标URL
	IconLocal   string    `gorm:"column:icon_local;size:255" json:"icon_local"`    // 本地图标路径
	Image       string    `gorm:"column:image;size:255" json:"image"`              // 图片URL
	ImageLocal  string    `gorm:"column:image_local;size:255" json:"image_local"`  // 本地图片路径
	Description string    `gorm:"column:description;type:text" json:"description"` // 类目描述
	Status      int8      `gorm:"column:status;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "category"
}

// Sku SKU模型
type Sku struct {
	ID            uint64         `gorm:"primaryKey;column:id" json:"id"`
	ProductID     uint64         `gorm:"column:product_id;not null;index" json:"product_id"`
	SkuCode       string         `gorm:"column:sku_code;uniqueIndex;not null;size:50" json:"sku_code"`
	Name          string         `gorm:"column:name;not null;size:200" json:"name"`
	Specs         string         `gorm:"column:specs;type:json;not null" json:"specs"` // JSON 格式存储规格
	Price         float64        `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	OriginalPrice *float64       `gorm:"column:original_price;type:decimal(10,2)" json:"original_price"`
	Stock         int            `gorm:"column:stock;default:0" json:"stock"`
	Image         string         `gorm:"column:image;size:255" json:"image"`
	Weight        *float64       `gorm:"column:weight;type:decimal(8,2)" json:"weight"`
	Volume        *float64       `gorm:"column:volume;type:decimal(8,2)" json:"volume"`
	Status        int8           `gorm:"column:status;default:1" json:"status"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Sku) TableName() string {
	return "sku"
}

// Banner Banner模型 横幅广告
type Banner struct {
	ID          uint64     `gorm:"primaryKey;column:id" json:"id"`
	Title       string     `gorm:"column:title;size:200" json:"title"`
	Description string     `gorm:"column:description;size:500" json:"description"`
	Image       string     `gorm:"column:image;not null;size:255" json:"image"`    // 封面图片URL
	ImageLocal  string     `gorm:"column:image_local;size:255" json:"image_local"` // 本地图片路径
	Link        string     `gorm:"column:link;size:500" json:"link"`               // 跳转链接
	LinkType    int8       `gorm:"column:link_type;default:1" json:"link_type"`    // 1-商品详情, 2-分类页面, 3-外部链接, 4-无链接
	Sort        int        `gorm:"column:sort;default:0;index" json:"sort"`        // 排序值
	Status      int8       `gorm:"column:status;default:1;index" json:"status"`    // 0-禁用, 1-启用
	StartTime   *time.Time `gorm:"column:start_time;index" json:"start_time"`      // 开始时间
	EndTime     *time.Time `gorm:"column:end_time;index" json:"end_time"`          // 结束时间
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Banner) TableName() string {
	return "banner"
}
