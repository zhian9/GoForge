import request from '@/utils/request'

export interface Product {
  id: number
  name: string
  description: string
  /** 商品详情（富文本），列表接口也会返回 */
  detail?: string
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

/**
 * 商品归一化。
 *
 * 网关返回 camelCase（mainImage / localMainImage / categoryId / originalPrice / isHot），
 * 而 Product 类型与各页面读的是 snake_case —— 不做转换的话，后台商品列表的
 * 「图片」列永远是空的，分类、原价、热销标记也读不到。
 */
const normalizeProduct = (dto: any): Product => ({
  id: Number(dto?.id ?? 0),
  name: dto?.name ?? '',
  description: dto?.description ?? '',
  price: Number(dto?.price ?? 0),
  original_price: Number(dto?.originalPrice ?? dto?.original_price ?? 0),
  stock: Number(dto?.stock ?? 0),
  category_id: Number(dto?.categoryId ?? dto?.category_id ?? 0),
  brand_id: dto?.brandId !== undefined ? Number(dto.brandId || 0) : dto?.brand_id,
  main_image: dto?.mainImage ?? dto?.main_image ?? '',
  local_main_image: dto?.localMainImage ?? dto?.local_main_image ?? '',
  images: Array.isArray(dto?.images) ? dto.images.slice() : [],
  local_images: Array.isArray(dto?.localImages) ? dto.localImages.slice() : dto?.local_images,
  detail: dto?.detail ?? '',
  status: Number(dto?.status ?? 0),
  is_hot: dto?.isHot !== undefined ? Number(dto.isHot) : dto?.is_hot,
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
  updated_at: dto?.updatedAt ?? dto?.updated_at ?? '',
})

// 获取商品列表
export const getProductList = (params?: {
  page?: number
  page_size?: number
  category_id?: number
  keyword?: string
  /** -1-全部, 0-下架, 1-上架（SKU 页的下拉框需要拉全量商品） */
  status?: number
}) => {
  return request.get<ProductListResponse>('/v1/products', { params }).then((res) => {
    const raw: any = res?.data || {}
    const list = Array.isArray(raw.list) ? raw.list.map(normalizeProduct) : []
    return { ...res, data: { ...raw, list } } as ProductListResponse
  })
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
export const updateProduct = (id: number, data: Partial<CreateProductRequest>) => {
  return request.put<{ code: number; message: string; data: Product }>(`/v1/products/${id}`, data)
}

// 删除商品
export const deleteProduct = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/products/${id}`)
}

