package repository

import (
	"context"
	"time"

	"github.com/zhian9/GoForge/server/internal/service/message/model"
	"gorm.io/gorm"
)

// MessageRepository 消息仓库接口
type MessageRepository interface {
	// Create 创建消息
	Create(ctx context.Context, message *model.Message) error
	// GetList 获取消息列表
	GetList(ctx context.Context, userID uint64, msgType int8, page, pageSize int) ([]*model.Message, int64, error)
	// MarkAsRead 标记已读
	MarkAsRead(ctx context.Context, userID, messageID uint64) error
	// BatchMarkAsRead 批量标记已读
	BatchMarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64) error
	// GetUnreadCount 获取未读数量
	GetUnreadCount(ctx context.Context, userID uint64) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓库
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetList 获取消息列表
func (r *messageRepository) GetList(ctx context.Context, userID uint64, msgType int8, page, pageSize int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Message{}).Where("user_id = ?", userID)
	if msgType > 0 {
		query = query.Where("type = ?", msgType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// MarkAsRead 标记已读
func (r *messageRepository) MarkAsRead(ctx context.Context, userID, messageID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": &now,
		}).Error
}

// BatchMarkAsRead 批量标记已读
func (r *messageRepository) BatchMarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id IN ? AND user_id = ?", messageIDs, userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": &now,
		}).Error
}

// GetUnreadCount 获取未读数量
func (r *messageRepository) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, 0).
		Count(&count).Error
	return count, err
}
