package product

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/api/product/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/outbox"
	"github.com/zhian9/GoForge/server/internal/service/product/repository"
	"github.com/zhian9/GoForge/server/internal/service/product/service"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config       Config
	DB           *gorm.DB
	Redis        *redis.Client
	Cache        *cache.CacheOperations
	MQProducer   *mq.Producer
	OutboxRepo   *outbox.Repo
	ProductRepo  repository.ProductRepository
	CategoryRepo repository.CategoryRepository
	SkuRepo      repository.SkuRepository
	BannerRepo   repository.BannerRepository
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

	// 初始化Kafka生产者
	if len(c.Kafka.Brokers) > 0 {
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
		ctx.ProductRepo = repository.NewProductRepository(db)
		ctx.CategoryRepo = repository.NewCategoryRepository(db)
		ctx.SkuRepo = repository.NewSkuRepository(db)
		ctx.BannerRepo = repository.NewBannerRepository(db)
		ctx.OutboxRepo = outbox.NewRepo(db)
	}

	// 启动 Outbox Relay：轮询 outbox_event 并投递到 Kafka（TopicDataSync）
	// 注意：投递失败不会影响主业务写库，只会导致 ES 最终一致延迟
	if ctx.OutboxRepo != nil && ctx.MQProducer != nil {
		relay := outbox.NewRelay(ctx.OutboxRepo, ctx.MQProducer, outbox.RelayConfig{})
		go relay.Start(context.Background())
	}

	return ctx
}

// ProductService 商品服务
type ProductService struct {
	v1.UnimplementedProductServiceServer
	svcCtx *ServiceContext
	logic  *service.ProductLogic
}

// NewProductService 创建商品服务
func NewProductService(svcCtx *ServiceContext) *ProductService {
	// 创建业务逻辑层
	productLogic := service.NewProductLogic(
		svcCtx.DB,
		svcCtx.OutboxRepo,
		svcCtx.ProductRepo,
		svcCtx.CategoryRepo,
		svcCtx.SkuRepo,
		svcCtx.BannerRepo,
		svcCtx.Cache,
		svcCtx.MQProducer,
	)

	return &ProductService{
		svcCtx: svcCtx,
		logic:  productLogic,
	}
}
