import { resolveAssetUrl } from '@/utils/api'
import { placeholderImage } from '@/utils/placeholder'

/**
 * 把后端返回的图片路径补全为可访问 URL。
 * 具体规则见 utils/api.ts：默认走同源相对路径。
 * 没有图片时返回内联占位图，避免再发起一次必然 404 的请求。
 */
export function getPublicUrl(url: string): string {
  return url ? resolveAssetUrl(url) : placeholderImage(400, '暂无图片')
}


