package inventory

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/inventory/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config           Config
	DB               *gorm.DB
	Redis            *redis.Client
	Cache            *cache.CacheOperations
	MQProducer       *mq.Producer
	InventoryRepo    repository.InventoryRepository
	InventoryLogRepo repository.InventoryLogRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	var db *gorm.DB
	var redisClient *redis.Client
	var err error

	// 初始化数据库连接
	db, err = database.NewMySQL(&database.Config{
		Host:            c.Database.Host,
		Port:            c.Database.Port,
		User:            c.Database.User,
		Password:        c.Database.Password,
		Database:        c.Database.Database,
		Charset:         c.Database.Charset,
		MaxOpenConns:    c.Database.MaxOpenConns,
		MaxIdleConns:    c.Database.MaxIdleConns,
		ConnMaxLifetime: c.Database.ConnMaxLifetime,
		ConnMaxIdleTime: c.Database.ConnMaxIdleTime,
	})
	if err != nil {
		logx.Errorf("初始化数据库连接失败: %v", err)
	}

	// 初始化Redis连接
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
		DB:     db,
		Redis:  redisClient,
	}

	// 初始化缓存操作（仅在Redis连接成功时）
	if redisClient != nil {
		ctx.Cache = cache.NewCacheOperations(redisClient)
	}

	// 初始化Kafka生产者（可选）
	if c.Kafka != nil && len(c.Kafka.Brokers) > 0 {
		mqProducer, err := mq.NewProducer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ProducerAsync: true,
		})
		if err != nil {
			logx.Errorf("初始化Kafka生产者失败: %v", err)
		} else {
			ctx.MQProducer = mqProducer
		}
	}

	// 初始化Repository（仅在数据库连接成功时）
	if db != nil {
		ctx.InventoryRepo = repository.NewInventoryRepository(db)
		ctx.InventoryLogRepo = repository.NewInventoryLogRepository(db)
	}

	return ctx
}
