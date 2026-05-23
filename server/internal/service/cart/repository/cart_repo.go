package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhian9/GoForge/server/internal/service/cart/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CartRepository 购物车仓库接口
type CartRepository interface {
	// GetByUserID 获取用户购物车（从Redis）
	GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error)
	// AddItem 添加商品到购物车
	AddItem(ctx context.Context, cart *model.Cart) error
	// UpdateQuantity 更新商品数量
	UpdateQuantity(ctx context.Context, userID, skuID uint64, quantity int) error
	// RemoveItem 删除商品
	RemoveItem(ctx context.Context, userID uint64, skuIDs []uint64) error
	// ClearCart 清空购物车
	ClearCart(ctx context.Context, userID uint64) error
	// SelectItem 选择/取消选择商品
	SelectItem(ctx context.Context, userID, skuID uint64, isSelected int8) error
	// BatchSelect 批量选择/取消选择
	BatchSelect(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error
	// SyncToDB 同步到数据库（持久化）
	SyncToDB(ctx context.Context, userID uint64) error
}

type cartRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewCartRepository 创建购物车仓库
func NewCartRepository(db *gorm.DB, redis *redis.Client) CartRepository {
	return &cartRepository{
		db:    db,
		redis: redis,
	}
}

// getCartKey 获取购物车Redis key
func (r *cartRepository) getCartKey(userID uint64) string {
	return fmt.Sprintf("cart-service:user:%d", userID)
}

// GetByUserID 获取用户购物车（从Redis）
func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error) {
	key := r.getCartKey(userID)

	// 从Redis获取
	items, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []*model.Cart{}, nil
	}

	carts := make([]*model.Cart, 0, len(items))
	for _, itemJSON := range items {
		var cart model.Cart
		if err := json.Unmarshal([]byte(itemJSON), &cart); err != nil {
			continue
		}
		carts = append(carts, &cart)
	}

	return carts, nil
}

// AddItem 添加商品到购物车
func (r *cartRepository) AddItem(ctx context.Context, cart *model.Cart) error {
	key := r.getCartKey(cart.UserID)
	itemKey := fmt.Sprintf("%d", cart.SkuID)

	// 检查是否已存在
	exists, err := r.redis.HExists(ctx, key, itemKey).Result()
	if err != nil {
		return err
	}

	if exists {
		// 更新数量和价格
		var existingCart model.Cart
		itemJSON, _ := r.redis.HGet(ctx, key, itemKey).Result()
		err := json.Unmarshal([]byte(itemJSON), &existingCart)
		if err != nil {
			return err
		}
		existingCart.Quantity += cart.Quantity
		if cart.ProductName != "" {
			existingCart.ProductName = cart.ProductName
		}
		if cart.Price > 0 {
			existingCart.Price = cart.Price
		}
		if cart.ProductImage != "" {
			existingCart.ProductImage = cart.ProductImage
		}
		existingCart.UpdatedAt = time.Now()

		itemJSONBytes, _ := json.Marshal(existingCart)
		err = r.redis.HSet(ctx, key, itemKey, string(itemJSONBytes)).Err()
		if err != nil {
			return err
		}
		// 更新过期时间
		return r.redis.Expire(ctx, key, 30*24*time.Hour).Err()
	}

	// 新增
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()
	itemJSONBytes, _ := json.Marshal(cart)

	// 保存到 Redis
	err = r.redis.HSet(ctx, key, itemKey, string(itemJSONBytes)).Err()
	if err != nil {
		return err
	}

	// 设置过期时间30天（对整个 hash 设置过期时间）
	return r.redis.Expire(ctx, key, 30*24*time.Hour).Err()
}

// UpdateQuantity 更新商品数量
func (r *cartRepository) UpdateQuantity(ctx context.Context, userID, skuID uint64, quantity int) error {
	key := r.getCartKey(userID)
	itemKey := fmt.Sprintf("%d", skuID)

	itemJSON, err := r.redis.HGet(ctx, key, itemKey).Result()
	if err != nil {
		return err
	}

	var cart model.Cart
	if err := json.Unmarshal([]byte(itemJSON), &cart); err != nil {
		return err
	}

	cart.Quantity = quantity
	cart.UpdatedAt = time.Now()

	itemJSONBytes, _ := json.Marshal(cart)
	return r.redis.HSet(ctx, key, itemKey, string(itemJSONBytes)).Err()
}

// RemoveItem 删除商品
func (r *cartRepository) RemoveItem(ctx context.Context, userID uint64, skuIDs []uint64) error {
	key := r.getCartKey(userID)

	itemKeys := make([]string, 0, len(skuIDs))
	for _, skuID := range skuIDs {
		itemKeys = append(itemKeys, fmt.Sprintf("%d", skuID))
	}

	return r.redis.HDel(ctx, key, itemKeys...).Err()
}

// ClearCart 清空购物车
func (r *cartRepository) ClearCart(ctx context.Context, userID uint64) error {
	key := r.getCartKey(userID)
	return r.redis.Del(ctx, key).Err()
}

// SelectItem 选择/取消选择商品
func (r *cartRepository) SelectItem(ctx context.Context, userID, skuID uint64, isSelected int8) error {
	key := r.getCartKey(userID)
	itemKey := fmt.Sprintf("%d", skuID)

	itemJSON, err := r.redis.HGet(ctx, key, itemKey).Result()
	if err != nil {
		return err
	}

	var cart model.Cart
	if err := json.Unmarshal([]byte(itemJSON), &cart); err != nil {
		return err
	}

	cart.IsSelected = isSelected
	cart.UpdatedAt = time.Now()

	itemJSONBytes, _ := json.Marshal(cart)
	return r.redis.HSet(ctx, key, itemKey, string(itemJSONBytes)).Err()
}

// BatchSelect 批量选择/取消选择
func (r *cartRepository) BatchSelect(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error {
	key := r.getCartKey(userID)

	for _, skuID := range skuIDs {
		itemKey := fmt.Sprintf("%d", skuID)
		itemJSON, err := r.redis.HGet(ctx, key, itemKey).Result()
		if err != nil {
			continue
		}

		var cart model.Cart
		if err := json.Unmarshal([]byte(itemJSON), &cart); err != nil {
			continue
		}

		cart.IsSelected = isSelected
		cart.UpdatedAt = time.Now()

		itemJSONBytes, _ := json.Marshal(cart)
		r.redis.HSet(ctx, key, itemKey, string(itemJSONBytes))
	}

	return nil
}

// SyncToDB 同步到数据库（持久化）
func (r *cartRepository) SyncToDB(ctx context.Context, userID uint64) error {
	carts, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// 批量保存到数据库
	for _, cart := range carts {
		var existingCart model.Cart
		err := r.db.WithContext(ctx).
			Where("user_id = ? AND sku_id = ?", cart.UserID, cart.SkuID).
			First(&existingCart).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新增
			r.db.WithContext(ctx).Create(cart)
		} else {
			// 更新
			r.db.WithContext(ctx).Model(&existingCart).Updates(cart)
		}
	}

	return nil
}
