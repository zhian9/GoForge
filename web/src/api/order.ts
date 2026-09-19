import request from '@/utils/request'

export interface Order {
  id: number
  order_no: string
  user_id: number
  total_amount: number | string
  status: number
  payment_status?: number
  shipping_status?: number
  created_at: string
  updated_at: string
  items?: OrderItem[]
  // 网关用 protobuf 的 json 编码返回 camelCase，这里同时兼容两种命名，
  // 避免视图里写 order.orderNo 时类型报错（运行时本来就存在）。
  orderNo?: string
  userId?: number
  totalAmount?: number | string
  createdAt?: string
  updatedAt?: string
}

export interface OrderItem {
  id: number
  order_id: number
  product_id: number
  product_name?: string
  sku_id: number
  sku_code?: string
  sku_name?: string
  // 同上：兼容网关返回的 camelCase
  orderId?: number
  productId?: number
  productName?: string
  skuId?: number
  skuCode?: string
  skuName?: string
  skuImage?: string
  totalAmount?: number | string
  sku_image?: string
  sku_specs?: Record<string, string>
  quantity: number
  price: number | string // 后端返回的是字符串格式
  total_price: number | string // 后端返回的是字符串格式
  total_amount?: number | string // 别名，兼容不同字段名
}

export interface CreateOrderRequest {
  user_id?: number
  items: Array<{
    sku_id: number
    quantity: number
  }>
  address_id: number
  receiver_name?: string
  receiver_phone?: string
  receiver_address?: string
  clear_cart?: boolean
  order_type?: number
  coupon_id?: number
  remark?: string
}

export interface CreateOrderResponse {
  code: number
  message: string
  data: Order
}

export interface ListOrdersResponse {
  code: number
  message: string
  data: {
    list: Order[]
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}

export interface OrderDetailResponse {
  code: number
  message: string
  data: Order
}

// 创建订单
export const createOrder = (data: CreateOrderRequest) => {
  return request.post<CreateOrderResponse>('/v1/orders', data)
}

// 获取订单列表
export const getOrderList = (params?: {
  user_id?: number
  page?: number
  page_size?: number
  status?: number
}) => {
  return request.get<ListOrdersResponse>('/v1/orders', { params })
}

// 获取订单详情
export const getOrderDetail = (id: number) => {
  return request.get<OrderDetailResponse>(`/v1/orders/${id}`)
}

// 取消订单
export const cancelOrder = (id: number, reason?: string) => {
  return request.put<{ code: number; message: string }>(`/v1/orders/${id}/cancel`, { reason })
}

// 确认收货
export const confirmReceive = (id: number) => {
  return request.put<{ code: number; message: string }>(`/v1/orders/${id}/receive`, {})
}

