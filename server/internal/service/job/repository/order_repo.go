package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// 订单状态 / 类型常量。
// 这里刻意用本地常量而不是 import order-service 的 model：定时任务是独立服务，
// 跨服务直接依赖对方内部 model 会把两个服务耦死。
const (
	OrderStatusCancelled = 0 // 已取消
	OrderStatusPending   = 1 // 待支付

	OrderTypeSeckill = 2 // 秒杀订单

	operatorTypeSystem = 3 // 操作人类型：系统
)

// CancelledOrderItem 被取消订单中的商品项
type CancelledOrderItem struct {
	SkuID    uint64
	Quantity int
}

// CancelledOrder 被取消的订单（含商品项）
// 上层拿它去做 Redis 补偿：秒杀订单需要把闸门库存和用户防重标记还回去。
type CancelledOrder struct {
	ID        uint64
	OrderNo   string
	UserID    uint64
	OrderType int8
	Items     []CancelledOrderItem
}

// OrderRepository 订单仓库接口（用于定时任务）
type OrderRepository interface {
	// CancelExpiredOrders 取消超时未支付的订单，并在同一事务内回补真源库存。
	// 返回被取消的订单明细（供上层补偿 Redis 秒杀配额）。
	CancelExpiredOrders(ctx context.Context, timeoutMinutes, limit int) ([]*CancelledOrder, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CancelExpiredOrders 取消超时订单。
//
// 修复前的实现有三个问题：
//  1. WHERE status = 0 —— 待支付是 1、0 是已取消，条件写反，一条都匹配不到；
//  2. SET status = 6 —— 6 是已退款，取消应该是 0；
//  3. 只改状态，没有回补 sku.stock / inventory，也没有释放 Redis 秒杀配额。
//
// 现在的实现：
//   - 用 status = 1（待支付）筛选，SET status = 0（已取消）；
//   - 行锁 + 条件更新（WHERE status = 待支付），避免和支付回调并发改同一单；
//   - 同一事务内回补 sku.stock 与 inventory；
//   - 写 order_log 流水（操作人=系统）；
//   - 返回明细给上层，由上层释放 Redis 秒杀配额。
func (r *orderRepository) CancelExpiredOrders(ctx context.Context, timeoutMinutes, limit int) ([]*CancelledOrder, error) {
	if timeoutMinutes <= 0 {
		timeoutMinutes = 30
	}
	if limit <= 0 {
		limit = 200
	}

	var cancelled []*CancelledOrder

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type pendingRow struct {
			ID        uint64
			OrderNo   string
			UserID    uint64
			OrderType int8
		}

		// 1. 捞出超时未支付的订单。FOR UPDATE 锁住这些行，
		//    防止与「支付成功回调」并发修改同一笔订单。
		var rows []pendingRow
		if err := tx.Raw(`
			SELECT id, order_no, user_id, order_type
			FROM orders
			WHERE status = ? AND created_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
			ORDER BY id ASC
			LIMIT ?
			FOR UPDATE
		`, OrderStatusPending, timeoutMinutes, limit).Scan(&rows).Error; err != nil {
			return fmt.Errorf("查询超时订单失败: %w", err)
		}
		if len(rows) == 0 {
			return nil
		}

		for _, o := range rows {
			// 2. 条件更新：仍然要求当前是「待支付」。
			//    如果期间用户刚好支付成功（status 已变成 2），RowsAffected 会是 0，直接跳过。
			res := tx.Exec(
				"UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ? AND status = ?",
				OrderStatusCancelled, o.ID, OrderStatusPending)
			if res.Error != nil {
				return fmt.Errorf("取消订单 %s 失败: %w", o.OrderNo, res.Error)
			}
			if res.RowsAffected == 0 {
				continue
			}

			// 3. 查订单项，逐项回补真源库存
			var items []CancelledOrderItem
			if err := tx.Raw(
				"SELECT sku_id, quantity FROM order_item WHERE order_id = ?", o.ID).
				Scan(&items).Error; err != nil {
				return fmt.Errorf("查询订单项失败: %w", err)
			}

			for _, it := range items {
				if it.Quantity <= 0 {
					continue
				}
				if err := tx.Exec(
					"UPDATE sku SET stock = stock + ? WHERE id = ?",
					it.Quantity, it.SkuID).Error; err != nil {
					return fmt.Errorf("回补 sku 库存失败: %w", err)
				}
				// sold_stock 用 GREATEST 兜底，避免减成负数
				if err := tx.Exec(
					`UPDATE inventory SET available_stock = available_stock + ?, sold_stock = GREATEST(sold_stock - ?, 0) WHERE sku_id = ?`,
					it.Quantity, it.Quantity, it.SkuID).Error; err != nil {
					return fmt.Errorf("回补 inventory 失败: %w", err)
				}
			}

			// 4. 记录订单日志（日志失败不影响主流程）
			_ = tx.Exec(`
				INSERT INTO order_log (order_id, order_no, operator_type, action, before_status, after_status, remark, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
			`, o.ID, o.OrderNo, operatorTypeSystem, "auto_cancel",
				OrderStatusPending, OrderStatusCancelled, "超时未支付，系统自动取消并回补库存").Error

			cancelled = append(cancelled, &CancelledOrder{
				ID:        o.ID,
				OrderNo:   o.OrderNo,
				UserID:    o.UserID,
				OrderType: o.OrderType,
				Items:     items,
			})
		}
		return nil
	})

	return cancelled, err
}
