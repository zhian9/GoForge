import request from '@/utils/request'

export interface Coupon {
  id: number
  name: string
  type: number
  discount_type: number
  discount_value: string
  min_amount: string
  max_discount: string
  valid_start_time: string
  valid_end_time: string
  status: number
}

export interface UserCoupon {
  id: number
  user_id: number
  coupon_id: number
  status: number
  order_id?: number
  expire_at: string
  created_at: string
}

export interface Promotion {
  id: number
  name: string
  type: number
  start_time: string
  end_time: string
  status: number
}

// 获取优惠券列表
export const getCouponList = (params?: {
  page?: number
  page_size?: number
  status?: number
}) => {
  return request.get<{ code: number; message: string; data: { coupons: Coupon[]; total: number } }>('/v1/promotion/coupons', { params })
}

// 获取用户优惠券列表
export const getUserCouponList = (userId: number, params?: {
  page?: number
  page_size?: number
  status?: number
}) => {
  return request.get<{ code: number; message: string; data: { user_coupons: UserCoupon[]; total: number } }>(`/v1/promotion/user-coupons/${userId}`, { params })
}

// 创建优惠券
export const createCoupon = (data: {
  name: string
  type: number
  discount_type: number
  discount_value: string
  min_amount: string
  max_discount?: string
  valid_start_time: string
  valid_end_time: string
}) => {
  return request.post<{ code: number; message: string; data: Coupon }>('/v1/promotion/coupons', data)
}

// 更新优惠券
export const updateCoupon = (id: number, data: Partial<Coupon>) => {
  return request.put<{ code: number; message: string; data: Coupon }>(`/v1/promotion/coupons/${id}`, data)
}

// 删除优惠券
export const deleteCoupon = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/promotion/coupons/${id}`)
}

// 获取促销活动列表
export const getPromotionList = (params?: {
  page?: number
  page_size?: number
  status?: number
}) => {
  return request.get<{ code: number; message: string; data: { promotions: Promotion[]; total: number } }>('/v1/promotion/promotions', { params })
}

