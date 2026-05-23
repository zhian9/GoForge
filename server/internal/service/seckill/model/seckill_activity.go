package model

import (
	"gorm.io/gorm"
	"time"
)

// SeckillActivity 秒杀活动
type SeckillActivity struct {
	ID           uint64         `gorm:"primaryKey;column:id" json:"id"`
	Name         string         `gorm:"column:name;not null;size:200" json:"name"`
	SkuID        uint64         `gorm:"column:sku_id;not null;index" json:"sku_id"`
	SeckillPrice float64        `gorm:"column:seckill_price;type:decimal(10,2);not null" json:"seckill_price"`
	Stock        int            `gorm:"column:stock;not null;default:0" json:"stock"` //初始库存
	StartTime    int64          `gorm:"column:start_time;not null;index"  json:"start_time"`
	EndTime      int64          `gorm:"column:end_time;not null;index" json:"end_time"`
	Status       int8           `gorm:"column:status;default:1;index" json:"status"` //可选:0-禁用,1-启用
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (SeckillActivity) TableName() string {
	return "seckill_activity"
}
