import request from '@/utils/request'

// 注意：后端通过 gRPC-Gateway 返回时，int64/uint64 会以 string 形式序列化，且字段名为 camelCase
export interface ProductDTO {
  id: string
  spuCode?: string
  name: string
  subtitle?: string
  description?: string
  price: number
  originalPrice?: number
  stock: number
  categoryId: string
  brandId?: string
  images?: string[]
  mainImage?: string
  localMainImage?: string
  localImages?: string[]
  detail?: string
  status: number
  isHot?: number
  createdAt?: string
  updatedAt?: string
}

export interface Product {
  id: number
  name: string
  subtitle?: string
  description?: string
  price: number
  original_price: number
  stock: number
  category_id: number
  brand_id?: number
  images: string[]
  main_image?: string
  local_main_image?: string
  local_images?: string[]
  detail?: string
  status: number
  is_hot?: number // 是否热门: 0-否, 1-是
  created_at: string
  updated_at: string
}

const normalizeProduct = (dto: ProductDTO): Product => {
  return {
    id: Number(dto.id),
    name: dto.name,
    subtitle: dto.subtitle,
    description: dto.description,
    price: Number(dto.price || 0),
    original_price: Number(dto.originalPrice ?? 0),
    stock: Number(dto.stock || 0),
    category_id: Number(dto.categoryId || 0),
    brand_id: dto.brandId !== undefined ? Number(dto.brandId || 0) : undefined,
    images: (dto.images || []).slice(),
    main_image: dto.mainImage,
    local_main_image: dto.localMainImage,
    local_images: dto.localImages ? dto.localImages.slice() : undefined,
    detail: dto.detail,
    status: Number(dto.status || 0),
    is_hot: dto.isHot !== undefined ? Number(dto.isHot) : undefined,
    created_at: dto.createdAt || '',
    updated_at: dto.updatedAt || '',
  }
}

export interface ProductListResponse {
  code: number
  message: string
  data: {
    list: Product[]
    total: number
    page?: number
    pageSize?: number
    page_size?: number
    totalPages?: number
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
  status?: number
  is_hot?: number // -1-全部, 0-否, 1-是
}) => {
  return request
    .get<{ code: number; message: string; data: any }>('/v1/products', { params })
    .then((res) => {
      const raw = res.data || {}
      const rawList = Array.isArray(raw.list) ? raw.list : []
      const list = rawList.map((x: any) => normalizeProduct(x as ProductDTO))

      // total / page / pageSize 可能是 string（gateway），这里统一转 number
      const total = Number(raw.total ?? 0)
      const page = Number(raw.page ?? 0) || undefined
      const pageSize = Number(raw.pageSize ?? raw.page_size ?? 0) || undefined
      const totalPages = Number(raw.totalPages ?? raw.total_pages ?? 0) || undefined

      return {
        code: res.code,
        message: res.message,
        data: {
          list,
          total,
          page,
          pageSize,
          page_size: pageSize,
          totalPages,
          total_pages: totalPages,
        },
      } as ProductListResponse
    })
}

// 获取商品详情
export const getProductDetail = (id: number) => {
  return request.get<{ code: number; message: string; data: any }>(`/v1/products/${id}`).then((res) => {
    return {
      code: res.code,
      message: res.message,
      data: normalizeProduct(res.data as ProductDTO),
    } as ProductDetailResponse
  })
}

