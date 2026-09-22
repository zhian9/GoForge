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

/** Banner 归一化：网关返回 camelCase（imageLocal / linkType / startTime），与 Product 同理 */
const normalizeBanner = (dto: any): Banner => ({
  id: Number(dto?.id ?? 0),
  title: dto?.title ?? '',
  description: dto?.description ?? '',
  image: dto?.image ?? '',
  image_local: dto?.imageLocal ?? dto?.image_local ?? '',
  link: dto?.link ?? '',
  link_type: Number(dto?.linkType ?? dto?.link_type ?? 4),
  sort: Number(dto?.sort ?? 0),
  status: Number(dto?.status ?? 0),
  start_time: dto?.startTime ?? dto?.start_time ?? '',
  end_time: dto?.endTime ?? dto?.end_time ?? '',
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
  updated_at: dto?.updatedAt ?? dto?.updated_at ?? '',
})

// 获取Banner列表
export const getBannerList = (params?: {
  status?: number // -1-全部, 0-禁用, 1-启用
  limit?: number // 限制数量，0表示不限制
}) => {
  return request.get<BannerListResponse>('/v1/banners', { params }).then((res) => ({
    ...res,
    data: Array.isArray(res.data) ? res.data.map(normalizeBanner) : [],
  }))
}

// 创建Banner
export interface CreateBannerRequest {
  title?: string
  description?: string
  image: string
  image_local?: string
  link?: string
  link_type?: number
  sort?: number
  status?: number
  start_time?: string
  end_time?: string
}

export const createBanner = (data: CreateBannerRequest) => {
  return request.post<BannerDetailResponse>('/v1/banners', data)
}

// 更新Banner
export const updateBanner = (id: number, data: Partial<CreateBannerRequest>) => {
  return request.put<BannerDetailResponse>(`/v1/banners/${id}`, data)
}

// 删除Banner
export const deleteBanner = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/banners/${id}`)
}

