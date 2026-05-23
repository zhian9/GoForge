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
	productImage := ""

	// 如果前端没传产品信息，自动从商品服务查询（先查 SKU，再查 Product）
	if productName == "" || price == 0 {
		if s.svcCtx.ProductRpc != nil {
			skuResp, err := s.svcCtx.ProductRpc.GetSku(ctx, &productpb.GetSkuRequest{Id: req.SkuId})
			if err == nil && skuResp != nil && skuResp.Data != nil {
				productName = skuResp.Data.Name
				price = skuResp.Data.Price
				productImage = skuResp.Data.Image
			} else {
				// SKU 不存在时回退到按 Product ID 查询
				prodResp, err2 := s.svcCtx.ProductRpc.GetProduct(ctx, &productpb.GetProductRequest{Id: req.SkuId})
				if err2 == nil && prodResp != nil && prodResp.Data != nil {
					productName = prodResp.Data.Name
					price = prodResp.Data.Price
					productImage = prodResp.Data.MainImage
				}
			}
		}
	}

	// 如果前端已传图片，保留
	if productImage == "" && req.ProductImage != "" {
		productImage = req.ProductImage
	}

	addReq := &service.AddItemRequest{
		UserID:       userID,
		SkuID:        uint64(req.SkuId),
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
