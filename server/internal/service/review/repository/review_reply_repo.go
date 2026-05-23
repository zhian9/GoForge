package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/service/review/model"
	"gorm.io/gorm"
)

// ReviewReplyRepository 评价回复仓库接口
type ReviewReplyRepository interface {
	// Create 创建回复
	Create(ctx context.Context, reply *model.ReviewReply) error
	// GetByReviewID 根据评价ID获取回复列表
	GetByReviewID(ctx context.Context, reviewID uint64) ([]*model.ReviewReply, error)
}

type reviewReplyRepository struct {
	db *gorm.DB
}

// NewReviewReplyRepository 创建评价回复仓库
func NewReviewReplyRepository(db *gorm.DB) ReviewReplyRepository {
	return &reviewReplyRepository{db: db}
}

// Create 创建回复
func (r *reviewReplyRepository) Create(ctx context.Context, reply *model.ReviewReply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

// GetByReviewID 根据评价ID获取回复列表
func (r *reviewReplyRepository) GetByReviewID(ctx context.Context, reviewID uint64) ([]*model.ReviewReply, error) {
	var replies []*model.ReviewReply
	err := r.db.WithContext(ctx).Where("review_id = ? AND status = ?", reviewID, 1).
		Order("created_at ASC").Find(&replies).Error
	return replies, err
}
