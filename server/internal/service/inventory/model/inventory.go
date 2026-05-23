package model

import (
	"time"
)

// Inventory 库存模型
// 用于存储每个 SKU 的库存数量信息，包括总库存、可用库存、锁定库存和销售库存等
type Inventory struct {
	// ID 主键，自增
	ID uint64 `gorm:"primaryKey;column:id" json:"id"`

	// SkuID 商品规格ID，唯一索引，不能为空
	// 用于关联商品系统中的 SKU 信息
	SkuID uint64 `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`

	// TotalStock 总库存数量
	// 表示该 SKU 当前仓库中的物理库存总量
	TotalStock int `gorm:"column:total_stock;default:0;not null" json:"total_stock"`

	// AvailableStock 可用库存数量
	// 表示当前可被订单使用的库存 = TotalStock - LockedStock
	AvailableStock int `gorm:"column:available_stock;default:0;not null" json:"available_stock"`

	// LockedStock 锁定库存数量
	// 表示已被订单预占但尚未完成支付的库存，防止超卖
	LockedStock int `gorm:"column:locked_stock;default:0;not null" json:"locked_stock"`

	// SoldStock 已售库存数量
	// 表示已完成支付并扣减的库存总量，用于销售统计
	SoldStock int `gorm:"column:sold_stock;default:0;not null" json:"sold_stock"`

	// LowStockThreshold 低库存预警阈值
	// 当可用库存低于此值时，可触发补货提醒（默认值：10）
	LowStockThreshold int `gorm:"column:low_stock_threshold;default:10" json:"low_stock_threshold"`

	// CreatedAt 记录创建时间
	// 由 GORM 自动管理
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	// UpdatedAt 记录最后更新时间
	// 由 GORM 自动管理
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定数据库表名为 "inventory"
func (Inventory) TableName() string {
	return "inventory"
}

// InventoryLog 库存流水模型
// 用于记录库存变更的完整轨迹，支持对账、审计和问题排查
type InventoryLog struct {
	// ID 主键，自增
	ID uint64 `gorm:"primaryKey;column:id" json:"id"`

	// SkuID 商品规格ID，建立普通索引
	// 用于快速查询某个商品的所有库存变动记录
	SkuID uint64 `gorm:"column:sku_id;not null;index" json:"sku_id"`

	// OrderID 关联的订单ID，可为空（如手动调库时无订单）
	// 建立索引便于按订单追溯库存变更
	OrderID *uint64 `gorm:"column:order_id;index" json:"order_id"`

	// Type 库存变动类型
	// 1-入库（采购/退货入库）, 2-出库（发货/损耗）, 3-锁定（下单预占）
	// 4-解锁（取消订单释放）, 5-扣减（支付成功正式扣减）, 6-回退（退款/取消后恢复）
	Type int8 `gorm:"column:type;not null;index" json:"type"`

	// Quantity 变动数量
	// 正值表示增加，负值表示减少，具体含义由 Type 字段决定
	Quantity int `gorm:"column:quantity;not null" json:"quantity"`

	// BeforeStock 变动前的库存数量（快照）
	// 用于对账和状态回溯
	BeforeStock int `gorm:"column:before_stock;not null" json:"before_stock"`

	// AfterStock 变动后的库存数量（快照）
	// 与 BeforeStock + Quantity 应保持一致，用于数据校验
	AfterStock int `gorm:"column:after_stock;not null" json:"after_stock"`

	// Remark 备注信息
	// 记录变动原因、操作人、来源系统等辅助信息（最大255字符）
	Remark string `gorm:"column:remark;size:255" json:"remark"`

	// CreatedAt 流水记录创建时间
	// 建立索引便于按时间范围查询
	CreatedAt time.Time `gorm:"column:created_at;index" json:"created_at"`
}

// TableName 指定数据库表名为 "inventory_log"
func (InventoryLog) TableName() string {
	return "inventory_log"
}
