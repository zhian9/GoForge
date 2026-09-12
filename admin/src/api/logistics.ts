import request from '@/utils/request'

// 注意：go-zero gateway 返回 camelCase（orderId/logisticsNo/companyName/currentLocation），
// 这里统一归一化为 snake_case 供页面使用。
const normalizeLogistics = (raw: any): Logistics => {
  if (!raw) return raw
  return {
    id: Number(raw.id ?? 0),
    order_id: Number(raw.orderId ?? raw.order_id ?? 0),
    order_no: raw.orderNo ?? raw.order_no ?? '',
    logistics_no: raw.logisticsNo ?? raw.logistics_no ?? '',
    company_code: raw.companyCode ?? raw.company_code ?? '',
    company_name: raw.companyName ?? raw.company_name ?? '',
    status: Number(raw.status ?? 0),
    receiver_name: raw.receiverName ?? raw.receiver_name ?? '',
    receiver_phone: raw.receiverPhone ?? raw.receiver_phone ?? '',
    receiver_address: raw.receiverAddress ?? raw.receiver_address ?? '',
    current_location: raw.currentLocation ?? raw.current_location ?? '',
    shipped_at: raw.shippedAt ?? raw.shipped_at ?? '',
    delivered_at: raw.deliveredAt ?? raw.delivered_at ?? '',
    created_at: raw.createdAt ?? raw.created_at ?? '',
    updated_at: raw.updatedAt ?? raw.updated_at ?? '',
  }
}

export interface Logistics {
  id: number
  order_id: number
  order_no: string
  logistics_no: string
  company_code: string
  company_name: string
  status: number
  receiver_name: string
  receiver_phone: string
  receiver_address: string
  current_location?: string
  shipped_at?: string
  delivered_at?: string
  created_at: string
  updated_at: string
}

export interface TrackingNode {
  time: string
  status: string
  location: string
  remark: string
}

// 获取物流列表
export const listLogistics = (params?: { page?: number; page_size?: number; order_no?: string }) => {
  return request.get<{ code: number; message: string; data: any[]; total: number }>('/v1/logistics', { params }).then((res: any) => {
    return {
      ...res,
      data: Array.isArray(res.data) ? res.data.map(normalizeLogistics) : [],
      total: Number(res.total ?? 0),
    }
  })
}

// 获取物流信息（按订单ID）
export const getLogistics = (orderId: number) => {
  return request.get<{ code: number; message: string; data: Logistics }>(`/v1/logistics/${orderId}`).then((res: any) => {
    return { ...res, data: normalizeLogistics(res.data) }
  })
}

// 更新物流状态（按物流单号）
export const updateLogisticsStatus = (logisticsNo: string, status: number, remark?: string) => {
  return request.put<{ code: number; message: string }>(`/v1/logistics/${logisticsNo}/status`, { status, remark })
}

// 查询物流轨迹
export const queryTracking = (logisticsNo: string) => {
  return request.get<{ code: number; message: string; data: TrackingNode[] }>('/v1/logistics/tracking', {
    params: { logistics_no: logisticsNo },
  })
}
