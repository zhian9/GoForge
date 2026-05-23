package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zhian9/GoForge/server/api/product/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/product/model"
	"github.com/zhian9/GoForge/server/internal/service/product/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductResponse, error) {
	// 转换请求
	getReq := &service.GetProductRequest{
		ID:      uint64(req.Id),
		SpuCode: req.SpuCode,
	}

	// 调用业务逻辑
	resp, err := s.logic.GetProduct(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	productProto := convertProductToProto(resp.Product)
	// 注意：GetProductResponse 中没有 Skus 字段，SKU 信息需要通过 GetSku 方法单独获取

	return &v1.GetProductResponse{
		Code:    0,
		Message: "成功",
		Data:    productProto,
	}, nil
}

// ListProducts 获取商品列表
func (s *ProductService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	// 转换请求
	// 注意：proto 的 int32 字段如果没有传递，默认值是 0
	// 我们需要使用 -1 表示"查询所有状态"，0 表示"查询 status=0 的商品"
	// 但为了兼容性，当 status=0 时，我们将其视为"查询所有状态"
	var status int8 = -1 // 默认查询所有状态
	if req.Status > 0 {
		status = int8(req.Status)
	}

	// 处理 is_hot 参数
	// proto 的 int32 字段如果没有传递，默认值是 0
	// 为了区分"未传递"和"传递了0"，我们使用简单规则：
	// - 如果 req.IsHot == -1，明确表示查询全部
	// - 如果 req.IsHot == 1，明确表示查询热门
	// - 如果 req.IsHot == 0，可能是"未传递"（默认值）或"明确传递0"
	//   为了兼容，默认将 0 视为"未传递"，查询全部
	//   如果前端需要查询非热门商品，应该明确传递其他标识
	var isHot int8 = -1 // 默认查询全部
	if req.IsHot == -1 {
		isHot = -1 // 明确查询全部
	} else if req.IsHot == 1 {
		isHot = 1 // 明确查询热门
	} else {
		// req.IsHot 是 0，可能是"未传递"（默认值）或"明确传递0"
		// 为了兼容，默认查询全部
		isHot = -1
	}

	listReq := &service.ListProductsRequest{
		CategoryID: uint64(req.CategoryId),
		BrandID:    uint64(req.BrandId),
		Keyword:    req.Keyword,
		Status:     status,
		IsHot:      isHot,
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
		Sort:       req.Sort,
	}

	// 调用业务逻辑
	resp, err := s.logic.ListProducts(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	products := make([]*v1.Product, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, convertProductToProto(p))
	}

	return &v1.ListProductsResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.ProductListData{
			List:       products,
			Page:       int32(resp.Page),
			PageSize:   int32(resp.PageSize),
			Total:      resp.Total,
			TotalPages: int32(resp.TotalPages),
		},
	}, nil
}

// GetSku 获取SKU详情
func (s *ProductService) GetSku(ctx context.Context, req *v1.GetSkuRequest) (*v1.GetSkuResponse, error) {
	// 转换请求
	getReq := &service.GetSkuRequest{
		ID:      uint64(req.Id),
		SkuCode: req.SkuCode,
	}

	// 调用业务逻辑
	resp, err := s.logic.GetSku(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.GetSkuResponse{
		Code:    0,
		Message: "成功",
		Data:    convertSkuToProto(resp.Sku),
	}, nil
}

// ListSkus 获取SKU列表（管理后台）
func (s *ProductService) ListSkus(ctx context.Context, req *v1.ListSkusRequest) (*v1.ListSkusResponse, error) {
	// 转换请求
	// proto 的 int32 字段如果没有传递，默认值是 0
	// 但我们需要区分：0 可能是"未传递"（应该查询全部）或"明确传递0"（查询下架）
	// 由于 proto 无法区分，我们使用 -1 表示查询全部，0 表示下架，1 表示上架
	var status int8 = -1 // 默认查询所有状态
	if req.Status == -1 {
		status = -1 // 明确查询全部
	} else if req.Status >= 0 {
		status = int8(req.Status) // 0 表示下架，1 表示上架
	}

	// 处理分页参数，确保有默认值
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	listReq := &service.ListSkusRequest{
		ProductID: uint64(req.ProductId),
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	}

	// 调用业务逻辑
	resp, err := s.logic.ListSkus(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	skus := make([]*v1.Sku, 0, len(resp.Skus))
	for _, sku := range resp.Skus {
		if sku != nil {
			skuProto := convertSkuToProto(sku)
			if skuProto != nil {
				skus = append(skus, skuProto)
			}
		}
	}

	// 调试日志
	fmt.Printf("[ListSkus] Request: ProductID=%d, Status=%d, Page=%d, PageSize=%d, Response: Total=%d, Found=%d, Converted=%d\\n",
		req.ProductId, req.Status, page, pageSize, resp.Total, len(resp.Skus), len(skus))

	return &v1.ListSkusResponse{
		Code:    0,
		Message: "成功",
		Data: &v1.SkuListData{
			List:       skus,
			Page:       int32(resp.Page),
			PageSize:   int32(resp.PageSize),
			Total:      resp.Total,
			TotalPages: int32(resp.TotalPages),
		},
	}, nil
}

// CreateSku 创建SKU（管理后台）
func (s *ProductService) CreateSku(ctx context.Context, req *v1.CreateSkuRequest) (*v1.CreateSkuResponse, error) {
	// 转换请求
	createReq := &service.CreateSkuRequest{
		ProductID: uint64(req.ProductId),
		SkuCode:   req.SkuCode,
		Name:      req.Name,
		Specs:     req.Specs,
		Price:     req.Price,
		Stock:     int(req.Stock),
		Image:     req.Image,
		Status:    int8(req.Status),
	}

	if req.OriginalPrice > 0 {
		originalPrice := req.OriginalPrice
		createReq.OriginalPrice = &originalPrice
	}
	if req.Weight > 0 {
		weight := req.Weight
		createReq.Weight = &weight
	}
	if req.Volume > 0 {
		volume := req.Volume
		createReq.Volume = &volume
	}

	// 调用业务逻辑
	resp, err := s.logic.CreateSku(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.CreateSkuResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertSkuToProto(resp.Sku),
	}, nil
}

// UpdateSku 更新SKU（管理后台）
func (s *ProductService) UpdateSku(ctx context.Context, req *v1.UpdateSkuRequest) (*v1.UpdateSkuResponse, error) {
	// 转换请求
	updateReq := &service.UpdateSkuRequest{
		ID:      uint64(req.Id),
		SkuCode: req.SkuCode,
		Name:    req.Name,
		Specs:   req.Specs,
		Price:   req.Price,
		Stock:   int(req.Stock),
		Image:   req.Image,
		Status:  int8(req.Status),
	}

	if req.OriginalPrice > 0 {
		originalPrice := req.OriginalPrice
		updateReq.OriginalPrice = &originalPrice
	}
	if req.Weight > 0 {
		weight := req.Weight
		updateReq.Weight = &weight
	}
	if req.Volume > 0 {
		volume := req.Volume
		updateReq.Volume = &volume
	}

	// 调用业务逻辑
	resp, err := s.logic.UpdateSku(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateSkuResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertSkuToProto(resp.Sku),
	}, nil
}

// DeleteSku 删除SKU（管理后台）
func (s *ProductService) DeleteSku(ctx context.Context, req *v1.DeleteSkuRequest) (*v1.DeleteSkuResponse, error) {
	// 转换请求
	deleteReq := &service.DeleteSkuRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	_, err := s.logic.DeleteSku(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.DeleteSkuResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// GetCategoryList 获取类目列表
func (s *ProductService) GetCategoryList(ctx context.Context, req *v1.GetCategoryListRequest) (*v1.GetCategoryListResponse, error) {
	// 转换请求
	getReq := &service.GetCategoryListRequest{
		ParentID: uint64(req.ParentId),
		Level:    int8(req.Level),
		Status:   int8(req.Status),
	}

	// 调用业务逻辑
	resp, err := s.logic.GetCategoryList(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	categories := make([]*v1.Category, 0, len(resp.Categories))
	for _, cat := range resp.Categories {
		categories = append(categories, convertCategoryToProto(cat))
	}

	return &v1.GetCategoryListResponse{
		Code:    0,
		Message: "成功",
		Data:    categories,
	}, nil
}

// GetCategoryTree 获取类目树
func (s *ProductService) GetCategoryTree(ctx context.Context, req *v1.GetCategoryTreeRequest) (*v1.GetCategoryTreeResponse, error) {
	// 转换请求
	getReq := &service.GetCategoryTreeRequest{
		Status: int8(req.Status),
	}

	// 调用业务逻辑
	resp, err := s.logic.GetCategoryTree(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	categories := make([]*v1.Category, 0)
	for _, node := range resp.Categories {
		categories = append(categories, convertCategoryTreeNodeToProto(node))
	}

	return &v1.GetCategoryTreeResponse{
		Code:    0,
		Message: "成功",
		Data:    categories,
	}, nil
}

// CreateProduct 创建商品（管理后台）
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	// 转换请求
	createReq := &service.CreateProductRequest{
		Name:           req.Name,
		Subtitle:       req.Subtitle,
		CategoryID:     uint64(req.CategoryId),
		MainImage:      req.MainImage,
		LocalMainImage: req.LocalMainImage,
		Images:         req.Images,
		LocalImages:    req.LocalImages,
		Detail:         req.Detail,
		Price:          req.Price,
		OriginalPrice:  req.OriginalPrice,
		Stock:          int(req.Stock),
		Status:         int8(req.Status),
		IsHot:          int8(req.IsHot),
	}
	if req.BrandId > 0 {
		brandID := uint64(req.BrandId)
		createReq.BrandID = &brandID
	}

	// 调用业务逻辑
	resp, err := s.logic.CreateProduct(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.CreateProductResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertProductToProto(resp.Product),
	}, nil
}

// UpdateProduct 更新商品（管理后台）
func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error) {
	// 添加详细日志，追踪图片字段
	fmt.Printf("[UpdateProduct] 请求 - ID: %d, MainImage: '%s', LocalMainImage: '%s', Images: %v, LocalImages: %v\\n",
		req.Id, req.MainImage, req.LocalMainImage, req.Images, req.LocalImages)

	// 转换请求
	// 注意：proto 的字段如果没有传递，会有默认值（string 是空字符串，int32 是 0）
	// 为了支持部分更新，我们需要区分"未传递"和"传递了默认值"
	// 但由于 proto 的限制，我们使用一个特殊值 -999 表示"未传递"（对于 Status 和 IsHot）
	// 或者检查其他字段来判断是否是部分更新
	updateReq := &service.UpdateProductRequest{
		ID:             uint64(req.Id),
		Name:           req.Name,
		Subtitle:       req.Subtitle,
		CategoryID:     uint64(req.CategoryId), // 必须传递，前端已验证
		MainImage:      req.MainImage,
		LocalMainImage: req.LocalMainImage,
		Images:         req.Images,
		LocalImages:    req.LocalImages,
		Detail:         req.Detail,
		Price:          req.Price,
		OriginalPrice:  req.OriginalPrice,
		Stock:          int(req.Stock),
		// Status 和 IsHot 使用 -999 表示未传递（前端应该传递 -999 表示不更新）
		// 如果前端传递了有效值（>= 0），则使用该值
		Status: int8(req.Status),
		IsHot:  int8(req.IsHot),
	}
	if req.BrandId > 0 {
		brandID := uint64(req.BrandId)
		updateReq.BrandID = &brandID
	}

	// 验证分类ID（如果提供了分类ID，则验证；如果只更新其他字段如is_hot，则不验证）
	// 注意：如果 CategoryID 为 0 且其他必填字段也为空，说明可能是部分更新，从数据库获取现有值
	if updateReq.CategoryID == 0 && req.Name == "" && req.Price == 0 {
		// 可能是部分更新（如只更新 is_hot），不验证 category_id
		// 但需要确保至少有一个字段被更新
	} else if updateReq.CategoryID == 0 {
		return nil, status.Error(codes.InvalidArgument, "分类ID不能为空")
	}

	// 记录转换后的请求
	fmt.Printf("[UpdateProduct] 转换后 - MainImage: '%s', LocalMainImage: '%s', Images: %v, LocalImages: %v\\n",
		updateReq.MainImage, updateReq.LocalMainImage, updateReq.Images, updateReq.LocalImages)

	// 调用业务逻辑
	resp, err := s.logic.UpdateProduct(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateProductResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertProductToProto(resp.Product),
	}, nil
}

// DeleteProduct 删除商品（管理后台）
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductResponse, error) {
	// 转换请求
	deleteReq := &service.DeleteProductRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	_, err := s.logic.DeleteProduct(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.DeleteProductResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}

	// 检查是否是 BusinessError
	if bizErr, ok := err.(*apperrors.BusinessError); ok {
		var grpcCode codes.Code
		switch bizErr.Code {
		case apperrors.CodeNotFound, apperrors.CodeProductNotFound, apperrors.CodeSkuNotFound, apperrors.CodeCategoryNotFound:
			grpcCode = codes.NotFound
		case apperrors.CodeInvalidParam:
			grpcCode = codes.InvalidArgument
		case apperrors.CodeUnauthorized:
			grpcCode = codes.Unauthenticated
		case apperrors.CodeForbidden:
			grpcCode = codes.PermissionDenied
		default:
			grpcCode = codes.Internal
		}
		return status.Error(grpcCode, bizErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// GetCategory 获取类目详情
func (s *ProductService) GetCategory(ctx context.Context, req *v1.GetCategoryRequest) (*v1.GetCategoryResponse, error) {
	// 转换请求
	getReq := &service.GetCategoryRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	resp, err := s.logic.GetCategory(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.GetCategoryResponse{
		Code:    0,
		Message: "成功",
		Data:    convertCategoryToProto(resp.Category),
	}, nil
}

// CreateCategory 创建类目（管理后台）
func (s *ProductService) CreateCategory(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.CreateCategoryResponse, error) {
	// 转换请求
	createReq := &service.CreateCategoryRequest{
		ParentID:    uint64(req.ParentId),
		Name:        req.Name,
		Level:       int8(req.Level),
		Sort:        int(req.Sort),
		Icon:        req.Icon,
		IconLocal:   req.IconLocal,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Description: req.Description,
		Status:      int8(req.Status),
	}

	// 调用业务逻辑
	resp, err := s.logic.CreateCategory(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.CreateCategoryResponse{
		Code:    0,
		Message: "创建成功",
		Data:    convertCategoryToProto(resp.Category),
	}, nil
}

// UpdateCategory 更新类目（管理后台）
func (s *ProductService) UpdateCategory(ctx context.Context, req *v1.UpdateCategoryRequest) (*v1.UpdateCategoryResponse, error) {
	// 转换请求
	updateReq := &service.UpdateCategoryRequest{
		ID:          uint64(req.Id),
		ParentID:    uint64(req.ParentId),
		Name:        req.Name,
		Level:       int8(req.Level),
		Sort:        int(req.Sort),
		Icon:        req.Icon,
		IconLocal:   req.IconLocal,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Description: req.Description,
		Status:      int8(req.Status),
	}

	// 调用业务逻辑
	resp, err := s.logic.UpdateCategory(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateCategoryResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertCategoryToProto(resp.Category),
	}, nil
}

// DeleteCategory 删除类目（管理后台）
func (s *ProductService) DeleteCategory(ctx context.Context, req *v1.DeleteCategoryRequest) (*v1.DeleteCategoryResponse, error) {
	// 转换请求
	deleteReq := &service.DeleteCategoryRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	_, err := s.logic.DeleteCategory(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.DeleteCategoryResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// convertProductToProto 转换商品模型为 Protobuf 消息
func convertProductToProto(p *model.Product) *v1.Product {
	if p == nil {
		return nil
	}

	// 解析图片列表
	images, _ := parseJSONArray(p.Images)
	localImages, _ := parseJSONArray(p.LocalImages)

	product := &v1.Product{
		Id:             int64(p.ID),
		SpuCode:        p.SpuCode,
		Name:           p.Name,
		Subtitle:       p.Subtitle,
		CategoryId:     int64(p.CategoryID),
		MainImage:      p.MainImage,
		LocalMainImage: p.LocalMainImage,
		Images:         images,
		LocalImages:    localImages,
		Detail:         p.Detail,
		Price:          p.Price,
		Stock:          int32(p.Stock),
		Sales:          int32(p.Sales),
		Status:         int32(p.Status),
		IsHot:          int32(p.IsHot),
		CreatedAt:      formatTime(&p.CreatedAt),
		UpdatedAt:      formatTime(&p.UpdatedAt),
	}

	if p.BrandID != nil {
		product.BrandId = int64(*p.BrandID)
	}
	if p.OriginalPrice != nil {
		product.OriginalPrice = *p.OriginalPrice
	}

	return product
}

// convertSkuToProto 转换SKU模型为 Protobuf 消息
func convertSkuToProto(sku *model.Sku) *v1.Sku {
	if sku == nil {
		return nil
	}

	// 解析规格
	specs, _ := parseJSONMap(sku.Specs)

	skuProto := &v1.Sku{
		Id:        int64(sku.ID),
		ProductId: int64(sku.ProductID),
		SkuCode:   sku.SkuCode,
		Name:      sku.Name,
		Specs:     specs,
		Price:     sku.Price,
		Stock:     int32(sku.Stock),
		Image:     sku.Image,
		Status:    int32(sku.Status),
	}

	if sku.OriginalPrice != nil {
		skuProto.OriginalPrice = *sku.OriginalPrice
	}
	if sku.Weight != nil {
		skuProto.Weight = *sku.Weight
	}
	if sku.Volume != nil {
		skuProto.Volume = *sku.Volume
	}

	return skuProto
}

// convertCategoryToProto 转换类目模型为 Protobuf 消息
func convertCategoryToProto(cat *model.Category) *v1.Category {
	if cat == nil {
		return nil
	}

	return &v1.Category{
		Id:          int64(cat.ID),
		ParentId:    int64(cat.ParentID),
		Name:        cat.Name,
		Level:       int32(cat.Level),
		Sort:        int32(cat.Sort),
		Icon:        cat.Icon,
		IconLocal:   cat.IconLocal,
		Image:       cat.Image,
		ImageLocal:  cat.ImageLocal,
		Description: cat.Description,
		Status:      int32(cat.Status),
	}
}

// convertCategoryTreeNodeToProto 转换类目树节点为 Protobuf 消息
func convertCategoryTreeNodeToProto(node *service.CategoryTreeNode) *v1.Category {
	if node == nil {
		return nil
	}

	cat := convertCategoryToProto(node.Category)
	if cat == nil {
		return nil
	}

	// 递归转换子节点
	if len(node.Children) > 0 {
		cat.Children = make([]*v1.Category, 0, len(node.Children))
		for _, child := range node.Children {
			cat.Children = append(cat.Children, convertCategoryTreeNodeToProto(child))
		}
	}

	return cat
}

// parseJSONArray 解析JSON数组字符串
func parseJSONArray(jsonStr string) ([]string, error) {
	if jsonStr == "" {
		return []string{}, nil
	}
	var result []string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// parseJSONMap 解析JSON对象字符串
func parseJSONMap(jsonStr string) (map[string]string, error) {
	if jsonStr == "" {
		return map[string]string{}, nil
	}
	var result map[string]string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// parseTime 解析时间字符串
func parseTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ==================== Banner 相关方法 ====================

// ListBanners 获取Banner列表
func (s *ProductService) ListBanners(ctx context.Context, req *v1.ListBannersRequest) (*v1.ListBannersResponse, error) {
	// 转换请求
	listReq := &service.ListBannersRequest{
		Status: int8(req.Status),
		Limit:  int(req.Limit),
	}

	// 调用业务逻辑
	resp, err := s.logic.ListBanners(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	banners := make([]*v1.Banner, 0, len(resp.Banners))
	for _, b := range resp.Banners {
		banners = append(banners, convertBannerToProto(b))
	}

	return &v1.ListBannersResponse{
		Code:    0,
		Message: "成功",
		Data:    banners,
	}, nil
}

// GetBanner 获取Banner详情
func (s *ProductService) GetBanner(ctx context.Context, req *v1.GetBannerRequest) (*v1.GetBannerResponse, error) {
	// 转换请求
	getReq := &service.GetBannerRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	resp, err := s.logic.GetBanner(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.GetBannerResponse{
		Code:    0,
		Message: "成功",
		Data:    convertBannerToProto(resp.Banner),
	}, nil
}

// CreateBanner 创建Banner（管理后台）
func (s *ProductService) CreateBanner(ctx context.Context, req *v1.CreateBannerRequest) (*v1.CreateBannerResponse, error) {
	// 解析时间
	startTime, _ := parseTime(req.StartTime)
	endTime, _ := parseTime(req.EndTime)

	// 转换请求
	createReq := &service.CreateBannerRequest{
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Link:        req.Link,
		LinkType:    int8(req.LinkType),
		Sort:        int(req.Sort),
		Status:      int8(req.Status),
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// 调用业务逻辑
	resp, err := s.logic.CreateBanner(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.CreateBannerResponse{
		Code:    0,
		Message: "成功",
		Data:    convertBannerToProto(resp.Banner),
	}, nil
}

// UpdateBanner 更新Banner（管理后台）
func (s *ProductService) UpdateBanner(ctx context.Context, req *v1.UpdateBannerRequest) (*v1.UpdateBannerResponse, error) {
	// 解析时间
	startTime, _ := parseTime(req.StartTime)
	endTime, _ := parseTime(req.EndTime)

	// 转换请求
	updateReq := &service.UpdateBannerRequest{
		ID:          uint64(req.Id),
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Link:        req.Link,
		LinkType:    int8(req.LinkType),
		Sort:        int(req.Sort),
		Status:      int8(req.Status),
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// 调用业务逻辑
	resp, err := s.logic.UpdateBanner(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	// 转换响应
	return &v1.UpdateBannerResponse{
		Code:    0,
		Message: "成功",
		Data:    convertBannerToProto(resp.Banner),
	}, nil
}

// DeleteBanner 删除Banner（管理后台）
func (s *ProductService) DeleteBanner(ctx context.Context, req *v1.DeleteBannerRequest) (*v1.DeleteBannerResponse, error) {
	// 转换请求
	deleteReq := &service.DeleteBannerRequest{
		ID: uint64(req.Id),
	}

	// 调用业务逻辑
	_, err := s.logic.DeleteBanner(ctx, deleteReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.DeleteBannerResponse{
		Code:    0,
		Message: "成功",
	}, nil
}

// convertBannerToProto 转换Banner模型为Proto
func convertBannerToProto(banner *model.Banner) *v1.Banner {
	if banner == nil {
		return nil
	}

	return &v1.Banner{
		Id:          int64(banner.ID),
		Title:       banner.Title,
		Description: banner.Description,
		Image:       banner.Image,
		ImageLocal:  banner.ImageLocal,
		Link:        banner.Link,
		LinkType:    int32(banner.LinkType),
		Sort:        int32(banner.Sort),
		Status:      int32(banner.Status),
		StartTime:   formatTime(banner.StartTime),
		EndTime:     formatTime(banner.EndTime),
		CreatedAt:   banner.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   banner.UpdatedAt.Format(time.RFC3339),
	}
}
