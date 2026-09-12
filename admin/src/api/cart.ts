import request from '@/utils/request'

export interface CartItem {
  id: number
  user_id: number
  sku_id: number
  quantity: number
  is_selected: number
  created_at: string
  updated_at: string
}

export interface GetCartResponse {
  code: number
  message: string
  data: CartItem[]
}

export interface AddItemRequest {
  user_id: number
  sku_id: number
  quantity: number
}

export interface AddItemResponse {
  code: number
  message: string
  data: CartItem
}

export interface UpdateQuantityRequest {
  user_id: number
  sku_id: number
  quantity: number
}

export interface RemoveItemRequest {
  user_id: number
  sku_ids: number[]
}

export interface SelectItemRequest {
  user_id: number
  sku_id: number
  is_selected: number
}

export interface BatchSelectRequest {
  user_id: number
  sku_ids: number[]
  is_selected: number
}

// 网关返回 protobuf JSON 默认是 camelCase（userId/skuId/isSelected），这里统一成 snake_case 供页面使用
const normalizeCartItem = (raw: any): CartItem => {
  return {
    id: Number(raw.id ?? 0),
    user_id: Number(raw.user_id ?? raw.userId ?? 0),
    sku_id: Number(raw.sku_id ?? raw.skuId ?? 0),
    quantity: Number(raw.quantity ?? 0),
    is_selected: Number(raw.is_selected ?? raw.isSelected ?? 0),
    created_at: raw.created_at ?? raw.createdAt ?? '',
    updated_at: raw.updated_at ?? raw.updatedAt ?? '',
  }
}

// 获取购物车
export const getCart = (userId: number) => {
  return request.get<GetCartResponse>('/v1/cart', { params: { user_id: userId } })
    .then((res) => ({
      ...res,
      data: (Array.isArray(res.data) ? res.data : []).map(normalizeCartItem),
    }))
}

// 添加商品到购物车
export const addCartItem = (data: AddItemRequest) => {
  return request.post<AddItemResponse>('/v1/cart', data)
}

// 更新购物车商品数量
export const updateCartQuantity = (skuId: number, data: UpdateQuantityRequest) => {
  return request.put<{ code: number; message: string }>(`/v1/cart/${skuId}`, data)
}

// 删除购物车商品
export const removeCartItem = (skuId: number, data: RemoveItemRequest) => {
  return request.delete<{ code: number; message: string }>(`/v1/cart/${skuId}`, { data })
}

// 清空购物车
export const clearCart = (userId: number) => {
  return request.delete<{ code: number; message: string }>('/v1/cart', { params: { user_id: userId } })
}

// 选择/取消选择商品
export const selectCartItem = (data: SelectItemRequest) => {
  return request.post<{ code: number; message: string }>('/v1/cart/select', data)
}

// 批量选择/取消选择
export const batchSelectCartItems = (data: BatchSelectRequest) => {
  return request.post<{ code: number; message: string }>('/v1/cart/select/batch', data)
}

