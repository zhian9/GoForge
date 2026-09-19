import request from '@/utils/request'

/**
 * 商品搜索（走 Elasticsearch 全文检索，不是 MySQL LIKE）
 *
 * 后端链路：网关 /api/v1/search/products → search-service → ES 索引 products_v1
 * 索引数据由「商品变更 → outbox → Kafka → search 服务消费」同步，
 * 服务启动时还会做一次全量重建兜底。
 */

/** ES 返回的搜索结果项（注意后端 int64 会序列化成字符串） */
export interface SearchProductItem {
  productId: string
  name: string
  mainImage?: string
  price: number | string
  sales?: number
  score?: number
}

export interface SearchProductsParams {
  keyword: string
  page?: number
  pageSize?: number
  categoryId?: number
  /** sales / price_asc / price_desc，留空按相关性排序 */
  sortBy?: string
}

export interface SearchProductsResult {
  code: number
  message: string
  data: SearchProductItem[]
  total: number | string
}

export function searchProducts(params: SearchProductsParams) {
  return request.get<SearchProductsResult>('/v1/search/products', {
    params: {
      keyword: params.keyword,
      page: params.page ?? 1,
      page_size: params.pageSize ?? 12,
      category_id: params.categoryId ?? 0,
      sort_by: params.sortBy ?? '',
    },
  }) as unknown as Promise<SearchProductsResult>
}
