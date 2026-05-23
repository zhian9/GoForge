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

// 获取购物车
export const getCart = (userId: number) => {
  return request.get<GetCartResponse>('/v1/cart', { params: { user_id: userId } })
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
  return request.put<{ code: number; message: string }>('/v1/cart/select', data)
}

// 批量选择/取消选择
export const batchSelectCartItems = (data: BatchSelectRequest) => {
  return request.put<{ code: number; message: string }>('/v1/cart/batch-select', data)
}

