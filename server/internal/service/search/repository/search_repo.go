package repository

import (
	"context"
	"github.com/zhian9/GoForge/server/internal/pkg/search"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// SearchRepository 搜索仓库接口
type SearchRepository interface {
	// SearchProducts 搜索商品
	SearchProducts(ctx context.Context, keyword string, category uint64, page, pageSize int, sortBy string) ([]map[string]interface{}, int64, error)
	// GetSearchSuggestions 获取搜索建议
	GetSearchSuggestions(ctx context.Context, keyword string, limit int) ([]string, error)
	// GetHotKeywords 获取搜索热词
	GetHotKeywords(ctx context.Context, limit int) ([]string, error)
	// BuildProductIndex 构建商品索引
	BuildProductIndex(ctx context.Context, productIDs []uint64) error
}

type searchRepository struct {
	redis    *redis.Client
	esClient *search.Client
}

// SearchProducts 搜索商品
func (s *searchRepository) SearchProducts(ctx context.Context, keyword string, categoryID uint64, page, pageSize int, sortBy string) ([]map[string]interface{}, int64, error) {
	if s.esClient == nil {
		return []map[string]interface{}{}, 0, nil
	}

	//构建查询
	query := map[string]interface{}{
		"from":  (page - 1) * pageSize,
		"size":  pageSize,
		"query": map[string]interface{}{},
	}

	//构建查询条件
	mustClauses := []map[string]interface{}{}

	//关键词搜索
	if keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  keyword,
				"fields": []string{"name^3", "subtitle^2", "detail"},
				"type":   "best_fields",
			},
		})
	}

	//类目筛选
	if categoryID > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": categoryID,
			},
		})
	}

	//状态筛选(只搜索上架商品)
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"status": 1, //1-上架
		},
	})

	if len(mustClauses) > 0 {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		}
	} else {
		query["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	//排序
	sort := []map[string]interface{}{}
	switch sortBy {
	case "sales":
		sort = append(sort, map[string]interface{}{"sales": "desc"})
	case "price_asc":
		sort = append(sort, map[string]interface{}{"price": "asc"})
	case "price_desc":
		sort = append(sort, map[string]interface{}{"price": "desc"})
	default:
		//默认 ：相关性 + 销量
		sort = append(sort, map[string]interface{}{"_score": "desc"})
		sort = append(sort, map[string]interface{}{"sales": "desc"})
	}

	query["sort"] = sort

	//执行搜索
	results, total, err := s.esClient.Search(ctx, ProductIndexName, query)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索失败: %w", err)
	}
	return results, total, nil
}

// GetSearchSuggestions 获取搜索建议
func (s *searchRepository) GetSearchSuggestions(ctx context.Context, keyword string, limit int) ([]string, error) {
	//从redis获取搜索建议
	key := "search:suggestions:" + keyword
	suggestions, err := s.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,                // 起始索引（包含）
		Stop:  int64(limit - 1), // 结束索引（包含）
		Rev:   true,             // 🔑 关键：启用反向排序，等价于 ZRevRange
	}).Result()
	if err != nil {
		return []string{}, nil
	}
	return suggestions, nil
}

// GetHotKeywords 获取搜索热词
func (s *searchRepository) GetHotKeywords(ctx context.Context, limit int) ([]string, error) {
	//从redis获取搜索热词
	key := "search:hot:keywords"
	keywords, err := s.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  int64(limit - 1),
		Rev:   true,
	}).Result()
	if err != nil {
		return []string{}, nil
	}
	return keywords, nil
}

// BuildProductIndex 构建商品索引
func (s *searchRepository) BuildProductIndex(ctx context.Context, productIDs []uint64) error {
	return nil
}

// NewSearchRepository 创建搜索仓库
func NewSearchRepository(redis *redis.Client, esClient *search.Client) SearchRepository {
	return &searchRepository{
		redis:    redis,
		esClient: esClient,
	}
}
