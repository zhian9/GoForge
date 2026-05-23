package service

import (
	"context"

	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/service/cart/model"
	"github.com/zhian9/GoForge/server/internal/service/cart/repository"
)

// CartLogic 购物车业务逻辑
type CartLogic struct {
	cartRepo repository.CartRepository
}

// NewCartLogic 创建购物车业务逻辑
func NewCartLogic(cartRepo repository.CartRepository) *CartLogic {
	return &CartLogic{
		cartRepo: cartRepo,
	}
}

// GetCartRequest 获取购物车请求
type GetCartRequest struct {
	UserID uint64
}

// GetCartResponse 获取购物车响应
type GetCartResponse struct {
	Items []*model.Cart
}

// GetCart 获取购物车
func (l *CartLogic) GetCart(ctx context.Context, req *GetCartRequest) (*GetCartResponse, error) {
	items, err := l.cartRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.NewInternalError("获取购物车失败")
	}

	return &GetCartResponse{
		Items: items,
	}, nil
}

// AddItemRequest 添加商品请求
type AddItemRequest struct {
	UserID       uint64
	SkuID        uint64
	ProductName  string
	Price        float64
	ProductImage string
	Quantity     int
}

// AddItemResponse 添加商品响应
type AddItemResponse struct {
	Cart *model.Cart
}

// AddItem 添加商品到购物车
func (l *CartLogic) AddItem(ctx context.Context, req *AddItemRequest) (*AddItemResponse, error) {
	cart := &model.Cart{
		UserID:       req.UserID,
		SkuID:        req.SkuID,
		ProductName:  req.ProductName,
		Price:        req.Price,
		ProductImage: req.ProductImage,
		Quantity:     req.Quantity,
		IsSelected:   1,
	}

	err := l.cartRepo.AddItem(ctx, cart)
	if err != nil {
		return nil, apperrors.NewInternalError("添加商品失败")
	}

	return &AddItemResponse{
		Cart: cart,
	}, nil
}

// UpdateQuantityRequest 更新数量请求
type UpdateQuantityRequest struct {
	UserID   uint64
	SkuID    uint64
	Quantity int
}

// UpdateQuantity 更新商品数量
func (l *CartLogic) UpdateQuantity(ctx context.Context, req *UpdateQuantityRequest) error {
	if req.Quantity <= 0 {
		return apperrors.NewInvalidParamError("数量必须大于0")
	}

	err := l.cartRepo.UpdateQuantity(ctx, req.UserID, req.SkuID, req.Quantity)
	if err != nil {
		return apperrors.NewInternalError("更新数量失败")
	}

	return nil
}

// RemoveItemRequest 删除商品请求
type RemoveItemRequest struct {
	UserID uint64
	SkuIDs []uint64
}

// RemoveItem 删除商品
func (l *CartLogic) RemoveItem(ctx context.Context, req *RemoveItemRequest) error {
	err := l.cartRepo.RemoveItem(ctx, req.UserID, req.SkuIDs)
	if err != nil {
		return apperrors.NewInternalError("删除商品失败")
	}

	return nil
}

// ClearCartRequest 清空购物车请求
type ClearCartRequest struct {
	UserID uint64
}

// ClearCart 清空购物车
func (l *CartLogic) ClearCart(ctx context.Context, req *ClearCartRequest) error {
	err := l.cartRepo.ClearCart(ctx, req.UserID)
	if err != nil {
		return apperrors.NewInternalError("清空购物车失败")
	}

	return nil
}

// SelectItemRequest 选择商品请求
type SelectItemRequest struct {
	UserID     uint64
	SkuID      uint64
	IsSelected int8
}

// SelectItem 选择/取消选择商品
func (l *CartLogic) SelectItem(ctx context.Context, req *SelectItemRequest) error {
	err := l.cartRepo.SelectItem(ctx, req.UserID, req.SkuID, req.IsSelected)
	if err != nil {
		return apperrors.NewInternalError("操作失败")
	}

	return nil
}

// BatchSelectRequest 批量选择请求
type BatchSelectRequest struct {
	UserID     uint64
	SkuIDs     []uint64
	IsSelected int8
}

// BatchSelect 批量选择/取消选择
func (l *CartLogic) BatchSelect(ctx context.Context, req *BatchSelectRequest) error {
	err := l.cartRepo.BatchSelect(ctx, req.UserID, req.SkuIDs, req.IsSelected)
	if err != nil {
		return apperrors.NewInternalError("批量操作失败")
	}

	return nil
}
