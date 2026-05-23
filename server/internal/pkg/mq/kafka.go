package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

// Config kafka配置
type Config struct {
	Brokers       []string `json:"required"`
	ProducerAsync bool     `json:"default=true"`
	Version       string   `json:"default=2.8.0"`
	ConsumerGroup string   `json:"optional"`
}

// Message 消息结构
type Message struct {
	Version   string                 `json:"version,omitempty"`
	MessageID string                 `json:"message_id,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	EventType string                 `json:"event_type,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewMessage 创建消息
func NewMessage(eventType string, data map[string]interface{}) *Message {
	return &Message{
		Version:   "1.0",
		MessageID: generateMessageID(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Data:      data,
	}
}

// ToJSON 转换为JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// Producer kafka生产者
type Producer struct {
	producer sarama.AsyncProducer
	config   *Config
}

// NewProducer 创建kafka生产者
func NewProducer(cfg *Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy

	//解析版本
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		version = sarama.V2_8_0_0
	}
	config.Version = version

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建kafka生产者失败: %w", err)
	}

	p := &Producer{
		producer: producer,
		config:   cfg,
	}

	// 启动错误处理协程
	go p.handleErrors()
	go p.handleSuccesses()

	return p, nil
}

// Publish 发布消息
func (p *Producer) Publish(ctx context.Context, topic string, message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.MessageID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(message.EventType),
			},
			{
				Key:   []byte("message_id"),
				Value: []byte(message.MessageID),
			},
		},
		Timestamp: time.Now(),
	}

	select {
	case p.producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PublishWithKey 发布消息（指定分区key）
func (p *Producer) PublishWithKey(ctx context.Context, topic string, partitionKey string, message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(partitionKey),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(message.EventType),
			},
			{
				Key:   []byte("message_id"),
				Value: []byte(message.MessageID),
			},
		},
		Timestamp: time.Now(),
	}

	select {
	case p.producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.producer.Close()
}

// handleErrors 处理错误
func (p *Producer) handleErrors() {
	for err := range p.producer.Errors() {
		logx.Errorf("Kafka生产者错误: topic=%s, error=%v", err.Msg.Topic, err.Err)
	}
}

// handleSuccesses 处理成功消息
func (p *Producer) handleSuccesses() {
	for msg := range p.producer.Successes() {
		logx.Infof("Kafka消息发送成功: topic=%s, partition=%d, offset=%d",
			msg.Topic, msg.Partition, msg.Offset)
	}
}

// Consumer Kafka消费者
type Consumer struct {
	consumer sarama.ConsumerGroup
	config   *Config
	handlers map[string]MessageHandler
}

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, message *Message) error

// NewConsumer 创建Kafka消费者
func NewConsumer(cfg *Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version, _ = sarama.ParseKafkaVersion(cfg.Version)

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		config:   cfg,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// RegisterHandler 注册消息处理器
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

// Start 启动消费者
func (c *Consumer) Start(ctx context.Context, topics []string) error {
	handler := &consumerGroupHandler{
		handlers: c.handlers,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.consumer.Consume(ctx, topics, handler)
			if err != nil {
				logx.Errorf("消费消息失败: %v", err)
				time.Sleep(time.Second)
			}
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// consumerGroupHandler 消费者组处理器
type consumerGroupHandler struct {
	handlers map[string]MessageHandler
}

// Setup 会话开始
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 会话结束
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg := <-claim.Messages():
			if msg == nil {
				continue
			}

			// 解析消息
			var message Message
			if err := json.Unmarshal(msg.Value, &message); err != nil {
				logx.Errorf("解析消息失败: topic=%s, error=%v", msg.Topic, err)
				session.MarkMessage(msg, "")
				continue
			}

			// 查找处理器
			handler, ok := h.handlers[msg.Topic]
			if !ok {
				logx.Infof("未找到消息处理器: topic=%s", msg.Topic)
				session.MarkMessage(msg, "")
				continue
			}

			// 处理消息
			if err := handler(session.Context(), &message); err != nil {
				logx.Errorf("处理消息失败: topic=%s, message_id=%s, error=%v",
					msg.Topic, message.MessageID, err)
				// 这里可以实现重试逻辑或死信队列
			} else {
				session.MarkMessage(msg, "")
			}
		}
	}
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
