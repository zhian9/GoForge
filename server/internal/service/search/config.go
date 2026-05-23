package search

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 搜索服务配置
type Config struct {
	zrpc.RpcServerConf
	Elasticsearch ElasticsearchConfig
	BizRedis      RedisConfig // 业务侧使用的 Redis 配置，避免与 zrpc.RpcServerConf 内置的 Redis 字段冲突
	Database      DatabaseConfig
	Kafka         KafkaConfig
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses []string
	Username  string
	Password  string
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

// DatabaseConfig 数据库配置（用于读取商品快照写入 ES）
type DatabaseConfig struct {
	Driver          string
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

// KafkaConfig Kafka 配置（用于消费 Outbox 投递的同步消息）
type KafkaConfig struct {
	Brokers       []string
	Version       string
	ConsumerGroup string
}
