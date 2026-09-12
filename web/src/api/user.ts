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
  created_at: string
  updated_at: string
}

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
  return request.post<LoginResponse>('/v1/user/login', requestData)
}

// 用户注册
export const register = (data: RegisterRequest) => {
  return request.post<RegisterResponse>('/v1/user/register', data)
}

// 获取用户信息
// 注意：当前网关转发到 gRPC 时不会自动给 req.user_id 赋值，必须显式传 user_id（否则会 401 并触发前端重定向到 /login）
export const getUserInfo = (userId?: number) => {
  return request.get<{ code: number; message: string; data: UserInfo }>('/v1/user/info', {
    params: userId ? { user_id: userId } : undefined,
  })
}

// 每日签到
export const signIn = (userId: number) => {
  return request.post<{ code: number; message: string; addedPoints?: number; totalPoints?: number }>(
    '/v1/user/signin',
    { user_id: userId }
  )
}

// 更新用户信息
export const updateUserInfo = (data: {
  nickname?: string; gender?: number; avatar?: string; birthday?: string; phone?: string; email?: string
}) => {
  return request.put<{ code: number; message: string; data: UserInfo }>('/v1/user/info', data)
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

// 获取地址列表
export const getAddressList = (userId: number) => {
  return request.get<{ code: number; message: string; data: Address[] }>('/v1/user/address', {
    params: { user_id: userId },
  })
}

// 添加地址
export const addAddress = (data: AddAddressRequest) => {
  return request.post<{ code: number; message: string; data: Address }>('/v1/user/address', data)
}

// 更新地址
export const updateAddress = (id: number, data: UpdateAddressRequest) => {
  return request.put<{ code: number; message: string; data: Address }>(`/v1/user/address/${id}`, data)
}

// 删除地址
export const deleteAddress = (id: number, userId: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/user/address/${id}`, {
    params: { user_id: userId },
  })
}

