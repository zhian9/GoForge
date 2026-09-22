import type { Router } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

/**
 * 登录页回跳地址归一化。
 *
 * 只接受站内绝对路径：`//evil.com` 这类协议相对地址会被浏览器当成外站，
 * 若不拦截就成了开放重定向漏洞，所以一并挡掉。
 */
export const normalizeRedirect = (redirect?: string | null): string => {
  if (!redirect) return '/'
  if (!redirect.startsWith('/') || redirect.startsWith('//')) return '/'
  // 回跳到登录页本身没有意义，忽略掉避免登录后原地打转
  if (redirect === '/login' || redirect.startsWith('/login?') || redirect.startsWith('/login/')) return '/'
  return redirect
}

/** 组装登录页路由地址，登录成功后由登录页读回 redirect 参数跳回原页面 */
export const loginLocation = (redirect?: string) => {
  const target = normalizeRedirect(redirect)
  return target === '/' ? { path: '/login' } : { path: '/login', query: { redirect: target } }
}

/**
 * 需要登录才能继续的操作统一走这里。
 *
 * 页面浏览（首页/商品列表/详情/秒杀）不调用它，游客可以直接逛；
 * 只有加购、下单、秒杀、评价这类动作才要求登录。
 *
 * @returns 已登录返回 true，未登录（并已跳转登录页）返回 false
 */
export const ensureLogin = (router: Router, message = '请先登录'): boolean => {
  const userStore = useUserStore()
  if (userStore.token) return true
  if (message) ElMessage.warning(message)
  router.push(loginLocation(router.currentRoute.value.fullPath))
  return false
}
