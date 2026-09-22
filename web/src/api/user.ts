import request from '@/utils/request'

export interface LoginRequest {
  username: string
  password: string
  login_type?: number // 1-用户名, 2-手机号, 3-邮箱
  verify_code?: string // 验证码
}

export interface RegisterRequest {
  username: string
  password: string
  phone?: string
  email?: string
  verify_code?: string
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  phone: string
  email: string
  avatar: string
  gender: number
  birthday?: string
  status: number
  member_level: number
  points: number
  balance?: number
  created_at: string
  updated_at: string
}

/**
 * 用户信息归一化。
 *
 * 后端经 gRPC-Gateway 返回的是 camelCase（memberLevel / createdAt / updatedAt），
 * 而前端各处（个人中心、stores/user）一律按 snake_case 取值，导致 member_level、
 * created_at、updated_at 始终读不到：会员等级永远显示"普通会员"、注册时间显示为空。
 *
 * 这里在 API 边界统一转成 UserInfo 声明的形状，两种命名都兼容，调用方无需改动。
 */
const normalizeUserInfo = (dto: any): UserInfo => ({
  id: Number(dto?.id ?? 0),
  username: dto?.username ?? '',
  nickname: dto?.nickname ?? '',
  phone: dto?.phone ?? '',
  email: dto?.email ?? '',
  avatar: dto?.avatar ?? '',
  gender: Number(dto?.gender ?? 0),
  birthday: dto?.birthday || undefined,
  status: Number(dto?.status ?? 0),
  member_level: Number(dto?.memberLevel ?? dto?.member_level ?? 0),
  points: Number(dto?.points ?? 0),
  balance: dto?.balance === undefined || dto?.balance === null ? undefined : Number(dto.balance),
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
  updated_at: dto?.updatedAt ?? dto?.updated_at ?? '',
})

export interface LoginResponse {
  code: number
  message: string
  data: {
    user: UserInfo
    token: string
    expire_time: number
  }
}

export interface RegisterResponse {
  code: number
  message: string
  data: UserInfo
}

// 用户登录
export const login = (data: LoginRequest) => {
  const requestData = {
    username: data.username,
    password: data.password,
    login_type: data.login_type || 1,
    verify_code: data.verify_code,
  }
  return request.post<LoginResponse>('/v1/user/login', requestData).then((res) => ({
    ...res,
    data: {
      ...res.data,
      user: normalizeUserInfo(res.data?.user),
    },
  }))
}

// 用户注册
export const register = (data: RegisterRequest) => {
  return request.post<RegisterResponse>('/v1/user/register', data).then((res) => ({
    ...res,
    data: normalizeUserInfo(res.data),
  }))
}

// 获取用户信息
// 注意：当前网关转发到 gRPC 时不会自动给 req.user_id 赋值，必须显式传 user_id（否则会 401 并触发前端重定向到 /login）
export const getUserInfo = (userId?: number) => {
  return request.get<{ code: number; message: string; data: UserInfo }>('/v1/user/info', {
    params: userId ? { user_id: userId } : undefined,
  }).then((res) => ({
    ...res,
    data: normalizeUserInfo(res.data),
  }))
}

// 每日签到
export const signIn = (userId: number) => {
  return request.post<{ code: number; message: string; addedPoints?: number; totalPoints?: number }>(
    '/v1/user/signin',
    { user_id: userId }
  )
}

// 余额充值
export const recharge = (userId: number, amount: number) => {
  return request.post<{ code: number; message: string; balance?: number }>(
    '/v1/user/recharge',
    { user_id: userId, amount: String(amount) }
  )
}

// 更新用户信息
export const updateUserInfo = (data: {
  nickname?: string; gender?: number; avatar?: string; birthday?: string; phone?: string; email?: string
}) => {
  return request.put<{ code: number; message: string; data: UserInfo }>('/v1/user/info', data).then((res) => ({
    ...res,
    data: normalizeUserInfo(res.data),
  }))
}

// ===== 收货地址 =====
export interface Address {
  id: number
  user_id: number
  receiver_name: string
  receiver_phone: string
  province: string
  city: string
  district: string
  detail: string
  postal_code: string
  is_default: number
  created_at: string
  updated_at: string
}

export interface AddAddressRequest {
  user_id: number
  receiver_name: string
  receiver_phone: string
  province: string
  city: string
  district: string
  detail: string
  postal_code?: string
  is_default?: number
}

export interface UpdateAddressRequest {
  id: number
  user_id: number
  receiver_name: string
  receiver_phone: string
  province: string
  city: string
  district: string
  detail: string
  postal_code?: string
  is_default?: number
}

/**
 * 地址归一化。
 *
 * 与用户信息同理：写接口接受 snake_case，但网关读出来的是 camelCase
 * （receiverName / receiverPhone / postalCode / isDefault / userId），
 * 导致个人中心的地址卡、下单页的地址选择读不到收件人和默认标记。
 */
const normalizeAddress = (dto: any): Address => ({
  id: Number(dto?.id ?? 0),
  user_id: Number(dto?.userId ?? dto?.user_id ?? 0),
  receiver_name: dto?.receiverName ?? dto?.receiver_name ?? '',
  receiver_phone: dto?.receiverPhone ?? dto?.receiver_phone ?? '',
  province: dto?.province ?? '',
  city: dto?.city ?? '',
  district: dto?.district ?? '',
  detail: dto?.detail ?? '',
  postal_code: dto?.postalCode ?? dto?.postal_code ?? '',
  is_default: Number(dto?.isDefault ?? dto?.is_default ?? 0),
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
  updated_at: dto?.updatedAt ?? dto?.updated_at ?? '',
})

// 获取地址列表
export const getAddressList = (userId: number) => {
  return request.get<{ code: number; message: string; data: Address[] }>('/v1/user/address', {
    params: { user_id: userId },
  }).then((res) => ({
    ...res,
    data: Array.isArray(res.data) ? res.data.map(normalizeAddress) : [],
  }))
}

// 添加地址
export const addAddress = (data: AddAddressRequest) => {
  return request.post<{ code: number; message: string; data: Address }>('/v1/user/address', data).then((res) => ({
    ...res,
    data: normalizeAddress(res.data),
  }))
}

// 更新地址
export const updateAddress = (id: number, data: UpdateAddressRequest) => {
  return request.put<{ code: number; message: string; data: Address }>(`/v1/user/address/${id}`, data).then((res) => ({
    ...res,
    data: normalizeAddress(res.data),
  }))
}

// 删除地址
export const deleteAddress = (id: number, userId: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/user/address/${id}`, {
    params: { user_id: userId },
  })
}

