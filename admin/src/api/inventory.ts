import request from '@/utils/request'

export interface Inventory {
  sku_id: number
  total_stock: number
  available_stock: number
  locked_stock: number
  reserved_stock: number
}

export interface GetInventoryResponse {
  code: number
  message: string
  data: Inventory
}

export interface DeductStockRequest {
  sku_id: number
  quantity: number
  order_no: string
}

export interface DeductStockResponse {
  code: number
  message: string
}

// 获取库存信息
export const getInventory = (skuId: number) => {
  return request.get<GetInventoryResponse>(`/v1/inventory/${skuId}`)
}

// 扣减库存
export const deductStock = (data: DeductStockRequest) => {
  return request.post<DeductStockResponse>('/v1/inventory/deduct', data)
}

