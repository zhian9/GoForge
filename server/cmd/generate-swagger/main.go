// Command generate-swagger 生成网关的 Swagger/OpenAPI 文档。
//
// 本项目的 proto 没有 google.api.http 注解，HTTP 路由写在 configs/*/gateway.yaml 的
// Upstreams.Mappings 里，直接用 protoc-gen-openapiv2 生成只会得到空的 paths。
// 因此这里以网关配置为准，为每个上游服务生成一份 Swagger 2.0 文档，
// 供网关内嵌的 Swagger UI（:8095）加载。
//
// 用法（在 server/ 目录下执行）：
//
//	go run ./cmd/generate-swagger
//	go run ./cmd/generate-swagger -config ../configs/docker/gateway.yaml -host api.example.com
//
// 输出：docs/swagger/api/<service>/v1/<service>.swagger.json
//
// 说明：请求体这里只给出通用的 object 描述（字段级 schema 需要 proto 注解）；
// 接口清单、路径参数、Try it out 都已经可用。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	configPath = flag.String("config", "../configs/dev/gateway.yaml", "网关配置路径")
	outputDir  = flag.String("out", "docs/swagger", "输出目录")
	apiHost    = flag.String("host", "localhost:8080", "文档里的 API host（Swagger UI 的 Try it out 会请求它）")
	schemesArg = flag.String("schemes", "http", "协议，逗号分隔")
)

// GatewayConfig 只关心生成文档需要的字段
type GatewayConfig struct {
	Upstreams []Upstream `yaml:"Upstreams"`
}

// Upstream 一个上游服务
type Upstream struct {
	Name     string    `yaml:"Name"`
	Mappings []Mapping `yaml:"Mappings"`
}

// Mapping HTTP 路径到 gRPC 方法的映射
type Mapping struct {
	Method  string `yaml:"Method"`
	Path    string `yaml:"Path"`
	RpcPath string `yaml:"RpcPath"`
}

// SwaggerDoc Swagger 2.0 文档结构（只声明用到的字段）
type SwaggerDoc struct {
	Swagger     string                            `json:"swagger"`
	Info        map[string]interface{}            `json:"info"`
	Host        string                            `json:"host,omitempty"`
	BasePath    string                            `json:"basePath,omitempty"`
	Schemes     []string                          `json:"schemes,omitempty"`
	Tags        []map[string]interface{}          `json:"tags"`
	Consumes    []string                          `json:"consumes"`
	Produces    []string                          `json:"produces"`
	Paths       map[string]map[string]interface{} `json:"paths"`
	Definitions map[string]interface{}            `json:"definitions"`
}

// specEntry 供 Swagger UI 加载的清单项
type specEntry struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// serviceTitles 服务名 -> 文档里显示的中文名（没有的话直接用服务名）
var serviceTitles = map[string]string{
	"user":      "用户服务",
	"product":   "商品服务",
	"order":     "订单服务",
	"payment":   "支付服务",
	"cart":      "购物车服务",
	"inventory": "库存服务",
	"promotion": "营销服务",
	"review":    "评价服务",
	"logistics": "物流服务",
	"message":   "消息服务",
	"seckill":   "秒杀服务",
	"search":    "搜索服务",
	"recommend": "推荐服务",
	"file":      "文件服务",
	"job":       "任务服务",
}

func main() {
	flag.Parse()

	cfg, err := loadGatewayConfig(*configPath)
	if err != nil {
		log.Fatalf("加载网关配置失败: %v", err)
	}

	schemes := splitAndTrim(*schemesArg)
	totalFiles, totalOps := 0, 0
	entries := make([]specEntry, 0, len(cfg.Upstreams))

	for _, upstream := range cfg.Upstreams {
		short := shortServiceName(upstream.Name)
		doc, ops := buildDoc(short, upstream, schemes)
		if ops == 0 {
			// 该上游在网关配置里没有任何 HTTP 映射（也就不可能被访问到），不产出文档
			fmt.Printf("⚠️  跳过 %s：网关配置里没有可用的路由映射\n", upstream.Name)
			continue
		}

		relPath := filepath.Join("api", short, "v1", short+".swagger.json")
		outPath := filepath.Join(*outputDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			log.Fatalf("创建目录失败: %v", err)
		}

		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			log.Fatalf("序列化 %s 失败: %v", relPath, err)
		}
		if err := os.WriteFile(outPath, append(data, '\n'), 0o644); err != nil {
			log.Fatalf("写入 %s 失败: %v", outPath, err)
		}

		fmt.Printf("✅ %s（%d 个接口）\n", filepath.ToSlash(relPath), ops)
		entries = append(entries, specEntry{
			URL:  "/swagger/" + filepath.ToSlash(relPath),
			Name: displayName(short),
		})
		totalFiles++
		totalOps += ops
	}

	// 生成清单：Swagger UI 按这份清单加载，避免列出不存在的文档（打开就 404）
	indexData, err := json.MarshalIndent(map[string]interface{}{"specs": entries}, "", "  ")
	if err != nil {
		log.Fatalf("序列化 index.json 失败: %v", err)
	}
	indexPath := filepath.Join(*outputDir, "index.json")
	if err := os.WriteFile(indexPath, append(indexData, '\n'), 0o644); err != nil {
		log.Fatalf("写入 %s 失败: %v", indexPath, err)
	}

	fmt.Printf("\n✅ 已生成 %d 份文档，共 %d 个接口，输出目录: %s\n", totalFiles, totalOps, *outputDir)
	fmt.Printf("✅ 清单: %s（Swagger UI 用它决定加载哪些文档）\n", filepath.ToSlash(indexPath))
}

func loadGatewayConfig(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config GatewayConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if len(config.Upstreams) == 0 {
		return nil, fmt.Errorf("%s 里没有 Upstreams，确认路径是否正确", path)
	}

	return &config, nil
}

// buildDoc 依据某个上游的映射生成文档，返回文档与接口数量
func buildDoc(short string, upstream Upstream, schemes []string) (*SwaggerDoc, int) {
	doc := &SwaggerDoc{
		Swagger: "2.0",
		Info: map[string]interface{}{
			"title":       short + " service API",
			"version":     "1.0.0",
			"description": fmt.Sprintf("由 %s 生成（upstream: %s）", filepath.ToSlash(*configPath), upstream.Name),
		},
		Host:        *apiHost,
		Schemes:     schemes,
		Tags:        []map[string]interface{}{{"name": short}},
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       make(map[string]map[string]interface{}),
		Definitions: make(map[string]interface{}),
	}

	ops := 0
	for _, m := range upstream.Mappings {
		// OPTIONS 是给 CORS 预检用的，不需要出现在文档里
		if strings.EqualFold(m.Method, "options") {
			continue
		}

		path := convertPathParams(m.Path)
		method := strings.ToLower(strings.TrimSpace(m.Method))
		if path == "" || method == "" {
			continue
		}

		if doc.Paths[path] == nil {
			doc.Paths[path] = make(map[string]interface{})
		}
		doc.Paths[path][method] = buildOperation(short, method, m)
		ops++
	}

	// 统一引用一个通用的错误结构，避免 Swagger UI 报 unresolvable ref
	doc.Definitions["rpcStatus"] = map[string]interface{}{
		"type":        "object",
		"description": "gRPC 错误（code 非 0 时返回）",
		"properties": map[string]interface{}{
			"code":    map[string]interface{}{"type": "integer", "format": "int32"},
			"message": map[string]interface{}{"type": "string"},
		},
	}

	return doc, ops
}

func buildOperation(tag, method string, m Mapping) map[string]interface{} {
	rpcName := m.RpcPath
	if idx := strings.LastIndex(m.RpcPath, "/"); idx >= 0 && idx+1 < len(m.RpcPath) {
		rpcName = m.RpcPath[idx+1:]
	}

	op := map[string]interface{}{
		"tags":        []string{tag},
		"summary":     rpcName,
		"description": "gRPC 方法: " + m.RpcPath,
		"operationId": strings.ReplaceAll(m.RpcPath, "/", "_"),
		"produces":    []string{"application/json"},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{
				"description": "成功（响应体为 { code, message, data }）",
			},
			"default": map[string]interface{}{
				"description": "错误",
				"schema":      map[string]interface{}{"$ref": "#/definitions/rpcStatus"},
			},
		},
	}

	switch method {
	case "post", "put", "patch":
		op["consumes"] = []string{"application/json"}
		op["parameters"] = []map[string]interface{}{{
			"name":        "body",
			"in":          "body",
			"required":    true,
			"description": fmt.Sprintf("请求体（JSON，字段与 %s 的 proto message 一致）", m.RpcPath),
			"schema":      map[string]interface{}{"type": "object"},
		}}
	default:
		// GET/DELETE：把路径参数声明出来，Swagger UI 才好填
		if params := extractPathParams(convertPathParams(m.Path)); len(params) > 0 {
			op["parameters"] = params
		}
	}

	return op
}

func convertPathParams(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func extractPathParams(path string) []map[string]interface{} {
	var params []map[string]interface{}
	for _, part := range strings.Split(path, "/") {
		var name string
		switch {
		case strings.HasPrefix(part, ":"):
			name = strings.TrimPrefix(part, ":")
		case strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}"):
			name = strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
		}
		if name == "" {
			continue
		}
		params = append(params, map[string]interface{}{
			"name":        name,
			"in":          "path",
			"required":    true,
			"type":        "string",
			"description": name + " 参数",
		})
	}
	return params
}

// shortServiceName "cart-service" -> "cart"，用于拼文档路径 api/<short>/v1/<short>.swagger.json
func shortServiceName(name string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), "-service")
}

func splitAndTrim(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func displayName(short string) string {
	if title, ok := serviceTitles[short]; ok {
		return title
	}
	return short
}
