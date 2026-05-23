package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Coupon 优惠券模型
type Coupon struct {
	ID             uint64    `gorm:"primaryKey;column:id" json:"id"`
	Name           string    `gorm:"column:name;not null;size:100" json:"name"`
	Type           int8      `gorm:"column:type;not null" json:"type"`                                        // 1-满减券, 2-折扣券, 3-免运费券
	DiscountType   int8      `gorm:"column:discount_type;not null" json:"discount_type"`                      // 1-固定金额, 2-百分比折扣
	DiscountValue  float64   `gorm:"column:discount_value;type:decimal(10,2);not null" json:"discount_value"` //折扣值
	MinAmount      float64   `gorm:"column:min_amount;type:decimal(10,2);default:0" json:"min_amount"`        //最低使用金额
	MaxDiscount    *float64  `gorm:"column:max_discount;type:decimal(10,2)" json:"max_discount"`              //最大折扣金额
	TotalCount     int       `gorm:"column:total_count;default:-1" json:"total_count"`                        // 总数量 :-1表示不限
	UsedCount      int       `gorm:"column:used_count;default:0" json:"used_count"`                           //已使用数量
	PerUserLimit   int       `gorm:"column:per_user_limit;default:1" json:"per_user_limit"`                   //每用户限领数量
	ValidStartTime time.Time `gorm:"column:valid_start_time;not null;index" json:"valid_start_time"`          //有效开始时间
	ValidEndTime   time.Time `gorm:"column:valid_end_time;not null;index" json:"valid_end_time"`              // 有效结束时间
	Status         int8      `gorm:"column:status;default:1;index" json:"status"`                             //状态 0-禁用，1启用
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Coupon) TableName() string {
	return "coupon"
}

// UserCoupon 用户优惠券模型
type UserCoupon struct {
	ID        uint64     `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64     `gorm:"column:user_id;not null;index" json:"user_id"`
	CouponID  uint64     `gorm:"column:coupon_id;not null;index" json:"coupon_id"`
	Status    int8       `gorm:"column:status;default:0;index" json:"status"` // 0-未使用, 1-已使用, 2-已过期
	OrderID   *uint64    `gorm:"column:order_id;index" json:"order_id"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`                    //使用时间
	ExpireAt  time.Time  `gorm:"column:expire_at;not null;index" json:"expire_at"` //过期时间
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (UserCoupon) TableName() string {
	return "user_coupon"
}

// Promotion 促销活动模型
type Promotion struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;not null;size:100" json:"name"`
	Type        int8      `gorm:"column:type;not null;index" json:"type"`     // 1-满减, 2-折扣, 3-秒杀, 4-拼团
	Rule        JSONData  `gorm:"column:rule;type:json;not null" json:"rule"` //活动规则
	ProductIDs  JSONArray `gorm:"column:product_ids;type:json" json:"product_ids"`
	CategoryIDs JSONArray `gorm:"column:category_ids;type:json" json:"category_ids"`
	StartTime   time.Time `gorm:"column:start_time;not null;index" json:"start_time"`
	EndTime     time.Time `gorm:"column:end_time;not null;index" json:"end_time"`
	Status      int8      `gorm:"column:status;default:1;index" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Promotion) TableName() string {
	return "promotion"
}

// Points 积分模型
type Points struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	Total     int64     `gorm:"column:total;default:0;not null" json:"total"`         //总积分
	Used      int64     `gorm:"column:used;default:0;not null" json:"used"`           //已用积分
	Available int64     `gorm:"column:available;default:0;not null" json:"available"` //可用积分
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Points) TableName() string {
	return "points"
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

// JSONArray JSON数组类型
type JSONArray []uint64

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
