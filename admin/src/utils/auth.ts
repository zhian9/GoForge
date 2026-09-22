import { useUserStore } from '@/stores/user'

/**
 * 获取当前用户的 token
 */
export const getToken = (): string => {
  const userStore = useUserStore()
  return userStore.token || localStorage.getItem('token') || ''
}

/**
 * 解析 JWT 的 payload。
 * 只做本地解码（不校验签名）—— 用途是判断手上的 token 是否带管理员能力位，
 * 真正的鉴权仍然由服务端完成。
 */
export const decodeJwtPayload = (token: string): Record<string, any> | null => {
  try {
    const segment = token.split('.')[1]
    if (!segment) return null
    const normalized = segment.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized + '='.repeat((4 - (normalized.length % 4)) % 4)
    const binary = atob(padded)
    const json = decodeURIComponent(
      Array.from(binary).map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)).join('')
    )
    return JSON.parse(json)
  } catch {
    return null
  }
}

/**
 * 判断 token 是否携带管理员标识。
 *
 * `is_admin` 是登录时写进 JWT 的：订单等服务要据此决定「能不能查全部数据」。
 * 早于该字段上线时签发的 token 不带它，服务端只能按普通用户处理 ——
 * 表现为后台菜单都能点，但订单列表永远是空的。
 * 这里主动识别这种 token 并强制重新登录，而不是让页面静默显示空数据。
 */
export const tokenCarriesAdmin = (token: string): boolean => {
  const payload = decodeJwtPayload(token)
  return payload !== null && Number(payload.is_admin ?? 0) === 1
}

