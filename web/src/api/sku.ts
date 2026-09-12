import request from '@/utils/request'

export interface Sku {
  id: number
  product_id: number
  sku_code: string
  name: string
  specs: Record<string, string> // 规格属性
  price: number
  original_price?: number
  stock: number
  image?: string
  weight?: number
  volume?: number
  status: number
}

export interface SkuListResponse {
  code: number
  message: string
  data: Sku[] | {
    list: Sku[]
    total?: number
    page?: number
    page_size?: number
    total_pages?: number
  }
}

// 根据商品ID获取SKU列表
export const getSkusByProductId = (productId: number) => {
  return request.get<SkuListResponse>('/v1/skus', {
    params: {
      product_id: productId,
      status: 1, // 只获取上架的SKU
    },
  })
}

// 获取SKU详情
export const getSkuDetail = (id: number) => {
  return request.get<{ code: number; message: string; data: Sku }>(`/v1/skus/${id}`)
}

// 解析商品（SPU）的默认 SKU ID。
// 优先取「上架且有库存」中价格最低的 SKU，与列表页展示的最低价保持一致。
// 若该商品没有任何可用 SKU，返回 0。
export const getDefaultSkuId = async (productId: number): Promise<number> => {
  try {
    const res = await getSkusByProductId(productId) as any
    const list: any[] = Array.isArray(res?.data) ? res.data : (res?.data?.list || [])
    const active = list.filter((s: any) => Number(s.status) === 1 && Number(s.stock) > 0)
    const pool = active.length > 0 ? active : list
    if (pool.length === 0) return 0
    const cheapest = pool.reduce((min: any, s: any) => (Number(s.price) < Number(min.price) ? s : min), pool[0])
    return Number(cheapest.id)
  } catch {
    return 0
  }
}

