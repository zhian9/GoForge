export function getPublicUrl(url: string): string {
  if (!url) return '/placeholder.png'
  if (url.startsWith('http://') || url.startsWith('https://')) return url

  // 约定：VITE_API_BASE_URL 用于“静态资源域名”（默认 http://localhost:8080）
  // 兼容误配成 http://localhost:8080/api 的情况：静态资源不在 /api 下
  let base = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080'
  base = String(base).replace(/\/+$/, '')
  base = base.replace(/\/api$/, '')

  if (url.startsWith('/')) return `${base}${url}`
  return `${base}/${url}`
}


