package review

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 评价服务配置
type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	MongoDB  *MongoDBConfig
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI      string
	Database string
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
