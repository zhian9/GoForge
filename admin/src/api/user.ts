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
  balance?: number
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
  return request.get<{ code: number; message: string; data: UserInfo }>('/v1/user/info').then((res) => ({
    ...res,
    data: normalizeUserInfo(res.data),
  }))
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
    /** 网关实际返回的是 list，两个字段名都保留，调用方按哪种写都能取到 */
    list?: UserInfo[]
  }
}

/**
 * 用户信息归一化。
 *
 * 后端经 gRPC-Gateway 返回 camelCase（memberLevel / createdAt / updatedAt / isAdmin），
 * 而 UserInfo 声明的是 snake_case：用户管理页的「注册时间」列因此一直是空白，
 * 会员等级也读不到。这里在 API 边界统一成声明形状，两种命名都兼容。
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
  is_admin: Number(dto?.isAdmin ?? dto?.is_admin ?? 0),
  balance: dto?.balance === undefined || dto?.balance === null ? undefined : Number(dto.balance),
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
  updated_at: dto?.updatedAt ?? dto?.updated_at ?? '',
})

export const getUserList = (params: ListUsersParams) => {
  return request.get<ListUsersResponse>('/v1/users', { params }).then((res) => {
    const raw: any[] = (res?.data?.users || (res?.data as any)?.list || []) as any[]
    const users = raw.map(normalizeUserInfo)
    return { ...res, data: { ...res.data, users, list: users } }
  })
}

// 删除用户（管理后台）
export const deleteUser = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/users/${id}`)
}

