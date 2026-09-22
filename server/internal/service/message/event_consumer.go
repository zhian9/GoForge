package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/message/service"
)

// 消息类型（与 message 表注释保持一致）
const (
	msgTypeSystem int8 = 1 // 系统通知
	msgTypeOrder  int8 = 2 // 订单消息
)

// EventConsumer 订阅业务事件，自动生成站内信。
//
// 为什么用事件驱动而不是让业务服务直接调 gRPC：
// 下单、支付这些主链路本来就在往 Kafka 发事件（order.created / payment-service.success ...），
// 消息服务订阅即可，既不用给下单链路增加一个同步依赖，
// 也不会因为消息库抖动而影响下单；
// 同时事件里已经带了 user_id（支付事件本次也补上了），不需要跨服务查库。
type EventConsumer struct {
	logic    *service.MessageLogic
	consumer *mq.Consumer
}

// NewEventConsumer 创建事件消费者并注册各主题的处理器
func NewEventConsumer(cfg *mq.Config, logic *service.MessageLogic) (*EventConsumer, error) {
	consumer, err := mq.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	c := &EventConsumer{logic: logic, consumer: consumer}
	consumer.RegisterHandler(mq.TopicOrderCreated, c.onOrderCreated)
	consumer.RegisterHandler(mq.TopicOrderCancelled, c.onOrderCancelled)
	consumer.RegisterHandler(mq.TopicPaymentSuccess, c.onPaymentSuccess)
	consumer.RegisterHandler(mq.TopicPaymentRefunded, c.onPaymentRefunded)

	return c, nil
}

// Topics 订阅的主题列表
func (c *EventConsumer) Topics() []string {
	return []string{
		mq.TopicOrderCreated,
		mq.TopicOrderCancelled,
		mq.TopicPaymentSuccess,
		mq.TopicPaymentRefunded,
	}
}

// Start 阻塞消费（内部对 Consume 失败做了重试）
func (c *EventConsumer) Start(ctx context.Context) error {
	return c.consumer.Start(ctx, c.Topics())
}

// Close 关闭消费者
func (c *EventConsumer) Close() error {
	return c.consumer.Close()
}

// ---- 各事件 -> 站内信 ----

func (c *EventConsumer) onOrderCreated(ctx context.Context, m *mq.Message) error {
	userID := dataUint(m.Data, "user_id")
	if userID == 0 {
		return skipEvent(m, "缺少 user_id")
	}
	orderID := dataUint(m.Data, "order_id")
	orderNo := dataString(m.Data, "order_no")
	amount := dataString(m.Data, "total_amount")

	return c.send(ctx, userID, msgTypeOrder, "订单提交成功",
		fmt.Sprintf("订单 %s（¥%s）已创建，请在 30 分钟内完成支付，超时订单会自动取消。", orderNo, amount),
		orderLink(orderID))
}

func (c *EventConsumer) onOrderCancelled(ctx context.Context, m *mq.Message) error {
	userID := dataUint(m.Data, "user_id")
	if userID == 0 {
		return skipEvent(m, "缺少 user_id")
	}
	orderID := dataUint(m.Data, "order_id")
	orderNo := dataString(m.Data, "order_no")

	return c.send(ctx, userID, msgTypeOrder, "订单已取消",
		fmt.Sprintf("订单 %s 已取消，占用的库存已释放；若使用了优惠券，券已退回你的账户。", orderNo),
		orderLink(orderID))
}

func (c *EventConsumer) onPaymentSuccess(ctx context.Context, m *mq.Message) error {
	userID := dataUint(m.Data, "user_id")
	if userID == 0 {
		return skipEvent(m, "缺少 user_id")
	}
	orderID := dataUint(m.Data, "order_id")
	orderNo := dataString(m.Data, "order_no")
	amount := dataString(m.Data, "amount")

	return c.send(ctx, userID, msgTypeOrder, "支付成功",
		fmt.Sprintf("订单 %s 已支付 ¥%s，我们会尽快为你发货。", orderNo, amount),
		orderLink(orderID))
}

func (c *EventConsumer) onPaymentRefunded(ctx context.Context, m *mq.Message) error {
	userID := dataUint(m.Data, "user_id")
	if userID == 0 {
		return skipEvent(m, "缺少 user_id")
	}
	orderID := dataUint(m.Data, "order_id")
	orderNo := dataString(m.Data, "order_no")
	refundAmount := dataString(m.Data, "refund_amount")

	return c.send(ctx, userID, msgTypeOrder, "退款已受理",
		fmt.Sprintf("订单 %s 的退款 ¥%s 已受理，将原路退回；该订单占用的优惠券已退回账户。", orderNo, refundAmount),
		orderLink(orderID))
}

func (c *EventConsumer) send(ctx context.Context, userID uint64, msgType int8, title, content, link string) error {
	req := &service.SendMessageRequest{
		UserID:  userID,
		Type:    msgType,
		Title:   title,
		Content: content,
		Link:    link,
	}
	if err := c.logic.SendMessage(ctx, req); err != nil {
		// 返回错误让 Kafka 重试（消费者组失败会重新投递）
		return fmt.Errorf("写入站内信失败: %w", err)
	}

	logx.Infof("站内信已生成: user_id=%d type=%d title=%s", userID, msgType, title)
	return nil
}

// skipEvent 事件本身不完整（例如缺少 user_id）时只记日志并跳过。
// 这类消息重试也不会成功，一直返回 error 会让 Kafka 反复投递同一条消息、卡住该分区；
// 只有暂时性失败（例如写库失败）才返回 error 让 Kafka 重试。
func skipEvent(m *mq.Message, reason string) error {
	logx.Errorf("跳过无法处理的 %s 事件(%s): %s", m.EventType, m.MessageID, reason)
	return nil
}

// ---- 事件字段解析（JSON 反序列化后数字是 float64，gRPC 侧可能是字符串） ----

func dataUint(data map[string]interface{}, key string) uint64 {
	if data == nil {
		return 0
	}
	switch v := data[key].(type) {
	case float64:
		return uint64(v)
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	case json.Number:
		n, _ := v.Int64()
		return uint64(n)
	case string:
		n, _ := strconv.ParseUint(v, 10, 64)
		return n
	}
	return 0
}

func dataString(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	switch v := data[key].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case json.Number:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func orderLink(orderID uint64) string {
	if orderID == 0 {
		return ""
	}
	return fmt.Sprintf("/orders/%d", orderID)
}
