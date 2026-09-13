package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PaymentStatus 支付状态
const (
	PaymentStatusPending  int8 = 0 // 待支付
	PaymentStatusSuccess  int8 = 1 // 支付成功
	PaymentStatusFailed   int8 = 2 // 支付失败
	PaymentStatusRefunded int8 = 3 // 已退款
)

// PaymentMethod 支付方式
const (
	PaymentMethodWechat   int8 = 1 // 微信
	PaymentMethodAlipay   int8 = 2 // 支付宝
	PaymentMethodUnionPay int8 = 3 // 银联
	PaymentMethodBalance  int8 = 4 // 余额
)

// Payment 支付单模型
type Payment struct {
	ID                 uint64     `gorm:"primaryKey;column:id" json:"id"`
	PaymentNo          string     `gorm:"column:payment_no;uniqueIndex;not null;size:32" json:"payment_no"`
	OrderID            uint64     `gorm:"column:order_id;not null;index" json:"order_id"`
	OrderNo            string     `gorm:"column:order_no;not null;index;size:32" json:"order_no"`
	UserID             uint64     `gorm:"column:user_id;not null;index" json:"user_id"`
	Amount             float64    `gorm:"column:amount;type:decimal(10,2);not null" json:"amount"`
	PaymentMethod      int8       `gorm:"column:payment_method;not null" json:"payment_method"` // 1-微信, 2-支付宝, 3-银联
	Status             int8       `gorm:"column:status;not null;index" json:"status"`           // 0-待支付, 1-支付成功, 2-支付失败, 3-已退款
	ThirdPartyNo       *string    `gorm:"column:third_party_no;index;size:100" json:"third_party_no"`
	ThirdPartyResponse JSONData   `gorm:"column:third_party_response;type:json" json:"third_party_response"`
	PaidAt             *time.Time `gorm:"column:paid_at" json:"paid_at"`
	ExpireAt           *time.Time `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payment"
}

// PaymentLog 支付流水模型
type PaymentLog struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	PaymentID    uint64    `gorm:"column:payment_id;not null;index" json:"payment_id"`
	PaymentNo    string    `gorm:"column:payment_no;not null;index;size:32" json:"payment_no"`
	Action       string    `gorm:"column:action;not null;size:50" json:"action"` // create, pay, refund, cancel
	Amount       float64   `gorm:"column:amount;type:decimal(10,2);not null" json:"amount"`
	BeforeStatus *int8     `gorm:"column:before_status" json:"before_status"`
	AfterStatus  *int8     `gorm:"column:after_status" json:"after_status"`
	RequestData  JSONData  `gorm:"column:request_data;type:json" json:"request_data"`
	ResponseData JSONData  `gorm:"column:response_data;type:json" json:"response_data"`
	Remark       string    `gorm:"column:remark;size:500" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at;index" json:"created_at"`
}

// TableName 指定表名
func (PaymentLog) TableName() string {
	return "payment_log"
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
