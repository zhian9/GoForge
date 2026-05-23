package outbox

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// CreateInTx 在同一事务内创建 Outbox 事件
func (r *Repo) CreateInTx(ctx context.Context, tx *gorm.DB, evt *Event) error {
	if tx == nil {
		tx = r.db
	}
	now := time.Now()
	if evt.CreatedAt.IsZero() {
		evt.CreatedAt = now
	}
	if evt.Status == 0 {
		evt.Status = StatusPending
	}
	return tx.WithContext(ctx).Create(evt).Error
}

// ListPending 查询待投递事件
func (r *Repo) ListPending(ctx context.Context, limit int) ([]*Event, error) {
	if limit <= 0 {
		limit = 100
	}
	var events []*Event
	err := r.db.WithContext(ctx).
		Where("status = ?", StatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// MarkSent 标记已投递
func (r *Repo) MarkSent(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&Event{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":  StatusSent,
			"sent_at": &now,
		}).Error
}

// MarkFailed 标记失败（并递增重试次数）
func (r *Repo) MarkFailed(ctx context.Context, id uint64, errMsg string, maxRetry int) error {
	updates := map[string]any{
		"retry_count": gorm.Expr("retry_count + 1"),
		"last_error":  errMsg,
	}
	if maxRetry > 0 {
		updates["status"] = gorm.Expr("CASE WHEN retry_count + 1 >= ? THEN ? ELSE status END", maxRetry, StatusFailed)
	}
	return r.db.WithContext(ctx).
		Model(&Event{}).
		Where("id = ?", id).
		Updates(updates).Error
}
