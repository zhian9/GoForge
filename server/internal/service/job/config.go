package job

import (
	"github.com/zeromicro/go-zero/zrpc"
)

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

// RedisConfig Redis 配置
//
// job-service 需要 Redis 是为了「超时取消秒杀订单」时把秒杀配额还回去：
// 秒杀订单占用的库存有两个地方——MySQL 的 sku.stock（真源）和 Redis 的
// seckill:stock:{skuId}（闸门）。只回补真源不释放闸门，配额就永久损失了。
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}

// CancelExpiredOrderConfig 超时未支付订单自动取消的配置
type CancelExpiredOrderConfig struct {
	Enable         bool // 是否启用定时取消
	IntervalSecond int  // 扫描间隔（秒）
	TimeoutMinutes int  // 超过多少分钟未支付即视为超时
	BatchLimit     int  // 单次最多处理多少笔
}

// Config 定时任务服务配置
type Config struct {
	zrpc.RpcServerConf
	Database           DatabaseConfig
	BizRedis           RedisConfig
	CancelExpiredOrder CancelExpiredOrderConfig
}
