package search

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

// Config Elasticsearch配置
type Config struct {
	Addresses []string `json:",required"`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
}

// Client ES客户端
type Client struct {
	es *elasticsearch.Client
}

func NewElasticsearchClient(cfg *Config) (*Client, error) {
	esCfg := elasticsearch.Config{Addresses: cfg.Addresses}

	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建Elasticsearch客户端失败: %w", err)
	}

	//测试连接
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("连接Elasticsearch失败: %w", err)
	}
	defer res.Body.Close()
	client := &Client{es: es}
	logx.Infow("Elasticsearch连接成功")

	return client, nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, indexName string, mapping string) error {
	exists, err := c.es.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}
	defer exists.Body.Close()

	if exists.StatusCode == 200 {
		logx.Infof("索引 %s 已存在", indexName)
		return nil
	}

	//创建索引
	res, err := c.es.Indices.Create(indexName, c.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引失败: %s", res.String())
	}

	logx.Infof("索引 %s 创建成功", indexName)
	return nil
}

// IndexDocument 索引文档
func (c *Client) IndexDocument(ctx context.Context, indexName string, docID string, document interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %w", err)
	}

	req := esapi.IndexRequest{
		DocumentID: docID,
		Index:      indexName,
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文档失败: %s", res.String())
	}

	return nil
}

// Search 搜索文档
func (c *Client) Search(ctx context.Context, indexName string, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化查询失败: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(indexName),
		c.es.Search.WithBody(strings.NewReader(string(queryBody))),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("搜索失败: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	//解析结果
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return []map[string]interface{}{}, 0, nil
	}

	total, _ := hits["total"].(map[string]interface{})
	totalValue, _ := total["value"].(float64)

	hitsArray, _ := hits["hits"].([]interface{})
	documents := make([]map[string]interface{}, 0, len(hitsArray))

	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		source, ok := hitMap["_source"].(map[string]interface{})
		if ok {
			// 把相关性得分一并带出：_score 在 _source 之外，
			// 之前没取，导致上层拿到的 score 永远是 0。
			if score, ok := hitMap["_score"].(float64); ok {
				source["score"] = score
			}
			documents = append(documents, source)
		}
	}

	return documents, int64(totalValue), nil
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, indexName string, docID string) error {
	res, err := c.es.Delete(indexName, docID)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("删除文档失败: %s", res.String())
	}

	return nil
}

// BulkIndex 批量索引文档
func (c *Client) BulkIndex(ctx context.Context, indexName string, documents []map[string]interface{}) error {
	var bulkBody strings.Builder

	for _, doc := range documents {
		docID, ok := doc["id"].(string)
		if !ok {
			continue
		}

		// 构建批量操作
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    docID,
			},
		}

		actionJSON, _ := json.Marshal(action)
		bulkBody.WriteString(string(actionJSON))
		bulkBody.WriteString("\n")

		docJSON, _ := json.Marshal(doc)
		bulkBody.WriteString(string(docJSON))
		bulkBody.WriteString("\n")
	}

	res, err := c.es.Bulk(strings.NewReader(bulkBody.String()))
	if err != nil {
		return fmt.Errorf("批量索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量索引失败: %s", res.String())
	}

	return nil
}
