package recommend

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 推荐服务配置
type Config struct {
	zrpc.RpcServerConf
	BizRedis RedisConfig // 业务侧使用的 Redis 配置，避免与 zrpc.RpcServerConf 内置的 Redis 字段冲突
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}
