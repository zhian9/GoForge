package model

import (
	"gorm.io/gorm"
	"time"
)

// Cart 购物车模型
type Cart struct {
	ID          uint64         `gorm:"column:id; primaryKey" json:"id"`
	UserID      uint64         `gorm:"column:user_id;not null; index" json:"user_id"`
	SkuID       uint64         `gorm:"column:sku_id;not null" json:"sku_id"`
	ProductName  string         `gorm:"column:product_name;size:200" json:"product_name"`
	Price        float64        `gorm:"column:price;type:decimal(10,2)" json:"price"`
	ProductImage string         `gorm:"column:product_image;size:500" json:"product_image"`
	Quantity    int            `gorm:"column:quantity;default:1;not null" json:"quantity"`
	IsSelected  int8           `gorm:"column:is_selected;default:1" json:"is_selected"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

// TableName 指定表名
func (Cart) TableName() string {
	return "cart"
}
