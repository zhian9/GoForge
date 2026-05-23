package service

import (
	"context"
	"time"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/review/model"
	"github.com/zhian9/GoForge/server/internal/service/review/repository"
)

// ReviewLogic 评价业务逻辑
type ReviewLogic struct {
	reviewRepo      repository.ReviewRepository
	reviewReplyRepo repository.ReviewReplyRepository
}

// NewReviewLogic 创建评价业务逻辑
func NewReviewLogic(
	reviewRepo repository.ReviewRepository,
	reviewReplyRepo repository.ReviewReplyRepository,
) *ReviewLogic {
	return &ReviewLogic{
		reviewRepo:      reviewRepo,
		reviewReplyRepo: reviewReplyRepo,
	}
}

// CreateReviewRequest 创建评价请求
type CreateReviewRequest struct {
	UserID      uint64
	OrderID     uint64
	OrderItemID uint64
	ProductID   uint64
	SkuID       uint64
	Rating      int8
	Content     string
	Images      []string
	Videos      []string
}

// CreateReviewResponse 创建评价响应
type CreateReviewResponse struct {
	Review *model.Review
}

// CreateReview 创建评价
func (l *ReviewLogic) CreateReview(ctx context.Context, req *CreateReviewRequest) (*CreateReviewResponse, error) {
	// 验证评分范围
	if req.Rating < 1 || req.Rating > 5 {
		return nil, apperrors.NewInvalidParamError("评分必须在1-5之间")
	}

	review := &model.Review{
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		OrderItemID: req.OrderItemID,
		ProductID:   req.ProductID,
		SkuID:       req.SkuID,
		Rating:      req.Rating,
		Content:     req.Content,
		Images:      req.Images,
		Videos:      req.Videos,
		Status:      1, // 显示
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 构建MongoDB评价详情（包含图片、视频）
	reviewDetail := map[string]interface{}{
		"user_id":    req.UserID,
		"product_id": req.ProductID,
		"sku_id":     req.SkuID,
		"order_id":   req.OrderID,
		"rating":     req.Rating,
		"content":    req.Content,
		"images":     req.Images,
		"videos":     req.Videos,
		"status":     1,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	err := l.reviewRepo.Create(ctx, review, reviewDetail)
	if err != nil {
		return nil, apperrors.NewInternalError("创建评价失败")
	}

	return &CreateReviewResponse{
		Review: review,
	}, nil
}

// GetProductReviewsRequest 获取商品评价列表请求
type GetProductReviewsRequest struct {
	ProductID uint64
	Page      int
	PageSize  int
	Rating    int8
}

// GetProductReviewsResponse 获取商品评价列表响应
type GetProductReviewsResponse struct {
	Reviews []*model.Review
	Total   int64
}

// GetProductReviews 获取商品评价列表
func (l *ReviewLogic) GetProductReviews(ctx context.Context, req *GetProductReviewsRequest) (*GetProductReviewsResponse, error) {
	reviews, total, err := l.reviewRepo.GetByProductID(ctx, req.ProductID, req.Page, req.PageSize, req.Rating)
	if err != nil {
		return nil, apperrors.NewInternalError("获取评价列表失败")
	}

	return &GetProductReviewsResponse{
		Reviews: reviews,
		Total:   total,
	}, nil
}

// GetReviewRequest 获取评价详情请求
type GetReviewRequest struct {
	ReviewID uint64
}

// GetReviewResponse 获取评价详情响应
type GetReviewResponse struct {
	Review *model.Review
}

// GetReview 获取评价详情
func (l *ReviewLogic) GetReview(ctx context.Context, req *GetReviewRequest) (*GetReviewResponse, error) {
	review, err := l.reviewRepo.GetByID(ctx, req.ReviewID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("评价不存在")
	}

	return &GetReviewResponse{
		Review: review,
	}, nil
}

// ReplyReviewRequest 回复评价请求
type ReplyReviewRequest struct {
	ReviewID uint64
	UserID   uint64
	Content  string
	ParentID uint64
}

// ReplyReview 回复评价
func (l *ReviewLogic) ReplyReview(ctx context.Context, req *ReplyReviewRequest) error {
	// 验证评价是否存在
	_, err := l.reviewRepo.GetByID(ctx, req.ReviewID)
	if err != nil {
		return apperrors.NewNotFoundError("评价不存在")
	}

	reply := &model.ReviewReply{
		ReviewID:  req.ReviewID,
		UserID:    req.UserID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		Status:    1,
		CreatedAt: time.Now(),
	}

	err = l.reviewReplyRepo.Create(ctx, reply)
	if err != nil {
		return apperrors.NewInternalError("回复评价失败")
	}

	// 如果是商家回复，更新评价的回复信息
	if req.ParentID == 0 {
		now := time.Now()
		review, _ := l.reviewRepo.GetByID(ctx, req.ReviewID)
		if review != nil {
			review.ReplyContent = &req.Content
			review.ReplyTime = &now
			_ = l.reviewRepo.Update(ctx, review)
		}
	}

	return nil
}

// GetReviewStatsRequest 获取评价统计请求
type GetReviewStatsRequest struct {
	ProductID uint64
}

// GetReviewStatsResponse 获取评价统计响应
type GetReviewStatsResponse struct {
	Stats *model.ReviewStats
}

// GetReviewStats 获取评价统计
func (l *ReviewLogic) GetReviewStats(ctx context.Context, req *GetReviewStatsRequest) (*GetReviewStatsResponse, error) {
	stats, err := l.reviewRepo.GetStats(ctx, req.ProductID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取评价统计失败")
	}

	return &GetReviewStatsResponse{
		Stats: stats,
	}, nil
}
