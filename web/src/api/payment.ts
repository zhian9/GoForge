import request from '@/utils/request'
import { useUserStore } from '@/stores/user'

const getCurrentUserId = (): number => {
  const store = useUserStore()
  return Number(store.userId || localStorage.getItem('user_id') || 0)
}

export interface Payment {
  id: number
  payment_no: string
  order_id: number
  order_no: string
  user_id: number
  amount: string
  payment_method: number
  status: number
  third_party_no: string
  paid_at: string
  expire_at: string
  created_at: string
}

export interface CreatePaymentRequest {
  order_id: number
  order_no: string
  amount: string
  payment_method: number
}

export interface CreatePaymentResponse {
  code: number
  message: string
  data: Payment
  pay_url: string
}

// 创建支付单
export const createPayment = (data: CreatePaymentRequest) => {
  return request.post<CreatePaymentResponse>('/v1/payments', {
    order_id: data.order_id,
    order_no: data.order_no,
    user_id: getCurrentUserId(),
    amount: data.amount,
    payment_method: data.payment_method,
  })
}

// 查询支付状态
export const queryPaymentStatus = (paymentNo: string) => {
  return request.get<{ code: number; message: string; status: number }>(
    `/v1/payments/${paymentNo}/status`
  )
}

// 模拟第三方支付回调（Mock 渠道专用）
export const mockPayCallback = (data: {
  payment_no: string
  status: number
  third_party_no?: string
}) => {
  return request.post<{ code: number; message: string }>('/v1/payments/callback', data)
}
