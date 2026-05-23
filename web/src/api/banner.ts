import request from '@/utils/request'

export interface Banner {
  id: number
  title?: string
  description?: string
  image: string
  image_local?: string
  link?: string
  link_type: number // 1-商品详情, 2-分类页面, 3-外部链接, 4-无链接
  sort: number
  status: number // 0-禁用, 1-启用
  start_time?: string
  end_time?: string
  created_at: string
  updated_at: string
}

export interface BannerListResponse {
  code: number
  message: string
  data: Banner[]
}

export interface BannerDetailResponse {
  code: number
  message: string
  data: Banner
}

// 获取Banner列表
export const getBannerList = (params?: {
  status?: number // -1-全部, 0-禁用, 1-启用
  limit?: number // 限制数量，0表示不限制
}) => {
  return request.get<BannerListResponse>('/v1/banners', { params })
}

// 获取Banner详情
export const getBannerDetail = (id: number) => {
  return request.get<BannerDetailResponse>(`/v1/banners/${id}`)
}

