package job

import (
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhian9/GoForge/server/internal/service/job/repository"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config     Config
	DB         *gorm.DB
	Redis      *redis.Client
	OrderRepo  repository.OrderRepository
	CouponRepo repository.CouponRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	var db *gorm.DB
	var err error

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

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
	}

	// 初始化 Redis：用于「超时取消秒杀订单」时释放闸门配额。
	// 连接失败不阻断启动——真源库存的回补不依赖 Redis，
	// 只是秒杀配额释放会降级，日志里会明确记录。
	if c.BizRedis.Host != "" {
		redisClient, err := cache.NewRedis(&cache.Config{
			Host:         c.BizRedis.Host,
			Port:         c.BizRedis.Port,
			Password:     c.BizRedis.Password,
			Database:     c.BizRedis.Database,
			PoolSize:     c.BizRedis.PoolSize,
			MinIdleConns: c.BizRedis.MinIdleConns,
		})
		if err != nil {
			logx.Errorf("初始化 Redis 连接失败（秒杀配额释放将降级）: %v", err)
		} else {
			ctx.Redis = redisClient
		}
	} else {
		logx.Info("未配置 BizRedis，跳过 Redis 初始化（秒杀配额释放将降级）")
	}

	if db != nil {
		ctx.OrderRepo = repository.NewOrderRepository(db)
		ctx.CouponRepo = repository.NewCouponRepository(db)
	}

	return ctx
}
