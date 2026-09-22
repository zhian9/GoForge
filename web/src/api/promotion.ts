import request from '@/utils/request'

/** 优惠券模板（券本身，不含用户领取记录） */
export interface Coupon {
  id: number
  name: string
  type: number // 1-满减券 2-折扣券 3-免运费券
  discount_type: number // 1-固定金额 2-百分比折扣
  discount_value: number
  min_amount: number
  max_discount?: number | null
  total_count: number // -1 表示不限量
  used_count: number
  per_user_limit: number
  valid_start_time: string
  valid_end_time: string
  status: number
}

/** 用户领取到的券 */
export interface UserCoupon {
  id: number
  user_id: number
  coupon_id: number
  status: number // 0-未使用 1-已使用 2-已过期
  order_id?: number | null
  used_at?: string | null
  expire_at: string
  created_at: string
}

/**
 * 网关（gRPC-Gateway + protobuf JSON）返回的是 camelCase（minAmount / discountValue /
 * validEndTime…），而页面统一用 snake_case，这里做一次归一化，
 * 否则页面拿到的门槛、面额、有效期全是 undefined。
 */
const toNum = (v: any): number => {
  const n = Number(v ?? 0)
  return Number.isNaN(n) ? 0 : n
}

const normalizeCoupon = (dto: any): Coupon => ({
  id: toNum(dto?.id),
  name: dto?.name ?? '',
  type: toNum(dto?.type),
  discount_type: toNum(dto?.discountType ?? dto?.discount_type),
  discount_value: toNum(dto?.discountValue ?? dto?.discount_value),
  min_amount: toNum(dto?.minAmount ?? dto?.min_amount),
  max_discount: dto?.maxDiscount ?? dto?.max_discount ?? null,
  total_count: toNum(dto?.totalCount ?? dto?.total_count),
  used_count: toNum(dto?.usedCount ?? dto?.used_count),
  per_user_limit: toNum(dto?.perUserLimit ?? dto?.per_user_limit),
  valid_start_time: dto?.validStartTime ?? dto?.valid_start_time ?? '',
  valid_end_time: dto?.validEndTime ?? dto?.valid_end_time ?? '',
  status: toNum(dto?.status),
})

const normalizeUserCoupon = (dto: any): UserCoupon => ({
  id: toNum(dto?.id),
  user_id: toNum(dto?.userId ?? dto?.user_id),
  coupon_id: toNum(dto?.couponId ?? dto?.coupon_id),
  status: toNum(dto?.status),
  order_id: dto?.orderId ?? dto?.order_id ?? null,
  used_at: dto?.usedAt ?? dto?.used_at ?? null,
  expire_at: dto?.expireAt ?? dto?.expire_at ?? '',
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
})

/** 可领取的优惠券列表 */
export const getCouponList = (params?: { status?: number; page?: number; page_size?: number }) =>
  request.get<{ code: number; message: string; data: Coupon[] }>('/v1/promotion/coupons', { params })
    .then((res: any) => ({ ...res, data: (res?.data || []).map(normalizeCoupon) }))

/**
 * 领取优惠券。
 * 用户身份由后端从 token 解析（请求体不需要传 user_id），couponId 是券模板 ID。
 */
export const receiveCoupon = (couponId: number) =>
  request.post<{ code: number; message: string }>(`/v1/promotion/coupons/${couponId}/receive`)

/** 我的优惠券；status 传 -1 表示全部，0/1/2 分别对应未使用/已使用/已过期 */
export const getUserCoupons = (userId: number, status = -1) =>
  request.get<{ code: number; message: string; data: UserCoupon[] }>(`/v1/promotion/user-coupons/${userId}`, {
    params: { status },
  })
    .then((res: any) => ({ ...res, data: (res?.data || []).map(normalizeUserCoupon) }))

/**
 * 试算优惠金额（仅用于页面展示，最终金额以下单时的服务端校验为准）。
 * 注意：这里的 couponId 是券模板 ID，不是用户券 ID。
 */
export const calculateDiscount = (couponId: number, totalAmount: number) =>
  request.post<{ code: number; message: string; data: { discount_amount: string; final_amount: string } }>(
    '/v1/promotion/coupons/calculate',
    { coupon_id: couponId, total_amount: totalAmount.toFixed(2) },
  )
    // 注意：CalculateDiscountResponse 是平铺结构（discountAmount / finalAmount 在顶层，不在 data 里）
    .then((res: any) => ({
      ...res,
      data: {
        discount_amount: String(res?.discountAmount ?? res?.data?.discountAmount ?? res?.data?.discount_amount ?? '0'),
        final_amount: String(res?.finalAmount ?? res?.data?.finalAmount ?? res?.data?.final_amount ?? '0'),
      },
    }))

/** 券面文案：满 X 减 Y / X 折 / 免运费 */
export const couponFaceText = (coupon?: Coupon | null): string => {
  if (!coupon) return '优惠券'
  if (coupon.type === 3) return '免运费'
  if (coupon.discount_type === 2) return `${Number(coupon.discount_value)}% off`
  return `¥${Number(coupon.discount_value).toFixed(2)}`
}

/** 使用门槛文案 */
export const couponThresholdText = (coupon?: Coupon | null): string => {
  if (!coupon || Number(coupon.min_amount) <= 0) return '无门槛'
  return `满 ¥${Number(coupon.min_amount).toFixed(2)} 可用`
}
