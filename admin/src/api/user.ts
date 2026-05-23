import request from '@/utils/request'

export interface LoginRequest {
  username: string
  password: string
  login_type?: number // 1-用户名, 2-手机号, 3-邮箱
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
  is_admin: number
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
    is_admin: number
  }
}

export interface RegisterResponse {
  code: number
  message: string
  data: UserInfo
}

// 用户登录
export const login = (data: LoginRequest) => {
  // 确保字段名匹配后端 proto 定义（snake_case）
  const requestData = {
    username: data.username,
    password: data.password,
    login_type: data.login_type || 1, // 默认用户名登录
  }
  return request.post<LoginResponse>('/v1/user/login', requestData)
}

// 用户注册
export const register = (data: RegisterRequest) => {
  return request.post<RegisterResponse>('/v1/user/register', data)
}

// 获取用户信息
export const getUserInfo = () => {
  return request.get<{ code: number; message: string; data: UserInfo }>('/v1/user/info')
}

// 更新用户信息
export const updateUserInfo = (data: Partial<UserInfo>) => {
  return request.put<{ code: number; message: string; data: UserInfo }>('/v1/user/info', data)
}

// 获取用户列表（管理后台）
export interface ListUsersParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: number // 0-全部，1-正常，2-禁用
}

export interface ListUsersResponse {
  code: number
  message: string
  data: {
    users: UserInfo[]
    total: number
    page: number
    page_size: number
  }
}

export const getUserList = (params: ListUsersParams) => {
  return request.get<ListUsersResponse>('/v1/users', { params })
}

// 删除用户（管理后台）
export const deleteUser = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/users/${id}`)
}

