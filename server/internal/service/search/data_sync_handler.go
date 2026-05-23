package search

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/outbox"
	"github.com/zhian9/GoForge/server/internal/service/search/repository"
)

func (s *ServiceContext) handleDataSyncMessage(ctx context.Context, msg *mq.Message) error {
	if msg == nil || s == nil {
		return nil
	}
	if s.ESClient == nil {
		return nil
	}

	productID := extractProductID(msg)
	if productID == 0 {
		return nil
	}

	switch msg.EventType {
	case outbox.EventProductUpserted:
		if s.SnapshotRepo == nil {
			return nil
		}
		doc, err := s.SnapshotRepo.BuildProductDocument(ctx, productID)
		if err != nil {
			return err
		}
		docID := strconv.FormatUint(productID, 10)
		if err := s.ESClient.IndexDocument(ctx, repository.ProductIndexName, docID, doc); err != nil {
			return err
		}
		logx.Infof("ES upsert ok: index=%s product_id=%d", repository.ProductIndexName, productID)
		return nil

	case outbox.EventProductDeleted:
		docID := strconv.FormatUint(productID, 10)
		if err := s.ESClient.DeleteDocument(ctx, repository.ProductIndexName, docID); err != nil {
			return err
		}
		logx.Infof("ES delete ok: index=%s product_id=%d", repository.ProductIndexName, productID)
		return nil

	default:
		return nil
	}
}

func extractProductID(msg *mq.Message) uint64 {
	// 优先从 payload.product_id 取
	if msg.Data != nil {
		if p, ok := msg.Data["payload"]; ok {
			switch v := p.(type) {
			case map[string]interface{}:
				if id := toUint64(v["product_id"]); id > 0 {
					return id
				}
			}
		}
		// 其次从 aggregate_id（字符串）取
		if id := toUint64(msg.Data["aggregate_id"]); id > 0 {
			return id
		}
	}
	return 0
}

func toUint64(v interface{}) uint64 {
	switch x := v.(type) {
	case nil:
		return 0
	case uint64:
		return x
	case int64:
		if x < 0 {
			return 0
		}
		return uint64(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint64(x)
	case float64:
		if x < 0 {
			return 0
		}
		return uint64(x)
	case string:
		if x == "" {
			return 0
		}
		n, _ := strconv.ParseUint(x, 10, 64)
		return n
	default:
		if s := fmt.Sprintf("%v", x); s != "" {
			n, _ := strconv.ParseUint(s, 10, 64)
			return n
		}
		return 0
	}
}
