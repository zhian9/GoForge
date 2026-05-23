import request from '@/utils/request'

// 注意：后端通过网关返回时，int64/uint64 可能被序列化为 string

// go-zero gateway 返回的 protobuf JSON 默认是 camelCase（skuId/startTime/originalPrice...），
// 页面使用 snake_case（sku_id/start_time/original_price...）。这里做兼容归一化。
const normalizeActivity = (raw: any): SeckillActivity => {
  if (!raw) return raw
  return {
    id: raw.id ?? 0,
    name: raw.name ?? '',
    sku_id: raw.sku_id ?? raw.skuId ?? raw.skuID ?? 0,
    sku_name: raw.sku_name ?? raw.skuName,
    sku_image: raw.sku_image ?? raw.skuImage,
    seckill_price: raw.seckill_price ?? raw.seckillPrice,
    original_price: raw.original_price ?? raw.originalPrice,
    stock: raw.stock !== undefined ? Number(raw.stock) : undefined,
    sold: raw.sold !== undefined ? Number(raw.sold) : undefined,
    start_time: raw.start_time ?? raw.startTime,
    end_time: raw.end_time ?? raw.endTime,
    status: raw.status !== undefined ? Number(raw.status) : undefined,
  }
}

export interface SeckillActivity {
  id: number | string
  name: string
  sku_id: number | string
  sku_name?: string
  sku_image?: string
  seckill_price?: string
  original_price?: string
  stock?: number
  sold?: number
  start_time?: number | string
  end_time?: number | string
  status?: number // 0-未开始，1-进行中，2-已结束
}

export interface ListSeckillActivitiesResponse {
  code: number
  message: string
  data: {
    list: SeckillActivity[]
    page: number
    page_size: number
    total: number | string
    total_pages?: number
  }
}

export const listSeckillActivities = (params?: {
  page?: number
  page_size?: number
  status?: number // 0/1/2，-1 表示全部
}) => {
  return request.get<ListSeckillActivitiesResponse>('/v1/seckill/activities', { params }).then((res: any) => {
    const data: any = res?.data || {}
    const list = data.list || []
    return {
      ...res,
      data: {
        ...data,
        // 兼容 gateway 的 camelCase
        page: Number(data.page ?? 1),
        page_size: Number(data.page_size ?? data.pageSize ?? 10),
        total: Number(data.total ?? 0),
        total_pages: Number(data.total_pages ?? data.totalPages ?? 0),
        list: Array.isArray(list) ? list.map(normalizeActivity) : [],
      },
    } as ListSeckillActivitiesResponse
  })
}

export interface GetSeckillActivityResponse {
  code: number
  message: string
  data: SeckillActivity
}

export const getSeckillActivity = (id: number) => {
  return request.get<GetSeckillActivityResponse>(`/v1/seckill/activities/${id}`).then((res: any) => {
    return {
      ...res,
      data: normalizeActivity(res.data) as any,
    } as GetSeckillActivityResponse
  })
}

export interface SeckillRequest {
  user_id: number
  sku_id: number
  quantity?: number
}

export interface SeckillResponse {
  code: number
  message: string
  data?: {
    success?: boolean
    order_no?: string
    message?: string
  }
}

export const seckill = (data: SeckillRequest) => {
  return request.post<SeckillResponse>('/v1/seckill', {
    ...data,
    quantity: data.quantity ?? 1,
  })
}


