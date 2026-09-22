import axios from 'axios'
import { ElMessage } from 'element-plus'
import { apiBaseUrl } from '@/utils/api'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '@/stores/user'

// 创建 axios 实例
// 默认同源 /api（开发由 Vite 代理、生产由 Nginx 反代），也可用 VITE_API_BASE_URL 指定网关
const service: AxiosInstance = axios.create({
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
    const res = response.data

    // 如果返回的状态码不是 200，则视为错误
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

/**
 * 响应拦截器已经把 `res.data` 返回给调用方了，但 axios 的类型仍然把返回值标成
 * `AxiosResponse<T>` —— 类型与运行时不一致，导致所有调用方写 `res.code` / `res.data`
 * 时都会报 TS2339。
 *
 * 用户端 web/src/utils/request.ts 早已做了同样的修正，后台这边一直没同步，
 * 所以 admin 的 `npm run build`（vue-tsc）此前是失败的（Docker 里只跑 vite build，
 * 不做类型检查，所以线上一直是好的）。
 *
 * 这里把实例断言成「直接返回业务响应体」的形态，让类型和运行时对上。只影响类型。
 */
export interface RequestInstance {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  interceptors: AxiosInstance['interceptors']
  defaults: AxiosInstance['defaults']
  request<T = any>(config: AxiosRequestConfig): Promise<T>
}

export default service as unknown as RequestInstance

