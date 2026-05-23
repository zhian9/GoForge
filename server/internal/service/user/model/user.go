package model

import (
	"gorm.io/gorm"
	"time"
)

// User 用户模型
type User struct {
	ID       uint64     `gorm:"primaryKey;column:id" json:"id"`
	Username string     `gorm:"column:username;uniqueIndex;not null;size:50" json:"username"`
	Nickname string     `gorm:"column:nickname;size:50" json:"nickname"`
	Phone    string     `gorm:"column:phone;uniqueIndex;size:20" json:"phone"`
	Email    string     `gorm:"column:email;uniqueIndex;size:100" json:"email"`
	Avatar   string     `gorm:"column:avatar;size:255" json:"avatar"`
	Gender   int8       `gorm:"column:gender;default:0" json:"gender"`
	Birthday *time.Time `gorm:"column:birthday;type:date" json:"birthday"`
	//状态
	Status int8 `gorm:"column:status;default:1" json:"status"`
	//会员等级
	MemberLevel int8 `gorm:"column:member_level;default:0" json:"member_level"`
	//积分
	Points    int            `gorm:"column:points;default:0" json:"points"`
	IsAdmin   int8           `gorm:"column:is_admin;default:0" json:"is_admin"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// Address 用户地址模型
type Address struct {
	ID            uint64 `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint64 `gorm:"column:user_id;not null;index" json:"user_id"`
	ReceiverName  string `gorm:"column:receiver_name;not null;size:50" json:"receiver_name"`
	ReceiverPhone string `gorm:"column:receiver_phone;not null;size:20" json:"receiver_phone"`
	Province      string `gorm:"column:province;not null;size:50" json:"province"`
	City          string `gorm:"column:city;not null;size:50" json:"city"`
	//区县
	District string `gorm:"column:district;not null;size:50" json:"district"`
	//详细地址
	Detail string `gorm:"column:detail;not null;size:200" json:"detail"`
	//邮政编码
	PostalCode string `gorm:"column:postal_code;size:10" json:"postal_code"`
	//是否为默认地址
	IsDefault int8           `gorm:"column:is_default;default:0" json:"is_default"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Address) TableName() string {
	return "address"
}

// Credential 用户凭证模型
type Credential struct {
	ID              uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	CredentialType  int8      `gorm:"column:credential_type;not null" json:"credential_type"`
	CredentialKey   string    `gorm:"column:credential_key;not null;size:100" json:"credential_key"`
	CredentialValue string    `gorm:"column:credential_value;size:255" json:"-"`
	Extra           string    `gorm:"column:extra;type:json" json:"extra"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Credential) TableName() string {
	return "credential"
}
