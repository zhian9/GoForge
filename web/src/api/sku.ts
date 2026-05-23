import request from '@/utils/request'

export interface Sku {
  id: number
  product_id: number
  sku_code: string
  name: string
  specs: Record<string, string> // 规格属性
  price: number
  original_price?: number
  stock: number
  image?: string
  weight?: number
  volume?: number
  status: number
}

export interface SkuListResponse {
  code: number
  message: string
  data: Sku[] | {
    list: Sku[]
    total?: number
    page?: number
    page_size?: number
    total_pages?: number
  }
}

// 根据商品ID获取SKU列表
export const getSkusByProductId = (productId: number) => {
  return request.get<SkuListResponse>('/v1/skus', {
    params: {
      product_id: productId,
      status: 1, // 只获取上架的SKU
    },
  })
}

// 获取SKU详情
export const getSkuDetail = (id: number) => {
  return request.get<{ code: number; message: string; data: Sku }>(`/v1/skus/${id}`)
}

