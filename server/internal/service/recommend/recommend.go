package recommend

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/service/recommend/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	Redis         *redis.Client
	RecommendRepo repository.RecommendRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	var redisClient *redis.Client
	var err error

	redisClient, err = cache.NewRedis(&cache.Config{
		Host:         c.BizRedis.Host,
		Port:         c.BizRedis.Port,
		Password:     c.BizRedis.Password,
		Database:     c.BizRedis.Database,
		PoolSize:     c.BizRedis.PoolSize,
		MinIdleConns: c.BizRedis.MinIdleConns,
	})
	if err != nil {
		logx.Errorf("初始化Redis连接失败: %v", err)
	}

	ctx := &ServiceContext{
		Config: c,
		Redis:  redisClient,
	}

	ctx.RecommendRepo = repository.NewRecommendRepository(redisClient)

	return ctx
}
