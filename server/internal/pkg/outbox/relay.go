package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/mq"
)

type RelayConfig struct {
	PollInterval time.Duration
	BatchSize    int
	MaxRetry     int
}

type Relay struct {
	repo     *Repo
	producer *mq.Producer
	cfg      RelayConfig
}

func NewRelay(repo *Repo, producer *mq.Producer, cfg RelayConfig) *Relay {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.MaxRetry <= 0 {
		cfg.MaxRetry = 10
	}
	return &Relay{repo: repo, producer: producer, cfg: cfg}
}

func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Relay) tick(ctx context.Context) {
	if r.repo == nil || r.producer == nil {
		return
	}

	events, err := r.repo.ListPending(ctx, r.cfg.BatchSize)
	if err != nil {
		logx.Errorf("outbox relay: list pending failed: %v", err)
		return
	}
	if len(events) == 0 {
		return
	}

	for _, evt := range events {
		if evt == nil {
			continue
		}

		data := map[string]interface{}{
			"aggregate_type": evt.AggregateType,
			"aggregate_id":   evt.AggregateID,
			"outbox_id":      evt.ID,
		}

		// payload 是 JSON，尽量解开为 map 便于下游消费
		if evt.Payload != nil && *evt.Payload != "" {
			var payloadAny any
			if err := json.Unmarshal([]byte(*evt.Payload), &payloadAny); err == nil {
				data["payload"] = payloadAny
			} else {
				data["payload"] = *evt.Payload
			}
		}

		msg := mq.NewMessage(evt.EventType, data)
		if err := r.producer.Publish(ctx, mq.TopicDataSync, msg); err != nil {
			_ = r.repo.MarkFailed(ctx, evt.ID, err.Error(), r.cfg.MaxRetry)
			continue
		}
		_ = r.repo.MarkSent(ctx, evt.ID)
	}
}
