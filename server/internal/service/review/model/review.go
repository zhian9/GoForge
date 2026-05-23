package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Review 评价模型
type Review struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID      uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	OrderID     uint64    `gorm:"column:order_id;not null;index" json:"order_id"`
	OrderItemID uint64    `gorm:"column:order_item_id;not null" json:"order_item_id"`
	ProductID   uint64    `gorm:"column:product_id;not null;index" json:"product_id"`
	SkuID       uint64    `gorm:"column:sku_id;not null;index" json:"sku_id"`
	Rating      int8      `gorm:"column:rating;not null;index" json:"rating"` // 1-5星
	Content     string    `gorm:"column:content;type:text" json:"content"`
	Images      JSONArray `gorm:"column:images;type:json" json:"images"`
	Videos      JSONArray `gorm:"column:videos;type:json" json:"videos"`
	Status      int8      `gorm:"column:status;default:1;index" json:"status"`
	// ReplyContent 商家回复
	ReplyContent *string    `gorm:"column:reply_content;type:text" json:"reply_content"`
	ReplyTime    *time.Time `gorm:"column:reply_time" json:"reply_time"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Review) TableName() string {
	return "review"
}

// ReviewReply 评价回复模型
type ReviewReply struct {
	ID       uint64 `gorm:"primaryKey;column:id" json:"id"`
	ReviewID uint64 `gorm:"column:review_id;not null;index" json:"review_id"`
	UserID   uint64 `gorm:"column:user_id;not null;index" json:"user_id"`
	Content  string `gorm:"column:content;type:text;not null" json:"content"`
	// ParentID 父回复ID 0-直接回复
	ParentID uint64 `gorm:"column:parent_id;default:0;index" json:"parent_id"`
	// Status 状态 (0-隐藏，1-显示)
	Status    int8      `gorm:"column:status;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (ReviewReply) TableName() string {
	return "review_reply"
}

// JSONArray JSON数组类型
type JSONArray []string

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
