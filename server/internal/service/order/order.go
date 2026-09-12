package order

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/api/order/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/service/order/repository"
	"github.com/zhian9/GoForge/server/internal/service/order/service"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	DB            *gorm.DB
	Redis         *redis.Client
	Cache         *cache.CacheOperations
	MQProducer    *mq.Producer
	OrderRepo     repository.OrderRepository
	OrderItemRepo repository.OrderItemRepository
	OrderLogRepo  repository.OrderLogRepository
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
		ctx.OrderRepo = repository.NewOrderRepository(db)
		ctx.OrderItemRepo = repository.NewOrderItemRepository(db)
		ctx.OrderLogRepo = repository.NewOrderLogRepository(db)
	}

	return ctx
}

// OrderService 订单服务
// 注意：实现 gRPC 接口的代码在 order_service.go 中
type OrderService struct {
	v1.UnimplementedOrderServiceServer
	svcCtx *ServiceContext
	logic  *service.OrderLogic
}

// NewOrderService 创建订单服务
func NewOrderService(svcCtx *ServiceContext) *OrderService {
	// 创建业务逻辑层
	var orderLogic *service.OrderLogic
	if svcCtx.OrderRepo != nil && svcCtx.OrderItemRepo != nil && svcCtx.OrderLogRepo != nil {
		orderLogic = service.NewOrderLogic(
			svcCtx.OrderRepo,
			svcCtx.OrderItemRepo,
			svcCtx.OrderLogRepo,
			svcCtx.Cache,
			svcCtx.MQProducer,
			svcCtx.DB,
		)
	}

	return &OrderService{
		svcCtx: svcCtx,
		logic:  orderLogic,
	}
}
