import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getCart } from '@/api/cart'
import type { CartItem } from '@/api/cart'

export const useCartStore = defineStore('cart', () => {
  const cartItems = ref<CartItem[]>([])
  const loading = ref(false)

  // 购物车商品总数
  const totalCount = computed(() => {
    return cartItems.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 购物车总金额（只计算选中的商品）
  const totalAmount = computed(() => {
    return cartItems.value
      .filter(item => item.isSelected === 1)
      .reduce((sum, item) => sum + item.price * item.quantity, 0)
  })

  // 获取购物车
  const fetchCart = async () => {
    loading.value = true
    try {
      const response = await getCart()
      if (response.code === 0) {
        // api 层已统一成数组（并补齐了商品名/价格/图片）；这里再兜一层历史对象结构
        const payload: any = response.data
        const newItems: any[] = Array.isArray(payload) ? payload : (payload?.items ?? [])

        // 规范化数据：price 可能为字符串，需要转 number；补充缺失字段
        const normalized: CartItem[] = newItems.map((item: any) => ({
          id: Number(item.id || 0),
          userId: Number(item.userId || 0),
          skuId: Number(item.skuId || 0),
          quantity: Number(item.quantity || 1),
          price: typeof item.price === 'string' ? Number(item.price) : Number(item.price || 0),
          productName: item.productName || '',
          productImage: item.productImage || '',
          isSelected: Number(item.isSelected ?? 1),
          createdAt: item.createdAt || '',
          updatedAt: item.updatedAt || '',
        }))

        cartItems.value = normalized
      }
    } catch (error) {
      console.error('Fetch cart error:', error)
    } finally {
      loading.value = false
    }
  }

  // 更新购物车商品数量
  const updateItemQuantity = (skuId: number, quantity: number) => {
    const item = cartItems.value.find(item => item.skuId === skuId)
    if (item) {
      item.quantity = quantity
    }
  }

  // 移除购物车商品
  const removeItem = (skuId: number) => {
    const index = cartItems.value.findIndex(item => item.skuId === skuId)
    if (index > -1) {
      cartItems.value.splice(index, 1)
    }
  }

  // 清空购物车
  const clearCart = () => {
    cartItems.value = []
  }

  return {
    cartItems,
    loading,
    totalCount,
    totalAmount,
    fetchCart,
    updateItemQuantity,
    removeItem,
    clearCart,
  }
})
