package repository

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/mongodb"
	"github.com/zhian9/GoForge/server/internal/service/review/model"
	"gorm.io/gorm"
)

// ReviewRepository 评价仓库接口
type ReviewRepository interface {
	// Create 创建评价（同时写入MySQL和MongoDB）
	Create(ctx context.Context, review *model.Review, reviewDetail map[string]interface{}) error
	// GetByID 根据ID获取
	GetByID(ctx context.Context, id uint64) (*model.Review, error)
	// GetByProductID 根据商品ID获取评价列表
	GetByProductID(ctx context.Context, productID uint64, page, pageSize int, rating int8) ([]*model.Review, int64, error)
	// Update 更新评价
	Update(ctx context.Context, review *model.Review) error
	// GetStats 获取评价统计
	GetStats(ctx context.Context, productID uint64) (*model.ReviewStats, error)
	// GetReviewDetail 从MongoDB获取评价详情（包含图片、视频）
	GetReviewDetail(ctx context.Context, reviewID uint64) (map[string]interface{}, error)
}

type reviewRepository struct {
	db      *gorm.DB
	mongoDB *mongodb.Client
}

// NewReviewRepository 创建评价仓库
func NewReviewRepository(db *gorm.DB, mongoDB *mongodb.Client) ReviewRepository {
	return &reviewRepository{
		db:      db,
		mongoDB: mongoDB,
	}
}

// Create 创建评价（同时写入MySQL和MongoDB）
func (r *reviewRepository) Create(ctx context.Context, review *model.Review, reviewDetail map[string]interface{}) error {
	// 写入MySQL（基础信息）
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return err
	}

	// 写入MongoDB（完整信息，包含图片、视频）
	if r.mongoDB != nil && reviewDetail != nil {
		reviewDetail["review_id"] = review.ID
		collection := r.mongoDB.Collection("reviews")
		_, err := collection.InsertOne(ctx, reviewDetail)
		if err != nil {
			// MongoDB写入失败不影响主流程，记录日志即可
			logx.Errorf("写入MongoDB失败: %v", err)
		}
	}

	return nil
}

// GetByID 根据ID获取
func (r *reviewRepository) GetByID(ctx context.Context, id uint64) (*model.Review, error) {
	var review model.Review
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// GetByProductID 根据商品ID获取评价列表
func (r *reviewRepository) GetByProductID(ctx context.Context, productID uint64, page, pageSize int, rating int8) ([]*model.Review, int64, error) {
	var reviews []*model.Review
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Review{}).Where("product_id = ? AND status = ?", productID, 1)

	if rating > 0 {
		query = query.Where("rating = ?", rating)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error
	return reviews, total, err
}

// Update 更新评价
func (r *reviewRepository) Update(ctx context.Context, review *model.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

// GetReviewDetail 从MongoDB获取评价详情
func (r *reviewRepository) GetReviewDetail(ctx context.Context, reviewID uint64) (map[string]interface{}, error) {
	if r.mongoDB == nil {
		return nil, nil
	}

	collection := r.mongoDB.Collection("reviews")
	var result map[string]interface{}
	err := collection.FindOne(ctx, map[string]interface{}{"review_id": reviewID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetStats 获取评价统计
func (r *reviewRepository) GetStats(ctx context.Context, productID uint64) (*model.ReviewStats, error) {
	var stats model.ReviewStats

	// 总评价数
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ?", productID, 1).
		Count(&stats.TotalCount)

	// 各星级评价数
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, 1, 5).
		Count(&stats.Rating5Count)
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, 1, 4).
		Count(&stats.Rating4Count)
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, 1, 3).
		Count(&stats.Rating3Count)
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, 1, 2).
		Count(&stats.Rating2Count)
	r.db.WithContext(ctx).Model(&model.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, 1, 1).
		Count(&stats.Rating1Count)

	// 计算平均评分
	if stats.TotalCount > 0 {
		totalRating := stats.Rating5Count*5 + stats.Rating4Count*4 + stats.Rating3Count*3 + stats.Rating2Count*2 + stats.Rating1Count*1
		stats.AverageRating = float64(totalRating) / float64(stats.TotalCount)
	}

	return &stats, nil
}
