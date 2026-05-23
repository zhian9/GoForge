package service

import (
	"context"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/recommend/repository"
)

// RecommendLogic 推荐业务逻辑
type RecommendLogic struct {
	recommendRepo repository.RecommendRepository
}

// NewRecommendLogic 创建推荐业务逻辑
func NewRecommendLogic(recommendRepo repository.RecommendRepository) *RecommendLogic {
	return &RecommendLogic{
		recommendRepo: recommendRepo,
	}
}

// GetPersonalizedRecommendRequest 个性化推荐请求
type GetPersonalizedRecommendRequest struct {
	UserID uint64
	Limit  int
}

// RecommendProduct 推荐商品
type RecommendProduct struct {
	ProductID int64
	Name      string
	MainImage string
	Price     float64
	Score     float64
	Reason    string
}

// GetPersonalizedRecommendResponse 个性化推荐响应
type GetPersonalizedRecommendResponse struct {
	Products []*RecommendProduct
}

// GetPersonalizedRecommend 获取个性化推荐
func (l *RecommendLogic) GetPersonalizedRecommend(ctx context.Context, req *GetPersonalizedRecommendRequest) (*GetPersonalizedRecommendResponse, error) {
	results, err := l.recommendRepo.GetPersonalizedRecommend(ctx, req.UserID, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取个性化推荐失败")
	}

	products := make([]*RecommendProduct, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		productID, _ := r["product_id"].(int64)
		name, _ := r["name"].(string)
		mainImage, _ := r["main_image"].(string)
		price, _ := r["price"].(float64)
		score, _ := r["score"].(float64)
		reason, _ := r["reason"].(string)

		products = append(products, &RecommendProduct{
			ProductID: productID,
			Name:      name,
			MainImage: mainImage,
			Price:     price,
			Score:     score,
			Reason:    reason,
		})
	}

	return &GetPersonalizedRecommendResponse{
		Products: products,
	}, nil
}

// GetSimilarProductsRequest 相似商品推荐请求
type GetSimilarProductsRequest struct {
	ProductID uint64
	Limit     int
}

// GetSimilarProductsResponse 相似商品推荐响应
type GetSimilarProductsResponse struct {
	Products []*RecommendProduct
}

// GetSimilarProducts 获取相似商品
func (l *RecommendLogic) GetSimilarProducts(ctx context.Context, req *GetSimilarProductsRequest) (*GetSimilarProductsResponse, error) {
	results, err := l.recommendRepo.GetSimilarProducts(ctx, req.ProductID, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取相似商品失败")
	}

	products := make([]*RecommendProduct, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		productID, _ := r["product_id"].(int64)
		name, _ := r["name"].(string)
		mainImage, _ := r["main_image"].(string)
		price, _ := r["price"].(float64)
		score, _ := r["score"].(float64)

		products = append(products, &RecommendProduct{
			ProductID: productID,
			Name:      name,
			MainImage: mainImage,
			Price:     price,
			Score:     score,
			Reason:    "相似商品",
		})
	}

	return &GetSimilarProductsResponse{
		Products: products,
	}, nil
}

// GetHotProductsRequest 热门推荐请求
type GetHotProductsRequest struct {
	Limit      int
	CategoryID uint64
}

// GetHotProductsResponse 热门推荐响应
type GetHotProductsResponse struct {
	Products []*RecommendProduct
}

// GetHotProducts 获取热门商品
func (l *RecommendLogic) GetHotProducts(ctx context.Context, req *GetHotProductsRequest) (*GetHotProductsResponse, error) {
	results, err := l.recommendRepo.GetHotProducts(ctx, req.CategoryID, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取热门商品失败")
	}

	products := make([]*RecommendProduct, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		productID, _ := r["product_id"].(int64)
		name, _ := r["name"].(string)
		mainImage, _ := r["main_image"].(string)
		price, _ := r["price"].(float64)
		score, _ := r["score"].(float64)

		products = append(products, &RecommendProduct{
			ProductID: productID,
			Name:      name,
			MainImage: mainImage,
			Price:     price,
			Score:     score,
			Reason:    "热门商品",
		})
	}

	return &GetHotProductsResponse{
		Products: products,
	}, nil
}

// GetRealtimeRecommendRequest 实时推荐请求
type GetRealtimeRecommendRequest struct {
	UserID uint64
	Limit  int
}

// GetRealtimeRecommendResponse 实时推荐响应
type GetRealtimeRecommendResponse struct {
	Products []*RecommendProduct
}

// GetRealtimeRecommend 获取实时推荐
func (l *RecommendLogic) GetRealtimeRecommend(ctx context.Context, req *GetRealtimeRecommendRequest) (*GetRealtimeRecommendResponse, error) {
	results, err := l.recommendRepo.GetRealtimeRecommend(ctx, req.UserID, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取实时推荐失败")
	}

	products := make([]*RecommendProduct, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		productID, _ := r["product_id"].(int64)
		name, _ := r["name"].(string)
		mainImage, _ := r["main_image"].(string)
		price, _ := r["price"].(float64)
		score, _ := r["score"].(float64)

		products = append(products, &RecommendProduct{
			ProductID: productID,
			Name:      name,
			MainImage: mainImage,
			Price:     price,
			Score:     score,
			Reason:    "实时推荐",
		})
	}

	return &GetRealtimeRecommendResponse{
		Products: products,
	}, nil
}
