package payment

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 支付服务配置
type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	BizRedis RedisConfig // 业务侧使用的 Redis 配置，避免与 zrpc.RpcServerConf 内置的 Redis 字段冲突
	Payment  PaymentConfig
}

// DatabaseConfig 数据库配置
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
	ConnMaxLifetime int
	ConnMaxIdleTime int
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

// PaymentConfig 支付配置
type PaymentConfig struct {
	WeChat WeChatConfig
	Alipay AlipayConfig
}

// WeChatConfig 微信支付配置
type WeChatConfig struct {
	AppID     string
	MchID     string
	APIKey    string
	NotifyURL string
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string
	PrivateKey string
	PublicKey  string
	NotifyURL  string
}
