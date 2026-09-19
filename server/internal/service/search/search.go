package search

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	"github.com/zhian9/GoForge/server/internal/pkg/database"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/search"
	"github.com/zhian9/GoForge/server/internal/service/search/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config       Config
	Redis        *redis.Client
	ESClient     *search.Client
	DB           *gorm.DB
	MQConsumer   *mq.Consumer
	SearchRepo   repository.SearchRepository
	SnapshotRepo repository.ProductSnapshotRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	var redisClient *redis.Client
	var err error

	// 初始化Redis连接（用于缓存搜索热词等）
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

	// 初始化Elasticsearch客户端
	if len(c.Elasticsearch.Addresses) > 0 {
		esClient, err := search.NewElasticsearchClient(&search.Config{
			Addresses: c.Elasticsearch.Addresses,
			Username:  c.Elasticsearch.Username,
			Password:  c.Elasticsearch.Password,
		})
		if err != nil {
			logx.Errorf("初始化Elasticsearch客户端失败: %v", err)
		} else {
			ctx.ESClient = esClient
		}
	}

	// 初始化数据库连接（读取商品快照）
	var db *gorm.DB
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
	} else {
		ctx.DB = db
	}

	// 初始化Repository
	ctx.SearchRepo = repository.NewSearchRepository(redisClient, ctx.ESClient)
	if ctx.DB != nil {
		ctx.SnapshotRepo = repository.NewProductSnapshotRepository(ctx.DB)
	}

	// 确保 ES 索引存在（用于全文检索）
	if ctx.ESClient != nil {
		_ = ctx.ESClient.CreateIndex(context.Background(), repository.ProductIndexName, repository.ProductIndexMapping)
	}

	// 初始化 Kafka 消费者（用于消费 outbox relay 投递的同步消息）
	if len(c.Kafka.Brokers) > 0 && c.Kafka.ConsumerGroup != "" {
		consumer, err := mq.NewConsumer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ConsumerGroup: c.Kafka.ConsumerGroup,
		})
		if err != nil {
			logx.Errorf("初始化Kafka消费者失败: %v", err)
		} else {
			ctx.MQConsumer = consumer
			// 注册处理器：只关心商品相关事件
			ctx.MQConsumer.RegisterHandler(mq.TopicDataSync, func(cctx context.Context, msg *mq.Message) error {
				return ctx.handleDataSyncMessage(cctx, msg)
			})
			// 启动消费循环
			go func() {
				_ = ctx.MQConsumer.Start(context.Background(), []string{mq.TopicDataSync})
			}()
		}
	}

	// 启动时做一次全量索引重建（异步，不阻塞启动）。
	//
	// 为什么必须有这一步：增量同步依赖 outbox 事件，而 outbox 事件只在
	// 「通过接口变更商品」时才产生。像 seed.sql 直接 INSERT 的商品永远不会被索引，
	// 结果是索引长期不完整（实测：MySQL 11 个商品，ES 里只有 3 个）。
	go func() {
		if err := ctx.ReindexAllProducts(context.Background()); err != nil {
			logx.Errorf("启动时全量重建索引失败: %v", err)
		}
	}()

	return ctx
}

// ReindexAllProducts 把 MySQL 里全部在售商品重新写入 ES。
//
// 索引是按 product_id 覆盖写，所以重复执行幂等，不会产生重复文档。
func (s *ServiceContext) ReindexAllProducts(ctx context.Context) error {
	if s.ESClient == nil || s.SnapshotRepo == nil {
		logx.Info("跳过全量重建：ES 客户端或商品快照仓储未初始化")
		return nil
	}

	ids, err := s.SnapshotRepo.ListAllProductIDs(ctx)
	if err != nil {
		return err
	}

	ok, failed := 0, 0
	for _, id := range ids {
		doc, err := s.SnapshotRepo.BuildProductDocument(ctx, id)
		if err != nil {
			logx.Errorf("构建商品文档失败 product_id=%d: %v", id, err)
			failed++
			continue
		}
		if err := s.ESClient.IndexDocument(ctx, repository.ProductIndexName,
			strconv.FormatUint(id, 10), doc); err != nil {
			logx.Errorf("写入 ES 失败 product_id=%d: %v", id, err)
			failed++
			continue
		}
		ok++
	}

	logx.Infof("全量重建索引完成: 成功 %d, 失败 %d, 合计 %d", ok, failed, len(ids))
	return nil
}
