package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheOperations 缓存操作封装
type CacheOperations struct {
	client *redis.Client
}

// NewCacheOperations 创建缓存操作实例
func NewCacheOperations(client *redis.Client) *CacheOperations {
	return &CacheOperations{client: client}
}

// Set 设置缓存（字符串）
func (c *CacheOperations) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		val = string(data)
	}
	return c.client.Set(ctx, key, val, expiration).Err()
}

// Get 获取缓存（字符串）
func (c *CacheOperations) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// GetJSON 获取缓存并反序列化为JSON
func (c *CacheOperations) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete 删除缓存
func (c *CacheOperations) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *CacheOperations) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Expire 设置过期时间
func (c *CacheOperations) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *CacheOperations) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Increment 递增
func (c *CacheOperations) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrementBy 按指定值递增
func (c *CacheOperations) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decrement 递减
func (c *CacheOperations) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrementBy 按指定值递减
func (c *CacheOperations) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// HashSet 设置Hash字段
func (c *CacheOperations) HashSet(ctx context.Context, key string, values map[string]interface{}) error {
	return c.client.HSet(ctx, key, values).Err()
}

// HashGet 获取Hash字段
func (c *CacheOperations) HashGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HashGetAll 获取Hash所有字段
func (c *CacheOperations) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HashDelete 删除Hash字段
func (c *CacheOperations) HashDelete(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HashExists 检查Hash字段是否存在
func (c *CacheOperations) HashExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// ListPush 从左侧推入列表
func (c *CacheOperations) ListPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// ListPop 从左侧弹出列表
func (c *CacheOperations) ListPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// ListRange 获取列表范围
func (c *CacheOperations) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// ListLength 获取列表长度
func (c *CacheOperations) ListLength(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// SetAdd 添加Set成员
func (c *CacheOperations) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SetRemove 移除Set成员
func (c *CacheOperations) SetRemove(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SetMembers 获取Set所有成员
func (c *CacheOperations) SetMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SetIsMember 检查是否为Set成员
func (c *CacheOperations) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// DeletePattern 删除匹配模式的所有键（使用 SCAN 命令，避免阻塞）
func (c *CacheOperations) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keysToDelete []string
	batchSize := 100 // 每批删除100个键

	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, pattern, int64(batchSize)).Result()
		if err != nil {
			return err
		}

		keysToDelete = append(keysToDelete, keys...)

		// 如果 cursor 为 0，说明已经扫描完所有键
		if cursor == 0 {
			break
		}

		// 如果累积的键数量达到批次大小，批量删除
		if len(keysToDelete) >= batchSize {
			if len(keysToDelete) > 0 {
				if err := c.client.Del(ctx, keysToDelete...).Err(); err != nil {
					return err
				}
				keysToDelete = keysToDelete[:0] // 清空切片
			}
		}
	}

	// 删除剩余的键
	if len(keysToDelete) > 0 {
		return c.client.Del(ctx, keysToDelete...).Err()
	}

	return nil
}
