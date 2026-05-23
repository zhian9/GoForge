package service

import (
	"context"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/search/repository"
)

// SearchLogic 搜索业务逻辑
type SearchLogic struct {
	searchRepo repository.SearchRepository
}

// NewSearchLogic 创建搜索业务逻辑
func NewSearchLogic(searchRepo repository.SearchRepository) *SearchLogic {
	return &SearchLogic{
		searchRepo: searchRepo,
	}
}

// SearchProductsRequest 搜索商品请求
type SearchProductsRequest struct {
	Keyword    string
	Page       int
	PageSize   int
	CategoryID uint64
	SortBy     string
}

// ProductSearchResult 商品搜索结果
type ProductSearchResult struct {
	ProductID int64
	Name      string
	MainImage string
	Price     float64
	Sales     int
	Score     float64
}

// SearchProductsResponse 搜索商品响应
type SearchProductsResponse struct {
	Results []*ProductSearchResult
	Total   int64
}

// SearchProducts 搜索商品
func (l *SearchLogic) SearchProducts(ctx context.Context, req *SearchProductsRequest) (*SearchProductsResponse, error) {
	results, total, err := l.searchRepo.SearchProducts(ctx, req.Keyword, req.CategoryID, req.Page, req.PageSize, req.SortBy)
	if err != nil {
		return nil, apperrors.NewInternalError("搜索商品失败")
	}

	products := make([]*ProductSearchResult, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		productID, _ := r["product_id"].(int64)
		name, _ := r["name"].(string)
		mainImage, _ := r["main_image"].(string)
		price, _ := r["price"].(float64)
		sales, _ := r["sales"].(int)
		score, _ := r["score"].(float64)

		products = append(products, &ProductSearchResult{
			ProductID: productID,
			Name:      name,
			MainImage: mainImage,
			Price:     price,
			Sales:     sales,
			Score:     score,
		})
	}

	return &SearchProductsResponse{
		Results: products,
		Total:   total,
	}, nil
}

// GetSearchSuggestionsRequest 获取搜索建议请求
type GetSearchSuggestionsRequest struct {
	Keyword string
	Limit   int
}

// GetSearchSuggestionsResponse 获取搜索建议响应
type GetSearchSuggestionsResponse struct {
	Suggestions []string
}

// GetSearchSuggestions 获取搜索建议
func (l *SearchLogic) GetSearchSuggestions(ctx context.Context, req *GetSearchSuggestionsRequest) (*GetSearchSuggestionsResponse, error) {
	suggestions, err := l.searchRepo.GetSearchSuggestions(ctx, req.Keyword, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取搜索建议失败")
	}

	return &GetSearchSuggestionsResponse{
		Suggestions: suggestions,
	}, nil
}

// GetHotKeywordsRequest 获取搜索热词请求
type GetHotKeywordsRequest struct {
	Limit int
}

// GetHotKeywordsResponse 获取搜索热词响应
type GetHotKeywordsResponse struct {
	Keywords []string
}

// GetHotKeywords 获取搜索热词
func (l *SearchLogic) GetHotKeywords(ctx context.Context, req *GetHotKeywordsRequest) (*GetHotKeywordsResponse, error) {
	keywords, err := l.searchRepo.GetHotKeywords(ctx, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("获取搜索热词失败")
	}

	return &GetHotKeywordsResponse{
		Keywords: keywords,
	}, nil
}

// BuildProductIndexRequest 构建商品索引请求
type BuildProductIndexRequest struct {
	ProductIDs []uint64
}

// BuildProductIndex 构建商品索引
func (l *SearchLogic) BuildProductIndex(ctx context.Context, req *BuildProductIndexRequest) error {
	err := l.searchRepo.BuildProductIndex(ctx, req.ProductIDs)
	if err != nil {
		return apperrors.NewInternalError("构建商品索引失败")
	}
	return nil
}
