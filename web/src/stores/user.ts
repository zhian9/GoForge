import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, register as registerApi, getUserInfo } from '@/api/user'
import type { LoginRequest, RegisterRequest, UserInfo } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userId = ref<number>(Number(localStorage.getItem('user_id') || 0) || 0)
  const userInfo = ref<UserInfo | null>(null)

  // 登录
  const login = async (data: LoginRequest) => {
    try {
      const loginData = {
        username: data.username,
        password: data.password,
        login_type: data.login_type || 1,
        verify_code: data.verify_code,
      }
      
      const response = await loginApi(loginData)
      if (response.data?.token) {
        token.value = response.data.token
        localStorage.setItem('token', response.data.token)
        // 登录接口本身会返回 user，先用它填充，避免因 /user/info 失败导致无法跳转首页
        if (response.data.user) {
          userInfo.value = response.data.user
          userId.value = response.data.user.id
          localStorage.setItem('user_id', String(response.data.user.id))
        }
        // 再后台刷新一次用户信息（失败不影响登录态/跳转）
        fetchUserInfo()
        return response
      }
      throw new Error('登录失败：未返回 token')
    } catch (error: any) {
      console.error('Login error:', error)
      throw error
    }
  }

  // 注册
  const register = async (data: RegisterRequest) => {
    try {
      const response = await registerApi(data)
      return response
    } catch (error: any) {
      console.error('Register error:', error)
      throw error
    }
  }

  // 获取用户信息
  const fetchUserInfo = async () => {
    try {
      const id = userInfo.value?.id || userId.value || undefined
      const response = await getUserInfo(id)
      if (response.data) {
        userInfo.value = response.data
        userId.value = response.data.id
        localStorage.setItem('user_id', String(response.data.id))
      }
    } catch (error) {
      console.error('Get user info error:', error)
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    userId.value = 0
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user_id')
  }

  return {
    token,
    userId,
    userInfo,
    login,
    register,
    fetchUserInfo,
    logout,
  }
})

