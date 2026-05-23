package outbox

import "time"

// EventStatus Outbox 事件状态
const (
	StatusPending int8 = 0
	StatusSent    int8 = 1
	StatusFailed  int8 = 2
)

// 常用事件类型（供 search-service 消费）
const (
	EventProductUpserted = "product.upserted"
	EventProductDeleted  = "product.deleted"
)

// Event Outbox 事件（同事务写入，异步投递到 Kafka）
type Event struct {
	ID            uint64     `gorm:"primaryKey;column:id" json:"id"`
	AggregateType string     `gorm:"column:aggregate_type;type:varchar(32);not null" json:"aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id;type:varchar(64);not null" json:"aggregate_id"`
	EventType     string     `gorm:"column:event_type;type:varchar(64);not null" json:"event_type"`
	Payload       *string    `gorm:"column:payload;type:json" json:"payload"`
	Status        int8       `gorm:"column:status;not null;default:0" json:"status"`
	RetryCount    int        `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	LastError     *string    `gorm:"column:last_error;type:varchar(255)" json:"last_error"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	SentAt        *time.Time `gorm:"column:sent_at" json:"sent_at"`
}

func (Event) TableName() string { return "outbox_event" }
