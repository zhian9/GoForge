import request from '@/utils/request'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

export interface FileInfo {
  file_id: string
  file_name: string
  file_url: string
  file_size: number
  file_type: string
  created_at: string
}

export interface UploadFileResponse {
  code: number
  message: string
  data: FileInfo
}

export interface BatchUploadFileResponse {
  code: number
  message: string
  data: FileInfo[]
}

// 上传文件
export const uploadFile = async (file: File, category: string = 'image'): Promise<FileInfo> => {
  const userStore = useUserStore()
  const formData = new FormData()
  formData.append('file', file)
  formData.append('category', category)

  const response = await axios.post<UploadFileResponse>(
    'http://localhost:8080/api/v1/files/upload',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
        Authorization: userStore.token ? `Bearer ${userStore.token}` : '',
      },
    }
  )

  if (response.data.code === 0) {
    return response.data.data
  }
  throw new Error(response.data.message || '上传失败')
}

// 批量上传文件
export const batchUploadFile = async (files: File[], category: string = 'image'): Promise<FileInfo[]> => {
  const userStore = useUserStore()
  const formData = new FormData()
  files.forEach((file) => {
    formData.append('files', file)
  })
  formData.append('category', category)

  const response = await axios.post<BatchUploadFileResponse>(
    'http://localhost:8080/api/v1/files/batch-upload',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
        Authorization: userStore.token ? `Bearer ${userStore.token}` : '',
      },
    }
  )

  if (response.data.code === 0) {
    return response.data.data
  }
  throw new Error(response.data.message || '上传失败')
}

// 删除文件
export const deleteFile = async (fileId: string) => {
  return request.delete<{ code: number; message: string }>(`/v1/files/${fileId}`)
}

