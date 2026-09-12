package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "github.com/zhian9/GoForge/server/api/cart/v1"
	productpb "github.com/zhian9/GoForge/server/api/product/v1"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/utils"
	"github.com/zhian9/GoForge/server/internal/service/cart/model"
	"github.com/zhian9/GoForge/server/internal/service/cart/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CartService 实现 gRPC 服务接口
type CartService struct {
	v1.UnimplementedCartServiceServer
	svcCtx *ServiceContext
	logic  *service.CartLogic
}

// NewCartService 创建购物车服务
func NewCartService(svcCtx *ServiceContext) *CartService {
	logic := service.NewCartLogic(svcCtx.CartRepo)

	return &CartService{
		svcCtx: svcCtx,
		logic:  logic,
	}
}

// GetCart 获取购物车
func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartResponse, error) {
	// 从 context 中获取 user_id（由 JWT 中间件设置）
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		// 如果 context 中没有，尝试从请求参数获取（兼容性）
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	getReq := &service.GetCartRequest{
		UserID: userID,
	}

	resp, err := s.logic.GetCart(ctx, getReq)
	if err != nil {
		return nil, convertError(err)
	}

	items := make([]*v1.CartItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, convertCartItemToProto(item))
	}

	// 返回格式需要匹配前端期望的结构
	// 注意：proto 定义中 data 是 repeated CartItem，直接返回数组
	return &v1.GetCartResponse{
		Code:    0,
		Message: "成功",
		Data:    items,
	}, nil
}

// AddItem 添加商品到购物车
func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.AddItemResponse, error) {
	// 从 context 中获取 user_id（由 JWT 中间件设置）
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		// 如果 context 中没有，尝试从请求参数获取（兼容性）
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	productName := req.ProductName
	price := parsePrice(req.Price)
	productImage := req.ProductImage

	// 前端传的 sku_id 可能是：
	//  - 详情页：真实 SKU ID
	//  - 首页/列表页快捷加购：Product ID（历史遗留，字段被复用）
	// 统一在这里解析成「真实 SKU ID + 商品名称/图片 + 价格」。
	// 名称与图片以 Product 为准（和首页/列表展示保持一致），价格以 SKU 为准。
	skuID := uint64(req.SkuId)
	if productName == "" || price == 0 {
		if s.svcCtx.ProductRpc != nil {
			skuID, productName, price, productImage = s.resolveCartItemInfo(ctx, req.SkuId)
		}
	}

	addReq := &service.AddItemRequest{
		UserID:       userID,
		SkuID:        skuID,
		ProductName:  productName,
		Price:        price,
		ProductImage: productImage,
		Quantity:     int(req.Quantity),
	}

	resp, err := s.logic.AddItem(ctx, addReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.AddItemResponse{
		Code:    0,
		Message: "添加成功",
		Data:    convertCartItemToProto(resp.Cart),
	}, nil
}

// resolveCartItemInfo 解析购物车商品项：返回真实 SKU ID、商品名称、价格、商品图片。
// 入参 rawID 可能是真实 SKU ID（详情页）或 Product ID（首页/列表页快捷加购）。
// 名称与图片以 Product 为准（和首页/列表展示一致），价格以 SKU 为准。
func (s *CartService) resolveCartItemInfo(ctx context.Context, rawID int64) (uint64, string, float64, string) {
	rpc := s.svcCtx.ProductRpc
	if rpc == nil {
		return uint64(rawID), "", 0, ""
	}

	// 1) 先按 SKU ID 查
	if skuResp, err := rpc.GetSku(ctx, &productpb.GetSkuRequest{Id: rawID}); err == nil && skuResp != nil && skuResp.Data != nil {
		sku := skuResp.Data
		name := sku.Name
		image := sku.Image
		// 用商品名称/主图覆盖，保证购物车与首页/列表展示一致
		if sku.ProductId > 0 {
			if prodResp, e := rpc.GetProduct(ctx, &productpb.GetProductRequest{Id: sku.ProductId}); e == nil && prodResp != nil && prodResp.Data != nil {
				name = prodResp.Data.Name
				image = firstNonEmpty(prodResp.Data.LocalMainImage, prodResp.Data.MainImage, sku.Image)
			}
		}
		return uint64(sku.Id), name, sku.Price, image
	}

	// 2) SKU 不存在：rawID 可能是 Product ID（首页/列表页快捷加购），解析出该商品的一个上架 SKU
	if prodResp, err := rpc.GetProduct(ctx, &productpb.GetProductRequest{Id: rawID}); err == nil && prodResp != nil && prodResp.Data != nil {
		prod := prodResp.Data
		name := prod.Name
		image := firstNonEmpty(prod.LocalMainImage, prod.MainImage)
		price := prod.Price
		skuID := uint64(rawID)
		// 取价格最低的上架 SKU，与首页/列表展示的最低价保持一致
		if skuResp, e := rpc.ListSkus(ctx, &productpb.ListSkusRequest{ProductId: rawID, Status: 1, Page: 1, PageSize: 100}); e == nil && skuResp != nil && skuResp.Data != nil && len(skuResp.Data.List) > 0 {
			var bestSku *productpb.Sku
			for _, sku := range skuResp.Data.List {
				if sku == nil || sku.Price <= 0 {
					continue
				}
				if bestSku == nil || sku.Price < bestSku.Price {
					bestSku = sku
				}
			}
			if bestSku != nil {
				skuID = uint64(bestSku.Id)
				price = bestSku.Price
			}
		}
		return skuID, name, price, image
	}

	return uint64(rawID), "", 0, ""
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// UpdateQuantity 更新商品数量
func (s *CartService) UpdateQuantity(ctx context.Context, req *v1.UpdateQuantityRequest) (*v1.UpdateQuantityResponse, error) {
	// 从 context 中获取 user_id（由 JWT 中间件设置）
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	updateReq := &service.UpdateQuantityRequest{
		UserID:   userID,
		SkuID:    uint64(req.SkuId),
		Quantity: int(req.Quantity),
	}

	err := s.logic.UpdateQuantity(ctx, updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.UpdateQuantityResponse{
		Code:    0,
		Message: "更新成功",
	}, nil
}

// RemoveItem 删除商品
func (s *CartService) RemoveItem(ctx context.Context, req *v1.RemoveItemRequest) (*v1.RemoveItemResponse, error) {
	// 从 context 中获取 user_id（由 JWT 中间件设置）
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		if req.UserId > 0 {
			userID = uint64(req.UserId)
		} else {
			return nil, status.Error(codes.Unauthenticated, "未授权，请先登录")
		}
	}

	skuIDs := make([]uint64, 0, len(req.SkuIds))
	for _, id := range req.SkuIds {
		skuIDs = append(skuIDs, uint64(id))
	}

	removeReq := &service.RemoveItemRequest{
		UserID: userID,
		SkuIDs: skuIDs,
	}

	err := s.logic.RemoveItem(ctx, removeReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.RemoveItemResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, req *v1.ClearCartRequest) (*v1.ClearCartResponse, error) {
	clearReq := &service.ClearCartRequest{
		UserID: uint64(req.UserId),
	}

	err := s.logic.ClearCart(ctx, clearReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.ClearCartResponse{
		Code:    0,
		Message: "清空成功",
	}, nil
}

// SelectItem 选择/取消选择商品
func (s *CartService) SelectItem(ctx context.Context, req *v1.SelectItemRequest) (*v1.SelectItemResponse, error) {
	selectReq := &service.SelectItemRequest{
		UserID:     uint64(req.UserId),
		SkuID:      uint64(req.SkuId),
		IsSelected: int8(req.IsSelected),
	}

	err := s.logic.SelectItem(ctx, selectReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.SelectItemResponse{
		Code:    0,
		Message: "操作成功",
	}, nil
}

// BatchSelect 批量选择/取消选择
func (s *CartService) BatchSelect(ctx context.Context, req *v1.BatchSelectRequest) (*v1.BatchSelectResponse, error) {
	skuIDs := make([]uint64, 0, len(req.SkuIds))
	for _, id := range req.SkuIds {
		skuIDs = append(skuIDs, uint64(id))
	}

	batchReq := &service.BatchSelectRequest{
		UserID:     uint64(req.UserId),
		SkuIDs:     skuIDs,
		IsSelected: int8(req.IsSelected),
	}

	err := s.logic.BatchSelect(ctx, batchReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &v1.BatchSelectResponse{
		Code:    0,
		Message: "操作成功",
	}, nil
}

// convertError 转换业务错误为 gRPC 错误
func convertError(err error) error {
	if err == nil {
		return nil
	}

	// 检查是否是 BusinessError
	var bizErr *apperrors.BusinessError
	if errors.As(err, &bizErr) {
		var grpcCode codes.Code
		switch bizErr.Code {
		case apperrors.CodeNotFound:
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

// convertCartItemToProto 转换购物车商品为 Protobuf 消息
func convertCartItemToProto(item *model.Cart) *v1.CartItem {
	if item == nil {
		return nil
	}

	return &v1.CartItem{
		Id:          int64(item.ID),
		UserId:      int64(item.UserID),
		SkuId:       int64(item.SkuID),
		ProductName:  item.ProductName,
		Price:        formatPrice(item.Price),
		ProductImage: item.ProductImage,
		Quantity:    int32(item.Quantity),
		IsSelected:  int32(item.IsSelected),
		CreatedAt:   formatTime(&item.CreatedAt),
		UpdatedAt:   formatTime(&item.UpdatedAt),
	}
}

// parsePrice 解析价格字符串为 float64
func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// formatPrice 格式化价格
func formatPrice(f float64) string {
	if f == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", f)
}

// formatTime 格式化时间为字符串
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
