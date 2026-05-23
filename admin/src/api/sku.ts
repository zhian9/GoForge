import request from '@/utils/request'

export interface Sku {
  id: number
  product_id: number
  sku_code: string
  name: string
  specs: Record<string, string> // 规格属性，如 {"版本": "12GB+256GB", "颜色": "丹宁色"}
  price: number
  original_price?: number
  stock: number
  image?: string
  weight?: number
  volume?: number
  status: number // 0-下架, 1-上架
  created_at?: string
  updated_at?: string
}

export interface SkuListResponse {
  code: number
  message: string
  data: {
    list: Sku[]
    total: number
    page?: number
    page_size?: number
    total_pages?: number
  }
}

// 根据商品ID获取SKU列表（用于用户端）
export const getSkusByProductId = (productId: number) => {
  return request.get<SkuListResponse>('/v1/skus', {
    params: {
      product_id: productId,
      status: 1, // 只获取上架的SKU
    },
  })
}

export interface SkuDetailResponse {
  code: number
  message: string
  data: Sku
}

// 获取SKU列表
export const getSkuList = (params?: {
  product_id?: number
  status?: number // -1-全部, 0-下架, 1-上架
  page?: number
  page_size?: number
}) => {
  return request.get<SkuListResponse>('/v1/skus', { params })
}

// 获取SKU详情
export const getSkuDetail = (id: number) => {
  return request.get<SkuDetailResponse>(`/v1/skus/${id}`)
}

// 创建SKU
export interface CreateSkuRequest {
  product_id: number
  sku_code: string
  name: string
  specs: Record<string, string>
  price: number
  original_price?: number
  stock?: number
  image?: string
  weight?: number
  volume?: number
  status?: number // 0-下架, 1-上架
}

export const createSku = (data: CreateSkuRequest) => {
  return request.post<{ code: number; message: string; data: Sku }>('/v1/skus', data)
}

// 更新SKU
export interface UpdateSkuRequest extends Partial<CreateSkuRequest> {
  id: number
}

export const updateSku = (id: number, data: Partial<CreateSkuRequest>) => {
  return request.put<{ code: number; message: string; data: Sku }>(`/v1/skus/${id}`, data)
}

// 删除SKU
export const deleteSku = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/skus/${id}`)
}

