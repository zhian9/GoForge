package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// DistributedLock 分布式锁
type DistributedLock struct {
	client     *redis.Client
	key        string
	value      string // 锁的唯一标识，用于安全释放
	expiration time.Duration
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(client *redis.Client, key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		client:     client,
		key:        key,
		value:      uuid.New().String(),
		expiration: expiration,
	}
}

// Lock 尝试获取锁
func (dl *DistributedLock) Lock(ctx context.Context) (bool, error) {
	// 使用SET NX EX命令原子性地设置锁
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Unlock 释放锁（使用Lua脚本保证原子性）
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// Lua脚本：只有当锁的值匹配时才删除
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	// 执行 Lua 脚本：传入 key 作为 KEYS[1]，value 作为 ARGV[1]
	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to execute unlock script: %w", err)
	}

	// 检查返回值：1 表示成功删除，0 表示锁不匹配或已过期
	if result == int64(1) {
		return nil
	}
	return fmt.Errorf("failed to release lock: lock not held or already expired")
}

// TryLock 尝试获取锁，带重试
func (dl *DistributedLock) TryLock(ctx context.Context, maxRetries int, retryInterval time.Duration) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		locked, err := dl.Lock(ctx)
		if err != nil {
			return false, err
		}
		if locked {
			return true, nil
		}
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}
	return false, nil
}

// LockWithTimeout 带超时的锁获取
func LockWithTimeout(ctx context.Context, client *redis.Client, key string, timeout time.Duration) (*DistributedLock, error) {
	lock := NewDistributedLock(client, key, timeout)
	locked, err := lock.Lock(ctx)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, fmt.Errorf("获取锁失败: %s", key)
	}
	return lock, nil
}
