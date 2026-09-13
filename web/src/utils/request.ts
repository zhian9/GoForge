import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { useUserStore } from '@/stores/user'

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
service.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    const userStore = useUserStore()
    // 添加 token
    if (userStore.token) {
      config.headers = config.headers || {}
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data as any

    // 如果返回的状态码不是 0，则视为错误
    if (res.code !== 0 && res.code !== undefined) {
      ElMessage.error(res.message || '请求失败')
      // 401: 未授权，跳转到登录页
      if (res.code === 401) {
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }

    // 返回 response.data，这样调用方可以直接使用响应数据
    return res
  },
  (error) => {
    console.error('Response error:', error)
    
    // 处理 HTTP 错误状态码
    if (error.response) {
      const { status, data } = error.response
      let errorMessage = '请求失败'

      // 兼容 JSON 对象与 gRPC 错误文本（"rpc error: ... message: xxx"）
      let bizMsg = ''
      if (data) {
        if (typeof data === 'string') {
          const m = data.match(/message:\s*([^\n]+)/)
          if (m) bizMsg = m[1]
        } else if (data.message) {
          bizMsg = data.message
        }
      }

      if (bizMsg) {
        errorMessage = bizMsg
      } else if (status === 400) {
        errorMessage = '请求参数错误'
      } else if (status === 401) {
        errorMessage = '未授权，请重新登录'
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
      } else if (status === 404) {
        errorMessage = '请求的资源不存在'
      } else if (status === 500) {
        errorMessage = '服务器内部错误'
      }

      ElMessage.error(errorMessage)
      return Promise.reject(new Error(errorMessage))
    } else if (error.request) {
      ElMessage.error('网络错误，请检查网络连接')
      return Promise.reject(new Error('网络错误'))
    } else {
      ElMessage.error(error.message || '请求失败')
      return Promise.reject(error)
    }
  }
)

export default service

