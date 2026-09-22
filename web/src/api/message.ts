import request from '@/utils/request'

/** 站内消息 */
export interface Message {
  id: number
  user_id: number
  type: number // 1-系统通知 2-订单消息 3-营销消息 4-物流消息
  title: string
  content: string
  link?: string | null
  is_read: number // 0-未读 1-已读
  read_at?: string | null
  created_at: string
}

/** 网关返回 camelCase（userId / isRead / readAt / createdAt），统一归一化成 snake_case */
const normalizeMessage = (dto: any): Message => ({
  id: Number(dto?.id ?? 0),
  user_id: Number(dto?.userId ?? dto?.user_id ?? 0),
  type: Number(dto?.type ?? 0),
  title: dto?.title ?? '',
  content: dto?.content ?? '',
  link: dto?.link ?? null,
  is_read: Number(dto?.isRead ?? dto?.is_read ?? 0),
  read_at: dto?.readAt ?? dto?.read_at ?? null,
  created_at: dto?.createdAt ?? dto?.created_at ?? '',
})

/** 消息列表（用户身份由后端从 token 解析） */
export const getMessages = (params?: { type?: number; page?: number; page_size?: number }) =>
  request.get<{ code: number; message: string; data: Message[]; total: number }>('/v1/messages', { params })
    .then((res: any) => ({ ...res, data: (res?.data || []).map(normalizeMessage), total: Number(res?.total ?? 0) }))

/** 未读数量（响应是平铺的 count 字段，不是 data） */
export const getUnreadCount = () =>
  request.get<{ code: number; message: string; count: number }>('/v1/messages/unread-count')

/** 单条标记已读 */
export const markAsRead = (messageId: number) =>
  request.put<{ code: number; message: string }>(`/v1/messages/${messageId}/read`)

/** 批量标记已读 */
export const batchMarkAsRead = (messageIds: number[]) =>
  request.put<{ code: number; message: string }>('/v1/messages/batch-read', { message_ids: messageIds })

/** 消息类型文案 */
export const messageTypeText = (type: number): string =>
  ({ 1: '系统通知', 2: '订单消息', 3: '营销消息', 4: '物流消息' }[type] || '其他')
