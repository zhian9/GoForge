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

export interface ReviewStats {
  total_count: number
  rating_5_count: number
  rating_4_count: number
  rating_3_count: number
  rating_2_count: number
  rating_1_count: number
  average_rating: number
}

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

const normalizeStats = (raw: any): ReviewStats => {
  if (!raw) {
    return {
      total_count: 0,
      rating_5_count: 0,
      rating_4_count: 0,
      rating_3_count: 0,
      rating_2_count: 0,
      rating_1_count: 0,
      average_rating: 0,
    }
  }
  return {
    total_count: Number(raw.total_count ?? raw.totalCount ?? 0),
    rating_5_count: Number(raw.rating_5_count ?? raw.rating5Count ?? 0),
    rating_4_count: Number(raw.rating_4_count ?? raw.rating4Count ?? 0),
    rating_3_count: Number(raw.rating_3_count ?? raw.rating3Count ?? 0),
    rating_2_count: Number(raw.rating_2_count ?? raw.rating2Count ?? 0),
    rating_1_count: Number(raw.rating_1_count ?? raw.rating1Count ?? 0),
    average_rating: Number(raw.average_rating ?? raw.averageRating ?? 0),
  }
}

export const getProductReviews = (productId: number, params?: { page?: number; page_size?: number; rating?: number }) => {
  const p = {
    page: params?.page ?? 1,
    page_size: params?.page_size ?? 10,
    rating: params?.rating ?? 0,
  }
  return request.get<{ code: number; message: string; data: any[]; total: number }>(`/v1/reviews/product/${productId}`, { params: p })
    .then((res) => ({
      ...res,
      data: (res.data || []).map(normalizeReview),
      total: Number(res.total ?? 0),
    }))
}

export const getReviewStats = (productId: number) => {
  return request.get<{ code: number; message: string; data: any }>(`/v1/reviews/stats/${productId}`)
    .then((res) => ({
      ...res,
      data: normalizeStats(res.data),
    }))
}

export const createReview = (data: {
  user_id: number
  order_id: number
  order_item_id: number
  product_id: number
  sku_id: number
  rating: number
  content: string
  images?: string[]
  videos?: string[]
}) => {
  return request.post<{ code: number; message: string; data: any }>(`/v1/reviews`, data)
    .then((res) => ({
      ...res,
      data: normalizeReview(res.data),
    }))
}


