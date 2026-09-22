import request from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { getSkuDetail } from '@/api/sku'

const getCurrentUserId = (): number => {
  const store = useUserStore()
  const id = store.userId || Number(localStorage.getItem('user_id') || 0)
  return Number(id || 0)
}

export interface CartItem {
  id: number
  userId: number
  skuId: number
  quantity: number
  price: number
  productName: string
  productImage: string
  isSelected: number
  createdAt: string
  updatedAt: string
}

export interface GetCartResponse {
  code: number
  message: string
  data: CartItem[] | {
    items: CartItem[]
  }
}

export interface AddItemRequest {
  skuId: number
  quantity: number
}

export interface AddItemResponse {
  code: number
  message: string
  data: CartItem
}

export interface UpdateQuantityRequest {
  skuId: number
  quantity: number
}

/** 统一字段类型：网关返回的 id / 价格都是字符串，这里一次转干净 */
const normalizeCartItem = (raw: any): CartItem => ({
  id: Number(raw?.id || 0),
  userId: Number(raw?.userId || 0),
  skuId: Number(raw?.skuId || 0),
  quantity: Number(raw?.quantity || 1),
  price: Number(raw?.price || 0),
  productName: raw?.productName || '',
  productImage: raw?.productImage || '',
  isSelected: Number(raw?.isSelected ?? 1),
  createdAt: raw?.createdAt || '',
  updatedAt: raw?.updatedAt || '',
})

/**
 * 补齐购物车条目的商品信息（名称 / 价格 / 图片）。
 *
 * 加购时服务端本应通过商品服务补齐这三项，但该链路曾因 RPC 客户端未初始化而失效，
 * 于是历史条目里只剩下 skuId：购物车和下单页的商品名称、价格、图片全空，
 * 合计金额也会被算成 0。
 *
 * 这里用 SKU 详情兜底：只对确实缺信息的条目发请求，并按 skuId 缓存，
 * 同一会话内不重复查询。后端恢复正常后，这段逻辑不会产生额外请求。
 */
const skuInfoCache = new Map<number, { name: string; price: number; image: string }>()

const resolveSkuInfo = async (skuId: number) => {
  const cached = skuInfoCache.get(skuId)
  if (cached) return cached
  try {
    const res: any = await getSkuDetail(skuId)
    const sku = res?.data
    if (!sku) return null
    const info = {
      name: String(sku.name || ''),
      price: Number(sku.price || 0),
      image: String(sku.image || ''),
    }
    skuInfoCache.set(skuId, info)
    return info
  } catch {
    return null
  }
}

const enrichCartItems = async (items: CartItem[]): Promise<CartItem[]> => {
  const pending = items.filter((i) => i.skuId > 0 && (!i.productName || i.price <= 0 || !i.productImage))
  if (pending.length === 0) return items

  const resolved = await Promise.all(pending.map(async (item) => [item.skuId, await resolveSkuInfo(item.skuId)] as const))
  const infoBySku = new Map(resolved.filter(([, info]) => info) as Array<[number, { name: string; price: number; image: string }]>)

  return items.map((item) => {
    const info = infoBySku.get(item.skuId)
    if (!info) return item
    return {
      ...item,
      productName: item.productName || info.name,
      price: item.price > 0 ? item.price : info.price,
      productImage: item.productImage || info.image,
    }
  })
}

// 获取购物车（返回前统一字段类型，并补齐缺失的商品信息）
export const getCart = () => {
  const userId = getCurrentUserId()
  return request
    .get<GetCartResponse>('/v1/cart', {
      params: userId ? { user_id: userId } : undefined,
    })
    .then(async (res) => {
      const raw = Array.isArray(res.data) ? res.data : (res.data?.items ?? [])
      const items = await enrichCartItems(raw.map(normalizeCartItem))
      return { ...res, data: items }
    })
}

// 添加商品到购物车
export const addItem = (data: AddItemRequest) => {
  const userId = getCurrentUserId()
  return request.post<AddItemResponse>('/v1/cart', {
    user_id: userId,
    sku_id: data.skuId,
    quantity: data.quantity,
  })
}

// 更新购物车商品数量
export const updateQuantity = (skuId: number, data: UpdateQuantityRequest) => {
  const userId = getCurrentUserId()
  return request.put<{ code: number; message: string }>(`/v1/cart/${skuId}`, {
    user_id: userId,
    sku_id: skuId,
    quantity: data.quantity,
  })
}

// 移除购物车商品
export const removeItem = (skuId: number) => {
  const userId = getCurrentUserId()
  return request.delete<{ code: number; message: string }>(`/v1/cart/${skuId}`, {
    params: userId ? { user_id: userId } : undefined,
  })
}

// 清空购物车
export const clearCart = (userId: number) => {
  return request.delete<{ code: number; message: string }>('/v1/cart', {
    params: { user_id: userId },
  })
}

// 选中/取消选中单个商品（持久化到后端）
export const selectItem = (skuId: number, isSelected: number) => {
  const userId = getCurrentUserId()
  return request.post<{ code: number; message: string }>('/v1/cart/select', {
    user_id: userId,
    sku_id: skuId,
    is_selected: isSelected,
  })
}

// 批量选中/取消选中（持久化到后端）
export const batchSelect = (skuIds: number[], isSelected: number) => {
  const userId = getCurrentUserId()
  return request.post<{ code: number; message: string }>('/v1/cart/select/batch', {
    user_id: userId,
    sku_ids: skuIds,
    is_selected: isSelected,
  })
}

