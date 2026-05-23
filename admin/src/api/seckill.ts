import request from '@/utils/request'

// go-zero gateway 返回的 protobuf JSON 默认是 camelCase（skuId/startTime/enableStatus），
// 但页面使用 snake_case（sku_id/start_time/enable_status）。这里做一次兼容归一化。
const normalizeActivity = (raw: any): SeckillActivity => {
  if (!raw) return raw
  const id = Number(raw.id ?? raw.Id ?? 0)
  const skuId = Number(raw.sku_id ?? raw.skuId ?? raw.skuID ?? 0)
  const startTime = Number(raw.start_time ?? raw.startTime ?? 0)
  const endTime = Number(raw.end_time ?? raw.endTime ?? 0)
  const stock = raw.stock !== undefined ? Number(raw.stock) : undefined
  const sold = raw.sold !== undefined ? Number(raw.sold) : undefined
  const status = raw.status !== undefined ? Number(raw.status) : undefined
  const enableStatus = raw.enable_status ?? raw.enableStatus

  return {
    id,
    name: raw.name ?? '',
    sku_id: skuId,
    sku_name: raw.sku_name ?? raw.skuName,
    sku_image: raw.sku_image ?? raw.skuImage,
    seckill_price: raw.seckill_price ?? raw.seckillPrice,
    original_price: raw.original_price ?? raw.originalPrice,
    stock,
    sold,
    start_time: startTime || undefined,
    end_time: endTime || undefined,
    status,
    enable_status: enableStatus !== undefined ? Number(enableStatus) : undefined,
  }
}

export interface SeckillActivity {
  id: number
  name: string
  sku_id: number
  sku_name?: string
  sku_image?: string
  seckill_price?: string
  original_price?: string
  stock?: number
  sold?: number
  start_time?: number
  end_time?: number
  status?: number // 0-未开始，1-进行中，2-已结束
  enable_status?: number // 0-禁用，1-启用
}

export interface SeckillActivityListResponse {
  code: number
  message: string
  data: {
    list: SeckillActivity[]
    page: number
    page_size: number
    total: number
    total_pages?: number
  }
}

export const getSeckillActivityList = (params?: {
  page?: number
  page_size?: number
  status?: number // 0/1/2，-1 表示全部
  include_disabled?: boolean
}) => {
  return request.get<SeckillActivityListResponse>('/v1/seckill/activities', { params }).then((res) => {
    const list = res?.data?.list || []
    const data: any = res?.data || {}
    return {
      ...res,
      data: {
        ...data,
        // 兼容 gateway 的 camelCase 字段
        page: Number(data.page ?? 1),
        page_size: Number(data.page_size ?? data.pageSize ?? 10),
        total: Number(data.total ?? 0),
        total_pages: Number(data.total_pages ?? data.totalPages ?? 0),
        list: Array.isArray(list) ? list.map(normalizeActivity) : [],
      },
    }
  })
}

export interface SeckillActivityDetailResponse {
  code: number
  message: string
  data: SeckillActivity
}

export const getSeckillActivityDetail = (id: number) => {
  return request.get<SeckillActivityDetailResponse>(`/v1/seckill/activities/${id}`).then((res) => {
    return {
      ...res,
      data: normalizeActivity(res.data) as any,
    }
  })
}

export interface CreateSeckillActivityRequest {
  name: string
  sku_id: number
  seckill_price: string
  stock: number
  start_time: number // Unix 秒
  end_time: number // Unix 秒
  enable_status: number // 0/1
}

export const createSeckillActivity = (data: CreateSeckillActivityRequest) => {
  return request.post<{ code: number; message: string; data: SeckillActivity }>('/v1/seckill/activities', data).then((res) => {
    return {
      ...res,
      data: normalizeActivity(res.data) as any,
    }
  })
}

export interface UpdateSeckillActivityRequest extends CreateSeckillActivityRequest {
  id: number
}

export const updateSeckillActivity = (id: number, data: CreateSeckillActivityRequest) => {
  return request.put<{ code: number; message: string; data: SeckillActivity }>(`/v1/seckill/activities/${id}`, data).then((res) => {
    return {
      ...res,
      data: normalizeActivity(res.data) as any,
    }
  })
}

export const deleteSeckillActivity = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/seckill/activities/${id}`)
}


