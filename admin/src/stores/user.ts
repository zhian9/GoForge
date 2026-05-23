import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getUserInfo } from '@/api/user'
import type { LoginRequest, UserInfo } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  // 登录
  const login = async (data: LoginRequest) => {
    try {
      // 确保 login_type 字段存在（默认为 1-用户名登录）
      const loginData = {
        username: data.username,
        password: data.password,
        login_type: data.login_type || 1,
      }
      
      const response = await loginApi(loginData)
      if (response.data?.token) {
        token.value = response.data.token
        localStorage.setItem('token', response.data.token)
        // 登录接口已返回 user，先用它填充（兼容 camelCase / snake_case）
        if (response.data.user) {
          userInfo.value = {
            ...response.data.user,
            is_admin: response.data.user.isAdmin ?? response.data.user.is_admin ?? 0
          } as any
        }
        // 后台刷新用户信息（失败不影响登录）
        fetchUserInfo()
        return response
      }
      throw new Error('登录失败：未返回 token')
    } catch (error: any) {
      console.error('Login error:', error)
      throw error
    }
  }

  // 获取用户信息
  const fetchUserInfo = async () => {
    try {
      const response = await getUserInfo()
      if (response.data) {
        userInfo.value = response.data
      }
    } catch (error) {
      console.error('Get user info error:', error)
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return {
    token,
    userInfo,
    login,
    fetchUserInfo,
    logout,
  }
})

