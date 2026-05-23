package file

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 文件服务配置
type Config struct {
	zrpc.RpcServerConf
	Storage StorageConfig
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type      string // local, oss, s3
	LocalPath string
	OSS       OSSConfig
}

// OSSConfig OSS配置
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}
