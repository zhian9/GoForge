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
  /**
   * 网关（gRPC-Gateway）返回的是 camelCase，各页面模板里普遍写成
   * `row.orderNo || row.order_no` 这类兼容取值。这里把 camelCase 也声明出来，
   * 让类型和实际返回对齐，避免模板里报 TS2551。
   */
  orderNo?: string
  userId?: number
  totalAmount?: number
  receiverName?: string
  receiverPhone?: string
  receiverAddress?: string
  receiver_name?: string
  receiver_phone?: string
  receiver_address?: string
  createdAt?: string
  updatedAt?: string
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

// 发货
export const shipOrder = (id: number, data?: { company_code?: string; logistics_no?: string }) => {
  return request.post<{ code: number; message: string }>(`/v1/orders/${id}/ship`, data || {})
}

// 获取订单统计
export const getOrderStats = () => {
  return request.get<{ code: number; message: string; totalOrders?: number; totalSales?: string; todayOrders?: number }>('/v1/orders/stats')
}

