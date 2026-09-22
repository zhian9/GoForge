import request from '@/utils/request'

// 注意：后端通过 gRPC-Gateway 返回时，int64/uint64 会以 string 形式序列化，且字段名为 camelCase
export interface CategoryDTO {
  id: string
  parentId: string
  name: string
  level: number
  sort: number
  icon: string
  iconLocal: string
  image: string
  imageLocal: string
  description: string
  status: number
  children?: CategoryDTO[]
}

// 管理端统一使用 number id + camelCase 字段，避免到处写判断/兜底
export interface Category {
  id: number
  parentId: number
  name: string
  level: number
  sort: number
  icon: string
  iconLocal: string
  image: string
  imageLocal: string
  description: string
  status: number
  children: Category[]
}

const normalizeCategory = (dto: CategoryDTO): Category => {
  return {
    id: Number(dto.id),
    parentId: Number(dto.parentId),
    name: dto.name,
    level: dto.level,
    sort: dto.sort,
    icon: dto.icon,
    iconLocal: dto.iconLocal,
    image: dto.image,
    imageLocal: dto.imageLocal,
    description: dto.description,
    status: dto.status,
    children: (dto.children || []).map(normalizeCategory),
  }
}

// 获取分类树
export const getCategoryTree = (params?: { status?: number }) => {
  return request.get<{ code: number; message: string; data: CategoryDTO[] }>('/v1/categories/tree', { params }).then((res) => ({
    code: res.code,
    message: res.message,
    data: (res.data || []).map(normalizeCategory),
  }))
}

// 创建分类
export interface CreateCategoryRequest {
  parentId: number
  name: string
  level?: number
  sort?: number
  icon?: string
  iconLocal?: string
  image?: string
  imageLocal?: string
  description?: string
  status?: number
}

export const createCategory = (data: CreateCategoryRequest) => {
  return request.post<{ code: number; message: string; data: Category }>('/v1/categories', data)
}

// 更新分类
export const updateCategory = (id: number, data: Partial<CreateCategoryRequest>) => {
  return request.put<{ code: number; message: string; data: Category }>(`/v1/categories/${id}`, data)
}

// 删除分类
export const deleteCategory = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/categories/${id}`)
}

