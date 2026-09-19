package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/order/repository"
)

// PaymentConsumer 支付结果消费者：监听支付服务发布的事件，联动订单状态。
type PaymentConsumer struct {
	orderRepo repository.OrderRepository
	db        *gorm.DB
	cache     *cache.CacheOperations
}

// NewPaymentConsumer 创建支付结果消费者
func NewPaymentConsumer(orderRepo repository.OrderRepository, db *gorm.DB, cacheOps *cache.CacheOperations) *PaymentConsumer {
	return &PaymentConsumer{orderRepo: orderRepo, db: db, cache: cacheOps}
}

// paymentSuccessMessage 支付成功事件
type paymentSuccessMessage struct {
	OrderID       uint64  `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	PaymentNo     string  `json:"payment_no"`
	PaymentMethod int8    `json:"payment_method"`
	Amount        float64 `json:"amount"`
}

// paymentRefundMessage 退款事件
type paymentRefundMessage struct {
	OrderID   uint64 `json:"order_id"`
	OrderNo   string `json:"order_no"`
	PaymentNo string `json:"payment_no"`
}

// ConsumePaymentSuccess 处理支付成功：订单 待支付 -> 待发货
func (c *PaymentConsumer) ConsumePaymentSuccess(ctx context.Context, message *mq.Message) error {
	var msg paymentSuccessMessage
	if err := decodeMessage(message, &msg); err != nil {
		return err
	}
	if msg.OrderID == 0 {
		return fmt.Errorf("支付成功消息缺少 order_id")
	}

	// 幂等：MarkPaid 内部仅在待支付状态下更新
	if err := c.orderRepo.MarkPaid(ctx, msg.OrderID, msg.PaymentMethod); err != nil {
		logx.Errorf("标记订单已支付失败 order_id=%d: %v", msg.OrderID, err)
		return err
	}
	logx.Infof("订单支付成功联动: order_id=%d, payment_no=%s", msg.OrderID, msg.PaymentNo)

	// 下单支付成功发放积分（1元 = 1积分）
	if order, err := c.orderRepo.GetByID(ctx, msg.OrderID); err == nil && order != nil && c.db != nil {
		points := int64(msg.Amount)
		if points > 0 {
			if err := c.db.WithContext(ctx).Exec("UPDATE user SET points = points + ? WHERE id = ?", points, order.UserID).Error; err != nil {
				logx.Errorf("下单发放积分失败 user_id=%d: %v", order.UserID, err)
			} else {
				logx.Infof("下单发放积分: user_id=%d, +%d 积分", order.UserID, points)
				// 清理用户信息缓存，让积分变化即时生效
				if c.cache != nil {
					_ = c.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixUserInfo, order.UserID))
				}
			}
		}
	}
	return nil
}

// ConsumePaymentRefund 处理退款：订单 -> 已退款，并回增库存
func (c *PaymentConsumer) ConsumePaymentRefund(ctx context.Context, message *mq.Message) error {
	var msg paymentRefundMessage
	if err := decodeMessage(message, &msg); err != nil {
		return err
	}
	if msg.OrderID == 0 {
		return fmt.Errorf("退款消息缺少 order_id")
	}

	// 幂等标记退款：仅首次命中才继续（重复消息直接跳过）
	affected, err := c.orderRepo.MarkRefunded(ctx, msg.OrderID)
	if err != nil {
		logx.Errorf("标记订单为已退款失败 order_id=%d: %v", msg.OrderID, err)
		return err
	}
	if !affected {
		return nil
	}
	logx.Infof("订单退款联动: order_id=%d, payment_no=%s", msg.OrderID, msg.PaymentNo)

	// 退款回增库存：释放下单时扣减的 sku.stock / inventory
	c.restockOrder(ctx, msg.OrderID)
	return nil
}

// restockOrder 退款后回增订单所有商品的库存
func (c *PaymentConsumer) restockOrder(ctx context.Context, orderID uint64) {
	if c.db == nil {
		return
	}
	var items []struct {
		SkuID    uint64 `gorm:"column:sku_id"`
		Quantity int    `gorm:"column:quantity"`
	}
	if err := c.db.WithContext(ctx).Table("order_item").Select("sku_id, quantity").Where("order_id = ?", orderID).Scan(&items).Error; err != nil {
		logx.Errorf("查询订单商品项失败 order_id=%d: %v", orderID, err)
		return
	}
	for _, it := range items {
		if err := restockStock(ctx, c.db, c.cache, it.SkuID, it.Quantity); err != nil {
			logx.Errorf("退款回增库存失败 sku_id=%d: %v", it.SkuID, err)
		}
	}
}

// decodeMessage 从 mq.Message.Data 解析出目标结构体
func decodeMessage(message *mq.Message, dest interface{}) error {
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("序列化消息数据失败: %w", err)
	}
	if err := json.Unmarshal(dataBytes, dest); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}
	return nil
}
