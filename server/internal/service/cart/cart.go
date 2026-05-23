package cart

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"

	productpb "github.com/zhian9/GoForge/server/api/product/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/service/cart/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config      Config
	DB          *gorm.DB
	Redis       *redis.Client
	CartRepo    repository.CartRepository
	ProductRpc  productpb.ProductServiceClient
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

	// 初始化Redis连接（购物车主要存储在Redis）
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

	// 初始化商品服务 gRPC 客户端（用于查询 SKU 价格）
	if c.ProductRpc.Target != "" {
		conn, err := zrpc.NewClient(c.ProductRpc)
		if err != nil {
			logx.Errorf("初始化商品服务客户端失败: %v", err)
		} else {
			ctx.ProductRpc = productpb.NewProductServiceClient(conn.Conn())
		}
	}

	// 初始化Repository
	if db != nil {
		ctx.CartRepo = repository.NewCartRepository(db, redisClient)
	}

	return ctx
}
