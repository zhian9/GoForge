import request from '@/utils/request'
import { useUserStore } from '@/stores/user'

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

export interface RemoveItemRequest {
  skuIds: number[]
}

// 获取购物车
export const getCart = () => {
  const userId = getCurrentUserId()
  return request.get<GetCartResponse>('/v1/cart', {
    params: userId ? { user_id: userId } : undefined,
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

