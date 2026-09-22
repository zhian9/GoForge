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

/**
 * 优惠券 / 用户券 / 营销活动归一化。
 *
 * 三个接口都返回「数组 + camelCase」（discountType / minAmount / validEndTime ...），
 * 而页面按 `data.coupons` 这类对象结构与 snake_case 读取 —— 不转换的话
 * 营销管理页会一直显示空列表，这是后台此前就存在的问题。
 */
const normalizeCoupon = (dto: any): Coupon => ({
  id: Number(dto?.id ?? 0),
  name: dto?.name ?? '',
  type: Number(dto?.type ?? 0),
  discount_type: Number(dto?.discountType ?? dto?.discount_type ?? 1),
  discount_value: String(dto?.discountValue ?? dto?.discount_value ?? '0'),
  min_amount: String(dto?.minAmount ?? dto?.min_amount ?? '0'),
  max_discount: String(dto?.maxDiscount ?? dto?.max_discount ?? ''),
  valid_start_time: dto?.validStartTime ?? dto?.valid_start_time ?? '',
  valid_end_time: dto?.validEndTime ?? dto?.valid_end_time ?? '',
  status: Number(dto?.status ?? 0),
})

const normalizeUserCoupon = (dto: any): UserCoupon => ({
  id: Number(dto?.id ?? 0),
  user_id: Number(dto?.userId ?? dto?.user_id ?? 0),
  coupon_id: Number(dto?.couponId ?? dto?.coupon_id ?? 0),
  status: Number(dto?.status ?? 0),
  order_id: Number(dto?.orderId ?? dto?.order_id ?? 0),
  expire_at: dto?.expireAt ?? dto?.expire_at ?? '',
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
})

/** 列表接口把数组包成页面期望的 { xxx, total } 结构 */
const wrapList = <T>(res: any, key: string, item: T[], total?: any) => ({
  ...res,
  data: { [key]: item, total: Number(total ?? res?.total ?? item.length) },
})

// 获取优惠券列表
export const getCouponList = (params?: {
  page?: number
  page_size?: number
  status?: number
}) => {
  return request
    .get<any>('/v1/promotion/coupons', { params })
    .then((res) => {
      const raw: any = res?.data
      const list = (Array.isArray(raw) ? raw : raw?.coupons || raw?.list || []).map(normalizeCoupon)
      return wrapList(res, 'coupons', list, raw?.total)
    })
}

// 获取用户优惠券列表
export const getUserCouponList = (userId: number, params?: {
  page?: number
  page_size?: number
  status?: number
}) => {
  return request
    .get<any>(`/v1/promotion/user-coupons/${userId}`, { params })
    .then((res) => {
      const raw: any = res?.data
      const list = (Array.isArray(raw) ? raw : raw?.user_coupons || raw?.list || []).map(normalizeUserCoupon)
      return wrapList(res, 'user_coupons', list, raw?.total)
    })
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

