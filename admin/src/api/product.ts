import request from '@/utils/request'

export interface Product {
  id: number
  name: string
  description: string
  price: number
  original_price: number
  stock: number
  category_id: number
  brand_id?: number
  main_image?: string
  local_main_image?: string
  images: string[]
  local_images?: string[]
  status: number
  is_hot?: number // 是否热门: 0-否, 1-是
  created_at: string
  updated_at: string
}

export interface ProductListResponse {
  code: number
  message: string
  data: {
    list: Product[]  // 后端返回的是 list，不是 products
    total: number
    page?: number
    page_size?: number
    total_pages?: number
  }
}

export interface ProductDetailResponse {
  code: number
  message: string
  data: Product
}

// 获取商品列表
export const getProductList = (params?: {
  page?: number
  page_size?: number
  category_id?: number
  keyword?: string
}) => {
  return request.get<ProductListResponse>('/v1/products', { params })
}

// 获取商品详情
export const getProductDetail = (id: number) => {
  return request.get<ProductDetailResponse>(`/v1/products/${id}`)
}

// 创建商品
export interface CreateProductRequest {
  name: string
  description?: string
  subtitle?: string
  price: number
  original_price?: number
  stock?: number
  category_id: number
  brand_id?: number
  main_image?: string
  local_main_image?: string
  images?: string[]
  local_images?: string[]
  detail?: string
  status?: number
  is_hot?: number // 是否热门: 0-否, 1-是
}

export const createProduct = (data: CreateProductRequest) => {
  return request.post<{ code: number; message: string; data: Product }>('/v1/products', data)
}

// 更新商品
export interface UpdateProductRequest extends Partial<CreateProductRequest> {
  id: number
}

export const updateProduct = (id: number, data: Partial<CreateProductRequest>) => {
  return request.put<{ code: number; message: string; data: Product }>(`/v1/products/${id}`, data)
}

// 删除商品
export const deleteProduct = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/products/${id}`)
}

