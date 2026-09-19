package search

import (
	"context"
	"strconv"

	v1 "github.com/zhian9/GoForge/server/api/search/v1"
	"github.com/zhian9/GoForge/server/internal/service/search/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SearchService 实现 gRPC 服务接口
type SearchService struct {
	v1.UnimplementedSearchServiceServer
	svcCtx *ServiceContext
	logic  *service.SearchLogic
}

// NewSearchService 创建搜索服务
func NewSearchService(svcCtx *ServiceContext) *SearchService {
	logic := service.NewSearchLogic(svcCtx.SearchRepo)

	return &SearchService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// SearchProducts 搜索商品
func (s *SearchService) SearchProducts(ctx context.Context, req *v1.SearchProductsRequest) (*v1.SearchProductsResponse, error) {
	searchReq := &service.SearchProductsRequest{
		Keyword:    req.Keyword,
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
		CategoryID: uint64(req.CategoryId),
		SortBy:     req.SortBy,
	}

	resp, err := s.logic.SearchProducts(ctx, searchReq)
	if err != nil {
		return nil, convertError(err)
	}

	results := make([]*v1.ProductSearchResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, &v1.ProductSearchResult{
			ProductId: r.ProductID,
			Name:      r.Name,
			MainImage: r.MainImage,
			Price:     strconv.FormatFloat(r.Price, 'f', 2, 64),
			Sales:     int32(r.Sales),
			Score:     r.Score,
		})
	}

	return &v1.SearchProductsResponse{
		Code:    0,
		Message: "成功",
		Data:    results,
		Total:   int32(resp.Total),
	}, nil
}

// GetSearchSuggestions 获取搜索建议
func (s *SearchService) GetSearchSuggestions(ctx context.Context, req *v1.GetSearchSuggestionsRequest) (*v1.GetSearchSuggestionsResponse, error) {
	getReq := &service.GetSearchSuggestionsRequest{
		Keyword: req.Keyword,
		Limit:   int(req.Limit),
	}

	resp, err := s.logic.GetSearchSuggestions(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetSearchSuggestionsResponse{
		Code:        0,
		Message:     "成功",
		Suggestions: resp.Suggestions,
	}, nil
}

// GetHotKeywords 获取搜索热词
func (s *SearchService) GetHotKeywords(ctx context.Context, req *v1.GetHotKeywordsRequest) (*v1.GetHotKeywordsResponse, error) {
	getReq := &service.GetHotKeywordsRequest{
		Limit: int(req.Limit),
	}

	resp, err := s.logic.GetHotKeywords(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.GetHotKeywordsResponse{
		Code:     0,
		Message:  "成功",
		Keywords: resp.Keywords,
	}, nil
}

// BuildProductIndex 构建商品索引
func (s *SearchService) BuildProductIndex(ctx context.Context, req *v1.BuildProductIndexRequest) (*v1.BuildProductIndexResponse, error) {
	// 不传 product_ids 表示「全量重建」——这是种子数据（直插 MySQL、不产生 outbox 事件）
	// 也能进入索引的唯一途径。
	if len(req.ProductIds) == 0 {
		if err := s.svcCtx.ReindexAllProducts(ctx); err != nil {
			return nil, convertError(err)
		}
		return &v1.BuildProductIndexResponse{
			Code:    0,
			Message: "全量重建完成",
		}, nil
	}

	productIDs := make([]uint64, 0, len(req.ProductIds))
	for _, id := range req.ProductIds {
		productIDs = append(productIDs, uint64(id))
	}

	buildReq := &service.BuildProductIndexRequest{
		ProductIDs: productIDs,
	}

	err := s.logic.BuildProductIndex(ctx, buildReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.BuildProductIndexResponse{
		Code:    0,
		Message: "构建成功",
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}
