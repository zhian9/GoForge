package job

import (
	"github.com/zhian9/GoForge/server/internal/service/job/repository"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/database"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config     Config
	DB         *gorm.DB
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

	if db != nil {
		ctx.OrderRepo = repository.NewOrderRepository(db)
		ctx.CouponRepo = repository.NewCouponRepository(db)
	}

	return ctx
}
