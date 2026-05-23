import request from '@/utils/request'

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
  sender_name?: string
  sender_phone?: string
  sender_address?: string
  created_at: string
  updated_at: string
}

export interface TrackingInfo {
  time: string
  status: string
  location: string
  description: string
}

// 获取物流信息
export const getLogistics = (orderId: number) => {
  return request.get<{ code: number; message: string; data: Logistics }>(`/v1/logistics/${orderId}`)
}

// 创建物流单
export const createLogistics = (data: {
  order_id: number
  order_no: string
  company_code: string
  receiver_name: string
  receiver_phone: string
  receiver_address: string
}) => {
  return request.post<{ code: number; message: string; data: Logistics }>('/v1/logistics', data)
}

// 更新物流状态
export const updateLogisticsStatus = (orderId: number, status: number) => {
  return request.put<{ code: number; message: string; data: Logistics }>(`/v1/logistics/${orderId}/status`, { status })
}

// 查询物流轨迹
export const queryTracking = (logisticsNo: string, companyCode: string) => {
  return request.get<{ code: number; message: string; data: { tracking: TrackingInfo[] } }>(`/v1/logistics/tracking`, {
    params: { logistics_no: logisticsNo, company_code: companyCode },
  })
}

// 计算运费
export const calculateFreight = (data: {
  weight: number
  volume: number
  from_address: string
  to_address: string
}) => {
  return request.post<{ code: number; message: string; data: { freight: number } }>('/v1/logistics/calculate-freight', data)
}

