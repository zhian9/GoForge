package order

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // 秒
	ConnMaxIdleTime int // 秒
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

// Config 订单服务配置
type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	BizRedis RedisConfig // 业务侧使用的 Redis 配置，避免与 zrpc.RpcServerConf 内置的 Redis 字段冲突
	Kafka    KafkaConfig
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string
	Version string
}
