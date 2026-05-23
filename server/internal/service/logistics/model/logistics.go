package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Logistics 物流模型
type Logistics struct {
	ID               uint64     `gorm:"primaryKey;column:id" json:"id"`
	OrderID          uint64     `gorm:"column:order_id;not null;index" json:"order_id"`
	OrderNo          string     `gorm:"column:order_no;not null;index;size:32" json:"order_no"`
	LogisticsCompany string     `gorm:"column:logistics_company;not null;size:50" json:"logistics_company"`
	LogisticsNo      string     `gorm:"column:logistics_no;not null;index;size:50" json:"logistics_no"`
	ReceiverName     string     `gorm:"column:receiver_name;not null;size:50" json:"receiver_name"`
	ReceiverPhone    string     `gorm:"column:receiver_phone;not null;size:20" json:"receiver_phone"`
	ReceiverAddress  string     `gorm:"column:receiver_address;not null;size:500" json:"receiver_address"`
	SenderName       *string    `gorm:"column:sender_name;size:50" json:"sender_name"`
	SenderPhone      *string    `gorm:"column:sender_phone;size:20" json:"sender_phone"`
	SenderAddress    *string    `gorm:"column:sender_address;size:500" json:"sender_address"`
	Status           int8       `gorm:"column:status;default:0;index" json:"status"` // 0-待发货, 1-已发货, 2-运输中, 3-已送达, 4-异常
	CurrentLocation  *string    `gorm:"column:current_location;size:200" json:"current_location"`
	TrackingInfo     JSONData   `gorm:"column:tracking_info;type:json" json:"tracking_info"`
	ShippedAt        *time.Time `gorm:"column:shipped_at" json:"shipped_at"`
	DeliveredAt      *time.Time `gorm:"column:delivered_at" json:"delivered_at"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Logistics) TableName() string {
	return "logistics"
}

// JSONData JSON数据类型
type JSONData map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONData) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONData) Scan(value interface{}) error {
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
