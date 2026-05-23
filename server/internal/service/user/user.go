package user

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	v1 "github.com/zhian9/GoForge/server/api/user/v1"
	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/service/user/repository"
	userservice "github.com/zhian9/GoForge/server/internal/service/user/service"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config         Config
	DB             *gorm.DB
	Redis          *redis.Client
	Cache          *cache.CacheOperations
	UserRepo       repository.UserRepository
	CredentialRepo repository.CredentialRepository
	AddressRepo    repository.AddressRepository
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
		// 数据库连接失败时，db 为 nil，后续 Repository 初始化会跳过
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
		// Redis连接失败时，redisClient 为 nil
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

	// 初始化Repository（仅在数据库连接成功时）
	if db != nil {
		ctx.UserRepo = repository.NewUserRepository(db)
		ctx.CredentialRepo = repository.NewCredentialRepository(db)
		ctx.AddressRepo = repository.NewAddressRepository(db)
	}

	return ctx
}

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServiceServer
	svcCtx       *ServiceContext
	logic        *userservice.UserLogic
	addressLogic *userservice.AddressLogic
}

// NewUserService 创建用户服务
func NewUserService(svcCtx *ServiceContext) *UserService {
	// 创建业务逻辑层
	userLogic := userservice.NewUserLogic(
		svcCtx.UserRepo,
		svcCtx.CredentialRepo,
		svcCtx.AddressRepo,
		svcCtx.Cache,
	)

	// 创建地址业务逻辑层
	addressLogic := userservice.NewAddressLogic(svcCtx.AddressRepo, svcCtx.Cache)

	return &UserService{
		svcCtx:       svcCtx,
		logic:        userLogic,
		addressLogic: addressLogic,
	}
}
