package recommend

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"strconv"

	v1 "github.com/zhian9/GoForge/server/api/recommend/v1"
	"github.com/zhian9/GoForge/server/internal/service/recommend/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecommendService 实现 gRPC 服务接口
type RecommendService struct {
	v1.UnimplementedRecommendServiceServer
	svcCtx *ServiceContext
	logic  *service.RecommendLogic
}

// NewRecommendService 创建推荐服务
func NewRecommendService(svcCtx *ServiceContext) *RecommendService {
	logic := service.NewRecommendLogic(svcCtx.RecommendRepo)

	return &RecommendService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// GetPersonalizedRecommend 获取个性化推荐
func (s *RecommendService) GetPersonalizedRecommend(ctx context.Context, req *v1.GetPersonalizedRecommendRequest) (*v1.GetPersonalizedRecommendResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetPersonalizedRecommendRequest{
		UserID: userID,
		Limit:  int(req.Limit),
	}

	resp, err := s.logic.GetPersonalizedRecommend(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	products := make([]*v1.RecommendProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, &v1.RecommendProduct{
			ProductId: p.ProductID,
			Name:      p.Name,
			MainImage: p.MainImage,
			Price:     strconv.FormatFloat(p.Price, 'f', 2, 64),
			Score:     p.Score,
			Reason:    p.Reason,
		})
	}

	return &v1.GetPersonalizedRecommendResponse{
		Code:    0,
		Message: "成功",
		Data:    products,
	}, nil
}

// GetSimilarProducts 获取相似商品
func (s *RecommendService) GetSimilarProducts(ctx context.Context, req *v1.GetSimilarProductsRequest) (*v1.GetSimilarProductsResponse, error) {
	getReq := &service.GetSimilarProductsRequest{
		ProductID: uint64(req.ProductId),
		Limit:     int(req.Limit),
	}

	resp, err := s.logic.GetSimilarProducts(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	products := make([]*v1.RecommendProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, &v1.RecommendProduct{
			ProductId: p.ProductID,
			Name:      p.Name,
			MainImage: p.MainImage,
			Price:     strconv.FormatFloat(p.Price, 'f', 2, 64),
			Score:     p.Score,
			Reason:    p.Reason,
		})
	}

	return &v1.GetSimilarProductsResponse{
		Code:    0,
		Message: "成功",
		Data:    products,
	}, nil
}

// GetHotProducts 获取热门商品
func (s *RecommendService) GetHotProducts(ctx context.Context, req *v1.GetHotProductsRequest) (*v1.GetHotProductsResponse, error) {
	getReq := &service.GetHotProductsRequest{
		Limit:      int(req.Limit),
		CategoryID: uint64(req.CategoryId),
	}

	resp, err := s.logic.GetHotProducts(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	products := make([]*v1.RecommendProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, &v1.RecommendProduct{
			ProductId: p.ProductID,
			Name:      p.Name,
			MainImage: p.MainImage,
			Price:     strconv.FormatFloat(p.Price, 'f', 2, 64),
			Score:     p.Score,
			Reason:    p.Reason,
		})
	}

	return &v1.GetHotProductsResponse{
		Code:    0,
		Message: "成功",
		Data:    products,
	}, nil
}

// GetRealtimeRecommend 获取实时推荐
func (s *RecommendService) GetRealtimeRecommend(ctx context.Context, req *v1.GetRealtimeRecommendRequest) (*v1.GetRealtimeRecommendResponse, error) {
	// 只认 token 里的身份；请求参数里的 user_id 一律忽略
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
	}
	getReq := &service.GetRealtimeRecommendRequest{
		UserID: userID,
		Limit:  int(req.Limit),
	}

	resp, err := s.logic.GetRealtimeRecommend(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	products := make([]*v1.RecommendProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, &v1.RecommendProduct{
			ProductId: p.ProductID,
			Name:      p.Name,
			MainImage: p.MainImage,
			Price:     strconv.FormatFloat(p.Price, 'f', 2, 64),
			Score:     p.Score,
			Reason:    p.Reason,
		})
	}

	return &v1.GetRealtimeRecommendResponse{
		Code:    0,
		Message: "成功",
		Data:    products,
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}
