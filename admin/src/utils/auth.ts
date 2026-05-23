import { useUserStore } from '@/stores/user'

/**
 * 获取当前用户的 token
 */
export const getToken = (): string => {
  const userStore = useUserStore()
  return userStore.token || localStorage.getItem('token') || ''
}

/**
 * 设置 token
 */
export const setToken = (token: string): void => {
  const userStore = useUserStore()
  userStore.token = token
  localStorage.setItem('token', token)
}

/**
 * 移除 token
 */
export const removeToken = (): void => {
  const userStore = useUserStore()
  userStore.logout()
}

