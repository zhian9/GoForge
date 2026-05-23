package model

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID        uint64     `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64     `gorm:"column:user_id;not null;index" json:"user_id"`
	Type      int8       `gorm:"column:type;not null;index" json:"type"` // 1-系统通知, 2-订单消息, 3-营销消息, 4-物流消息
	Title     string     `gorm:"column:title;not null;size:200" json:"title"`
	Content   string     `gorm:"column:content;type:text;not null" json:"content"`
	Link      *string    `gorm:"column:link;size:500" json:"link"`
	IsRead    int8       `gorm:"column:is_read;default:0;index" json:"is_read"`
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt time.Time  `gorm:"column:created_at;index" json:"created_at"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "message"
}
