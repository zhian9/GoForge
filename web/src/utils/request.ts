import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '@/stores/user'
import { apiBaseUrl } from '@/utils/api'

/**
 * 401 时跳登录页。
 *
 * 必须带上当前地址：游客可能是在商品详情页点「加入购物车」触发的登录要求，
 * 登录后应当回到原页面继续操作，而不是被丢回首页。
 * 这里用 window.location 而不是 router，是为了不把路由实例耦合进请求层。
 */
const redirectToLogin = () => {
  const current = window.location.pathname + window.location.search
  if (current.startsWith('/login')) return
  window.location.href = `/login?redirect=${encodeURIComponent(current)}`
}

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  // 默认同源 /api（开发由 Vite 代理、生产由 Nginx 反代），也可用 VITE_API_BASE_URL 指定网关
  baseURL: apiBaseUrl,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
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
        redirectToLogin()
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
        redirectToLogin()
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

/**
 * 响应拦截器已经把 `res.data` 返回给调用方了，但 axios 的类型仍然把返回值标成
 * `AxiosResponse<T>` —— 类型与运行时不一致，导致所有调用方写 `res.code` / `res.data`
 * 时都会报 TS2339（本次修复前全项目累计 75 处）。
 *
 * 这里把实例断言成「直接返回业务响应体」的形态，让类型和运行时对上。
 * 只影响类型，不改任何运行时行为。
 */
export interface RequestInstance {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  // 保留原生实例上仍会被用到的成员，避免替换类型后其它地方报错
  interceptors: AxiosInstance['interceptors']
  defaults: AxiosInstance['defaults']
  request<T = any>(config: AxiosRequestConfig): Promise<T>
}

export default service as unknown as RequestInstance

