package review

import (
	"context"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/review/v1"
	"github.com/zhian9/GoForge/server/internal/service/review/model"
	"github.com/zhian9/GoForge/server/internal/service/review/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReviewService 实现 gRPC 服务接口
type ReviewService struct {
	v1.UnimplementedReviewServiceServer
	svcCtx *ServiceContext
	logic  *service.ReviewLogic
}

// NewReviewService 创建评价服务
func NewReviewService(svcCtx *ServiceContext) *ReviewService {
	logic := service.NewReviewLogic(
		svcCtx.ReviewRepo,
		svcCtx.ReviewReplyRepo,
	)

	return &ReviewService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// CreateReview 创建评价
func (s *ReviewService) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.CreateReviewResponse, error) {
	createReq := &service.CreateReviewRequest{
		UserID:      uint64(req.UserId),
		OrderID:     uint64(req.OrderId),
		OrderItemID: uint64(req.OrderItemId),
		ProductID:   uint64(req.ProductId),
		SkuID:       uint64(req.SkuId),
		Rating:      int8(req.Rating),
		Content:     req.Content,
		Images:      req.Images,
		Videos:      req.Videos,
	}

	resp, err := s.logic.CreateReview(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.CreateReviewResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertReviewToProto(resp.Review),
	}, nil
}

// GetProductReviews 获取商品评价列表
func (s *ReviewService) GetProductReviews(ctx context.Context, req *v1.GetProductReviewsRequest) (*v1.GetProductReviewsResponse, error) {
	getReq := &service.GetProductReviewsRequest{
		ProductID: uint64(req.ProductId),
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		Rating:    int8(req.Rating),
	}

	resp, err := s.logic.GetProductReviews(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	reviews := make([]*v1.Review, 0, len(resp.Reviews))
	for _, r := range resp.Reviews {
		reviews = append(reviews, convertReviewToProto(r))
	}

	return &v1.GetProductReviewsResponse{
		Code:    0,
		Message: "成功",
		Data:    reviews,
		Total:   int32(resp.Total),
	}, nil
}

// GetReview 获取评价详情
func (s *ReviewService) GetReview(ctx context.Context, req *v1.GetReviewRequest) (*v1.GetReviewResponse, error) {
	getReq := &service.GetReviewRequest{
		ReviewID: uint64(req.ReviewId),
	}

	resp, err := s.logic.GetReview(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetReviewResponse{
		Code:    0,
		Message: "成功",
		Data:    convertReviewToProto(resp.Review),
	}, nil
}

// ReplyReview 回复评价
func (s *ReviewService) ReplyReview(ctx context.Context, req *v1.ReplyReviewRequest) (*v1.ReplyReviewResponse, error) {
	replyReq := &service.ReplyReviewRequest{
		ReviewID: uint64(req.ReviewId),
		UserID:   uint64(req.UserId),
		Content:  req.Content,
		ParentID: uint64(req.ParentId),
	}

	err := s.logic.ReplyReview(ctx, replyReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ReplyReviewResponse{
		Code:    0,
		Message: "回复成功",
	}, nil
}

// GetReviewStats 获取评价统计
func (s *ReviewService) GetReviewStats(ctx context.Context, req *v1.GetReviewStatsRequest) (*v1.GetReviewStatsResponse, error) {
	getReq := &service.GetReviewStatsRequest{
		ProductID: uint64(req.ProductId),
	}

	resp, err := s.logic.GetReviewStats(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetReviewStatsResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.ReviewStats{
			TotalCount:    resp.Stats.TotalCount,
			Rating_5Count: resp.Stats.Rating5Count,
			Rating_4Count: resp.Stats.Rating4Count,
			Rating_3Count: resp.Stats.Rating3Count,
			Rating_2Count: resp.Stats.Rating2Count,
			Rating_1Count: resp.Stats.Rating1Count,
			AverageRating: resp.Stats.AverageRating,
		},
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

// convertReviewToProto 转换评价模型为 Protobuf 消息
func convertReviewToProto(r *model.Review) *v1.Review {
	if r == nil {
		return nil
	}

	var replyContent string
	if r.ReplyContent != nil {
		replyContent = *r.ReplyContent
	}

	return &v1.Review{
		Id:           int64(r.ID),
		UserId:       int64(r.UserID),
		OrderId:      int64(r.OrderID),
		ProductId:    int64(r.ProductID),
		SkuId:        int64(r.SkuID),
		Rating:       int32(r.Rating),
		Content:      r.Content,
		Images:       r.Images,
		Videos:       r.Videos,
		Status:       int32(r.Status),
		ReplyContent: replyContent,
		CreatedAt:    formatTime(&r.CreatedAt),
	}
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
