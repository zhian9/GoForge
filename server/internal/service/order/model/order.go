package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// OrderStatus 订单状态
const (
	OrderStatusCancelled int8 = 0 // 已取消
	OrderStatusPending   int8 = 1 // 待支付
	OrderStatusPaid      int8 = 2 // 待发货
	OrderStatusShipped   int8 = 3 // 待收货
	OrderStatusCompleted int8 = 4 // 已完成
	OrderStatusRefunded  int8 = 5 // 已退款
)

// OrderType 订单类型
const (
	OrderTypeNormal   int8 = 1 // 普通订单
	OrderTypeSeckill  int8 = 2 // 秒杀订单
	OrderTypeGroupBuy int8 = 3 // 拼团订单
)

// PaymentMethod 支付方式
const (
	PaymentMethodWeChat   int8 = 1 // 微信
	PaymentMethodAlipay   int8 = 2 // 支付宝
	PaymentMethodUnionPay int8 = 3 // 银联
)

// Order 订单表
type Order struct {
	ID              uint64      `gorm:"primaryKey;column:id" json:"id"`
	OrderNo         string      `gorm:"column:order_no;type:varchar(32);uniqueIndex;not null" json:"order_no"`
	UserID          uint64      `gorm:"column:user_id;type:bigint unsigned;not null;index" json:"user_id"`
	OrderType       int8        `gorm:"column:order_type;type:tinyint;default:1" json:"order_type"`
	Status          int8        `gorm:"column:status;type:tinyint;not null;index" json:"status"`
	TotalAmount     float64     `gorm:"column:total_amount;type:decimal(10,2);not null" json:"total_amount"`
	PayAmount       float64     `gorm:"column:pay_amount;type:decimal(10,2);not null" json:"pay_amount"`
	DiscountAmount  float64     `gorm:"column:discount_amount;type:decimal(10,2);default:0" json:"discount_amount"`
	FreightAmount   float64     `gorm:"column:freight_amount;type:decimal(10,2);default:0" json:"freight_amount"`
	ReceiverName    string      `gorm:"column:receiver_name;type:varchar(50);not null" json:"receiver_name"`
	ReceiverPhone   string      `gorm:"column:receiver_phone;type:varchar(20);not null" json:"receiver_phone"`
	ReceiverAddress string      `gorm:"column:receiver_address;type:varchar(500);not null" json:"receiver_address"`
	PaymentMethod   *int8       `gorm:"column:payment_method;type:tinyint" json:"payment_method"`
	PaymentTime     *time.Time  `gorm:"column:payment_time;type:datetime" json:"payment_time"`
	DeliveryTime    *time.Time  `gorm:"column:delivery_time;type:datetime" json:"delivery_time"`
	ReceiveTime     *time.Time  `gorm:"column:receive_time;type:datetime" json:"receive_time"`
	CancelTime      *time.Time  `gorm:"column:cancel_time;type:datetime" json:"cancel_time"`
	CancelReason    *string     `gorm:"column:cancel_reason;type:varchar(255)" json:"cancel_reason"`
	Remark          *string     `gorm:"column:remark;type:varchar(500)" json:"remark"`
	Items           []OrderItem `gorm:"foreignKey:OrderID;references:ID" json:"items"` // 关联订单项
	CreatedAt       time.Time   `gorm:"column:created_at;type:datetime;not null;index" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"column:updated_at;type:datetime;not null" json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	// 避免使用 MySQL 保留关键字，统一改为 orders
	return "orders"
}

// OrderItem 订单商品项表
type OrderItem struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	OrderID     uint64    `gorm:"column:order_id;type:bigint unsigned;not null;index" json:"order_id"`
	OrderNo     string    `gorm:"column:order_no;type:varchar(32);not null;index" json:"order_no"`
	ProductID   uint64    `gorm:"column:product_id;type:bigint unsigned;not null;index" json:"product_id"`
	ProductName string    `gorm:"column:product_name;type:varchar(200);not null" json:"product_name"`
	SkuID       uint64    `gorm:"column:sku_id;type:bigint unsigned;not null;index" json:"sku_id"`
	SkuCode     string    `gorm:"column:sku_code;type:varchar(50);not null" json:"sku_code"`
	SkuName     string    `gorm:"column:sku_name;type:varchar(200);not null" json:"sku_name"`
	SkuImage    *string   `gorm:"column:sku_image;type:varchar(255)" json:"sku_image"`
	SkuSpecs    JSONMap   `gorm:"column:sku_specs;type:json" json:"sku_specs"`
	Price       float64   `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	Quantity    int       `gorm:"column:quantity;type:int;not null" json:"quantity"`
	TotalAmount float64   `gorm:"column:total_amount;type:decimal(10,2);not null" json:"total_amount"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null" json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_item"
}

// OrderLog 订单操作日志表
type OrderLog struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	OrderID      uint64    `gorm:"column:order_id;type:bigint unsigned;not null;index" json:"order_id"`
	OrderNo      string    `gorm:"column:order_no;type:varchar(32);not null;index" json:"order_no"`
	OperatorType int8      `gorm:"column:operator_type;type:tinyint;not null" json:"operator_type"` // 1-用户, 2-系统, 3-管理员
	OperatorID   *uint64   `gorm:"column:operator_id;type:bigint unsigned" json:"operator_id"`
	Action       string    `gorm:"column:action;type:varchar(50);not null" json:"action"`
	BeforeStatus *int8     `gorm:"column:before_status;type:tinyint" json:"before_status"`
	AfterStatus  *int8     `gorm:"column:after_status;type:tinyint" json:"after_status"`
	Remark       *string   `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;not null;index" json:"created_at"`
}

// TableName 指定表名
func (OrderLog) TableName() string {
	return "order_log"
}

// JSONMap JSON映射类型
type JSONMap map[string]string

// Value 实现 driver.Valuer 接口
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}
