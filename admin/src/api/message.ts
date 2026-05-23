import request from '@/utils/request'

export interface Message {
  id: number
  user_id: number
  type: number
  title: string
  content: string
  link?: string
  is_read: number
  created_at: string
}

// 获取消息列表
export const getMessageList = (userId: number, params?: {
  page?: number
  page_size?: number
  type?: number
  is_read?: number
}) => {
  return request.get<{ code: number; message: string; data: { messages: Message[]; total: number } }>(`/v1/messages/${userId}`, { params })
}

// 发送消息
export const sendMessage = (data: {
  user_id: number
  type: number
  title: string
  content: string
  link?: string
}) => {
  return request.post<{ code: number; message: string }>('/v1/messages', data)
}

// 标记消息已读
export const markAsRead = (id: number) => {
  return request.put<{ code: number; message: string }>(`/v1/messages/${id}/read`)
}

// 批量标记已读
export const batchMarkAsRead = (ids: number[]) => {
  return request.put<{ code: number; message: string }>('/v1/messages/batch-read', { ids })
}

// 获取未读消息数量
export const getUnreadCount = (userId: number) => {
  return request.get<{ code: number; message: string; data: { count: number } }>(`/v1/messages/${userId}/unread-count`)
}

// 删除消息
export const deleteMessage = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/messages/${id}`)
}

