package review

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/pkg/mongodb"
	"github.com/zhian9/GoForge/server/internal/service/review/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config          Config
	DB              *gorm.DB
	MongoDB         *mongodb.Client
	ReviewRepo      repository.ReviewRepository
	ReviewReplyRepo repository.ReviewReplyRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	var db *gorm.DB
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

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
	}

	// 初始化MongoDB客户端（用于存储评价详情，包含图片、视频）
	if c.MongoDB != nil && c.MongoDB.URI != "" && c.MongoDB.Database != "" {
		mongoClient, err := mongodb.NewMongoDBClient(&mongodb.Config{
			URI:            c.MongoDB.URI,
			Database:       c.MongoDB.Database,
			MaxPoolSize:    100,
			MinPoolSize:    10,
			ConnectTimeout: 10,
		})
		if err != nil {
			logx.Errorf("初始化MongoDB客户端失败: %v", err)
		} else {
			ctx.MongoDB = mongoClient
			// 初始化评价集合索引
			_ = initReviewIndexes(context.Background(), mongoClient)
		}
	}

	// 初始化Repository
	if db != nil {
		ctx.ReviewRepo = repository.NewReviewRepository(db, ctx.MongoDB)
		ctx.ReviewReplyRepo = repository.NewReviewReplyRepository(db)
	}

	return ctx
}

// initReviewIndexes 初始化评价集合索引
func initReviewIndexes(ctx context.Context, mongoClient *mongodb.Client) error {
	if mongoClient == nil {
		return nil
	}

	// 创建评价集合索引
	indexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"product_id": 1,
				"status":     1,
				"created_at": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"user_id":    1,
				"created_at": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"rating":     1,
				"created_at": -1,
			},
		},
	}

	collection := mongoClient.Collection("reviews")
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
