import request from '@/utils/request'

export interface Order {
  id: number
  order_no: string
  user_id: number
  total_amount: number
  status: number
  payment_status: number
  shipping_status: number
  items: OrderItem[]
  created_at: string
  updated_at: string
}

export interface OrderItem {
  id: number
  product_id: number
  sku_id: number
  quantity: number
  price: number
  total_price: number
}

export interface OrderListResponse {
  code: number
  message: string
  data: {
    orders: Order[]
    total: number
  }
}

export interface OrderDetailResponse {
  code: number
  message: string
  data: Order
}

// 获取订单列表
export const getOrderList = (params?: {
  page?: number
  page_size?: number
  user_id?: number
  status?: number
}) => {
  return request.get<OrderListResponse>('/v1/orders', { params })
}

// 获取订单详情
export const getOrderDetail = (id: number) => {
  return request.get<OrderDetailResponse>(`/v1/orders/${id}`)
}

// 取消订单
export const cancelOrder = (id: number, reason?: string) => {
  return request.put<{ code: number; message: string }>(`/v1/orders/${id}/cancel`, { reason })
}

// 更新订单状态
export const updateOrderStatus = (id: number, status: number) => {
  return request.put<{ code: number; message: string; data: Order }>(`/v1/orders/${id}`, { status })
}

// 发货
export const shipOrder = (id: number, data?: { company_code?: string; logistics_no?: string }) => {
  return request.post<{ code: number; message: string }>(`/v1/orders/${id}/ship`, data || {})
}

