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

// 获取消息列表：user_id 由后端按 token 解析；管理员可以额外指定 user_id 查看某个用户的消息
export const getMessageList = (params?: {
  user_id?: number
  page?: number
  page_size?: number
  type?: number
  is_read?: number
}) => {
  return request.get<{ code: number; message: string; data: Message[]; total: number }>('/v1/messages', { params })
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

// 删除消息
export const deleteMessage = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/v1/messages/${id}`)
}

// 群发公告（仅管理员）：user_ids 为空表示发给所有启用用户
export const broadcastMessage = (data: {
  type: number
  title: string
  content: string
  link?: string
  user_ids?: number[]
}) => {
  return request.post<{ code: number; message: string; sent_count: number }>('/v1/messages/broadcast', data)
}

