package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// RecommendRepository 推荐仓库接口
type RecommendRepository interface {
	// GetPersonalizedRecommend 获取个性化推荐
	GetPersonalizedRecommend(ctx context.Context, userID uint64, limit int) ([]map[string]interface{}, error)

	// GetSimilarProducts 获取相似商品
	GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]map[string]interface{}, error)

	// GetHotProducts 获取热门商品
	GetHotProducts(ctx context.Context, categoryID uint64, limit int) ([]map[string]interface{}, error)

	// GetRealtimeRecommend 获取实时推荐
	GetRealtimeRecommend(ctx context.Context, userID uint64, limit int) ([]map[string]interface{}, error)
}

type recommendRepository struct {
	redis *redis.Client
}

// NewRecommendRepository 创建推荐仓库
func NewRecommendRepository(redis *redis.Client) RecommendRepository {
	return &recommendRepository{redis: redis}
}

// GetPersonalizedRecommend 获取个性化推荐
func (r *recommendRepository) GetPersonalizedRecommend(ctx context.Context, userID uint64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Redis Key 设计：user:personalized:{userID}
	key := fmt.Sprintf("user:personalized:%d", userID)

	// 从 Sorted Set 中按分数倒序获取（推荐分数越高越靠前）
	result, err := r.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	return convertToProductMaps(result), nil
}

// GetSimilarProducts 获取相似商品
func (r *recommendRepository) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Redis Key 设计：product:similar:{productID}
	key := fmt.Sprintf("product:similar:%d", productID)

	result, err := r.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	return convertToProductMaps(result), nil
}

// GetHotProducts 获取热门商品
func (r *recommendRepository) GetHotProducts(ctx context.Context, categoryID uint64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var key string
	if categoryID > 0 {
		key = fmt.Sprintf("recommend:hot:products:%d", categoryID)
	} else {
		key = "recommend:hot:products"
	}

	// 按热度分数倒序获取
	result, err := r.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	return convertToProductMaps(result), nil
}

// GetRealtimeRecommend 获取实时推荐
func (r *recommendRepository) GetRealtimeRecommend(ctx context.Context, userID uint64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Redis Key 设计：user:realtime:{userID}
	key := fmt.Sprintf("user:realtime:%d", userID)

	result, err := r.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	return convertToProductMaps(result), nil
}

// convertToProductMaps 辅助函数：将 Redis ZSet 结果转换为 []map[string]interface{}
func convertToProductMaps(zs []redis.Z) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0, len(zs))
	for _, z := range zs {
		productIDStr, ok := z.Member.(string)
		if !ok {
			continue
		}

		productID, _ := strconv.ParseUint(productIDStr, 10, 64)

		item := map[string]interface{}{
			"product_id": productID,
			"score":      z.Score, // 推荐分 / 相似度 / 热度等
		}
		maps = append(maps, item)
	}
	return maps
}
