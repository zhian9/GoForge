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
	Config     Config
	DB         *gorm.DB
	Redis      *redis.Client
	CartRepo   repository.CartRepository
	ProductRpc productpb.ProductServiceClient
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

	// 初始化商品服务 gRPC 客户端（用于补齐购物车条目的商品名 / 价格 / 图片）。
	//
	// 这里原来判断的是 `c.ProductRpc.Target != ""`，而部署配置用的是 Etcd 服务发现
	// （ProductRpc.Etcd.Hosts + Key），Target 始终为空 —— 客户端从未被创建，
	// resolveCartItemInfo 直接返回空信息，于是购物车和下单页的商品名称/价格/图片全是空的。
	// 现在按配置直接创建：配置缺失时 NewClient 会返回错误并记录日志，不影响服务启动。
	if conn, err := zrpc.NewClient(c.ProductRpc); err != nil {
		logx.Errorf("初始化商品服务客户端失败（购物车将无法补齐商品信息）: %v", err)
	} else {
		ctx.ProductRpc = productpb.NewProductServiceClient(conn.Conn())
	}

	// 初始化Repository
	if db != nil {
		ctx.CartRepo = repository.NewCartRepository(db, redisClient)
	}

	return ctx
}
