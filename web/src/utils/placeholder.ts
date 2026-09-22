/**
 * 商品图统一占位图。
 *
 * 改造前这段 SVG 在 8 个页面里各复制了一份，且用的是 #1a1f2e 纯色块 ——
 * 在现在的半透明玻璃卡片里会呈现成一块突兀的死灰。这里收敛成单一实现，
 * 换成与暗色玻璃底一致的斜向渐变 + 描边图形，并在卡片上弱化存在感。
 *
 * 另外 SeckillDetail 原先的兜底路径 /placeholder.png 在 public/ 下并不存在，
 * 图片加载失败时会二次 404，这里一并修正。
 */

const cache = new Map<string, string>()

/**
 * @param size  输出正方形的边长（CSS 像素）
 * @param label 占位图上的说明文字
 */
export const placeholderImage = (size = 200, label = '无图'): string => {
  const key = `${size}:${label}`
  const cached = cache.get(key)
  if (cached) return cached

  // 字号随尺寸缩放，保证 200 / 400 两种规格视觉一致
  const fontSize = Math.round((size / 200) * 13)
  const strokeWidth = size >= 400 ? 2 : 3

  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">` +
    '<defs>' +
    '<linearGradient id="g" x1="0" y1="0" x2="1" y2="1">' +
    '<stop offset="0" stop-color="#141B2B"/>' +
    '<stop offset="1" stop-color="#0A0E18"/>' +
    '</linearGradient>' +
    '</defs>' +
    `<rect width="${size}" height="${size}" fill="url(#g)"/>` +
    `<circle cx="${size / 2}" cy="${size / 2}" r="${size * 0.26}" fill="none" stroke="#4FD8FF" stroke-opacity="0.14" stroke-width="${strokeWidth / 2}"/>` +
    '<g fill="none" stroke="#4A5670" stroke-linecap="round" stroke-linejoin="round" ' +
    `stroke-width="${strokeWidth}" transform="translate(${size / 2 - 100} ${size / 2 - 100})">` +
    '<rect x="72" y="70" width="56" height="44" rx="8"/>' +
    '<circle cx="88" cy="85" r="4.5"/>' +
    '<path d="M78 110l14-13 13 11 9-7 12 12"/>' +
    '</g>' +
    `<text x="${size / 2}" y="${size * 0.71}" fill="#5C6880" font-size="${fontSize}" ` +
    `font-family="system-ui,-apple-system,'PingFang SC',sans-serif" text-anchor="middle">${label}</text>` +
    '</svg>'

  const uri = 'data:image/svg+xml,' + encodeURIComponent(svg)
  cache.set(key, uri)
  return uri
}

