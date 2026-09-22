import request from '@/utils/request'

export interface Review {
  id: number
  user_id: number
  order_id: number
  order_item_id?: number
  product_id: number
  sku_id: number
  rating: number
  content: string
  images: string[]
  videos: string[]
  status: number
  reply_content?: string
  created_at: string
}

// 网关返回 protobuf JSON 默认是 camelCase，这里统一成 snake_case 供页面使用
const normalizeReview = (raw: any): Review => {
  if (!raw) return raw
  return {
    id: Number(raw.id ?? 0),
    user_id: Number(raw.user_id ?? raw.userId ?? 0),
    order_id: Number(raw.order_id ?? raw.orderId ?? 0),
    order_item_id: raw.order_item_id ?? raw.orderItemId ? Number(raw.order_item_id ?? raw.orderItemId) : undefined,
    product_id: Number(raw.product_id ?? raw.productId ?? 0),
    sku_id: Number(raw.sku_id ?? raw.skuId ?? 0),
    rating: Number(raw.rating ?? 0),
    content: raw.content ?? '',
    images: (raw.images ?? []) as string[],
    videos: (raw.videos ?? []) as string[],
    status: Number(raw.status ?? 0),
    reply_content: raw.reply_content ?? raw.replyContent ?? '',
    created_at: raw.created_at ?? raw.createdAt ?? '',
  }
}

// 获取商品评价列表
export const getProductReviews = (productId: number, params?: {
  page?: number
  page_size?: number
  rating?: number
}) => {
  const p = {
    page: params?.page ?? 1,
    page_size: params?.page_size ?? 10,
    // proto 定义：0 表示全部
    rating: params?.rating ?? 0,
  }
  return request.get<{ code: number; message: string; data: any[]; total: number }>(`/v1/reviews/product/${productId}`, { params: p })
    .then((res) => ({
      ...res,
      data: (res.data || []).map(normalizeReview),
      total: Number(res.total ?? 0),
    }))
}

// 回复评价
export const replyReview = (id: number, replyContent: string) => {
  // proto: ReplyReviewRequest { review_id, user_id, content, parent_id }
  return request.put<{ code: number; message: string }>(`/v1/reviews/${id}/reply`, {
    user_id: 0,
    content: replyContent,
    parent_id: 0,
  })
}

// 说明：删除/隐藏评价的后端接口目前未定义（proto 无 DeleteReview）

