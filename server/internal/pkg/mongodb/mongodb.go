package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config MongoDB配置
type Config struct {
	URI            string `json:",required"`
	Database       string `json:",required"`
	MaxPoolSize    uint64 `json:",default=100"`
	MinPoolSize    uint64 `json:",default=10"`
	ConnectTimeout int    `json:",default=10"` // 秒
}

// Client MongoDB客户端
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDBClient 创建MongoDB客户端
func NewMongoDBClient(cfg *Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	db := client.Database(cfg.Database)

	logx.Infow("MongoDB连接成功", logx.Field("database", cfg.Database))

	return &Client{
		client:   client,
		database: db,
	}, nil
}

// Database 获取数据库实例
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection 获取集合
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close 关闭连接
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, collectionName string, indexModel mongo.IndexModel) error {
	collection := c.Collection(collectionName)
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// CreateIndexes 批量创建索引
func (c *Client) CreateIndexes(ctx context.Context, collectionName string, indexModels []mongo.IndexModel) error {
	collection := c.Collection(collectionName)
	_, err := collection.Indexes().CreateMany(ctx, indexModels)
	return err
}
