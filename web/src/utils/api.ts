/**
 * 网关地址统一解析。
 *
 * VITE_API_BASE_URL 指 API 网关根地址（如 http://localhost:8080 或 https://api.example.com）。
 * 未配置时回落到同源相对路径：开发环境由 Vite 代理把 /api、/uploads 转发到网关，
 * 生产环境由 Nginx 反代，因此代码里不写死主机名，部署到任意域名都能直接使用。
 */
const rawBase = String(import.meta.env.VITE_API_BASE_URL ?? '').trim()

/** 网关根地址（结尾不带 /，并兼容误配成 .../api 的情况）；为空字符串表示与前端同源 */
export const gatewayBaseUrl = rawBase.replace(/\/+$/, '').replace(/\/api$/, '')

/** axios 的 baseURL */
export const apiBaseUrl = gatewayBaseUrl ? `${gatewayBaseUrl}/api` : '/api'

/** 文件上传地址（el-upload 的 action、fetch 上传都用它） */
export const uploadUrl = `${apiBaseUrl}/v1/files/upload`

/** 把后端返回的相对资源路径补全为可访问 URL；绝对地址与内联数据原样返回 */
export const resolveAssetUrl = (url?: string | null): string => {
  if (!url) return ''
  if (/^(https?:)?\/\//i.test(url) || url.startsWith('data:') || url.startsWith('blob:')) return url
  return `${gatewayBaseUrl}${url.startsWith('/') ? '' : '/'}${url}`
}
