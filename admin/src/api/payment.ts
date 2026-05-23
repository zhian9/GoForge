import request from '@/utils/request'

export interface Payment {
  id: number
  payment_no: string
  order_no: string
  user_id: number
  amount: number
  status: number
  payment_method: number
  payment_time?: string
  created_at: string
  updated_at: string
}

export interface GetPaymentResponse {
  code: number
  message: string
  data: Payment
}

export interface CreatePaymentRequest {
  order_no: string
  user_id: number
  amount: number
  payment_method: number
}

export interface CreatePaymentResponse {
  code: number
  message: string
  data: Payment
}

export interface RefundRequest {
  payment_no: string
  refund_amount: number
  reason: string
}

export interface RefundResponse {
  code: number
  message: string
}

// 获取支付详情
export const getPayment = (paymentNo: string) => {
  return request.get<GetPaymentResponse>(`/v1/payments/${paymentNo}`)
}

// 创建支付
export const createPayment = (data: CreatePaymentRequest) => {
  return request.post<CreatePaymentResponse>('/v1/payments', data)
}

// 退款
export const refundPayment = (paymentNo: string, data: RefundRequest) => {
  return request.post<RefundResponse>(`/v1/payments/${paymentNo}/refund`, data)
}

