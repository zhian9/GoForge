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
	// Delete 删除消息（只能删自己的）
	Delete(ctx context.Context, userID, messageID uint64) error
	// CreateBatch 批量写入消息（群发公告用）
	CreateBatch(ctx context.Context, messages []*model.Message) error
	// ListActiveUserIDs 查询所有启用且未删除的用户 ID（群发公告用）
	ListActiveUserIDs(ctx context.Context) ([]uint64, error)
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
	res := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": &now,
		})
	if res.Error != nil {
		return res.Error
	}
	// 没有匹配到行（比如 message_id 传错/越权）要显式报错，
	// 否则接口会返回"成功"但消息其实没被标记
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// Delete 删除消息（带 user_id 条件，避免越权删除别人的消息）
func (r *messageRepository) Delete(ctx context.Context, userID, messageID uint64) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", messageID, userID).
		Delete(&model.Message{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateBatch 批量写入消息（群发时按批插入，避免一次塞进太多 SQL）
func (r *messageRepository) CreateBatch(ctx context.Context, messages []*model.Message) error {
	if len(messages) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(messages, 500).Error
}

// ListActiveUserIDs 查询所有启用且未删除的用户 ID
func (r *messageRepository) ListActiveUserIDs(ctx context.Context) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("user").
		Where("status = ? AND deleted_at IS NULL", 1).
		Pluck("id", &ids).Error
	return ids, err
}
