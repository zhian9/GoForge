package repository

// ProductIndexName ES 商品索引名
const ProductIndexName = "products_v1"

// ProductIndexMapping ES 索引 mapping（尽量使用内置 analyzer，避免依赖额外插件）
// 说明：
// - name/subtitle/detail 用 text 以支持全文检索
// - 保留 keyword 字段用于过滤/聚合
const ProductIndexMapping = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "product_id": { "type": "long" },
      "name": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "subtitle": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "detail": { "type": "text" },
      "category_id": { "type": "long" },
      "brand_id": { "type": "long" },
      "status": { "type": "integer" },
      "is_hot": { "type": "integer" },
      "sales": { "type": "integer" },
      "main_image": { "type": "keyword" },
      "price": { "type": "double" },
      "price_min": { "type": "double" },
      "price_max": { "type": "double" },
      "skus": {
        "type": "nested",
        "properties": {
          "sku_id": { "type": "long" },
          "sku_name": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
          "price": { "type": "double" },
          "stock": { "type": "integer" },
          "status": { "type": "integer" },
          "image": { "type": "keyword" },
          "specs": { "type": "object", "enabled": true }
        }
      },
      "updated_at": { "type": "date" }
    }
  }
}`
