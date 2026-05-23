<template>
  <div class="message-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>消息管理</span>
          <el-button type="primary" @click="handleAdd">发送消息</el-button>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchUserId"
          placeholder="搜索用户ID"
          style="width: 200px; margin-right: 10px;"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchType"
          placeholder="筛选类型"
          clearable
          style="width: 120px; margin-right: 10px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="null" />
          <el-option label="系统消息" :value="1" />
          <el-option label="订单消息" :value="2" />
          <el-option label="营销消息" :value="3" />
        </el-select>
        <el-select
          v-model="searchIsRead"
          placeholder="筛选状态"
          clearable
          style="width: 120px; margin-right: 10px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="null" />
          <el-option label="未读" :value="0" />
          <el-option label="已读" :value="1" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="messageList" v-loading="loading" border>
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            {{ getTypeText(row.type) }}
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="150" />
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="is_read" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_read === 1 ? 'success' : 'warning'">
              {{ row.is_read === 1 ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.is_read === 0"
              type="success"
              size="small"
              @click="handleMarkAsRead(row)"
            >
              标记已读
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="batch-actions" style="margin-top: 15px;">
        <el-button type="success" @click="handleBatchMarkAsRead">批量标记已读</el-button>
      </div>
    </el-card>

    <!-- 发送消息对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="发送消息"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createFormData"
        :rules="createFormRules"
        label-width="100px"
      >
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="createFormData.user_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="消息类型" prop="type">
          <el-select v-model="createFormData.type" style="width: 100%">
            <el-option label="系统消息" :value="1" />
            <el-option label="订单消息" :value="2" />
            <el-option label="营销消息" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createFormData.title" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="createFormData.content" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="链接" prop="link">
          <el-input v-model="createFormData.link" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getMessageList,
  sendMessage,
  markAsRead,
  batchMarkAsRead,
  deleteMessage,
  type Message,
} from '@/api/message'

const loading = ref(false)
const messageList = ref<Message[]>([])
const selectedMessages = ref<Message[]>([])
const searchUserId = ref('')
const searchType = ref<number | null>(null)
const searchIsRead = ref<number | null>(null)

const createDialogVisible = ref(false)
const submitting = ref(false)
const createFormRef = ref<FormInstance>()

const createFormData = ref({
  user_id: 0,
  type: 1,
  title: '',
  content: '',
  link: '',
})

const createFormRules: FormRules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  type: [{ required: true, message: '请选择消息类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
}

const getTypeText = (type: number) => {
  const typeMap: Record<number, string> = {
    1: '系统消息',
    2: '订单消息',
    3: '营销消息',
  }
  return typeMap[type] || '未知'
}

const fetchMessageList = async () => {
  if (!searchUserId.value) {
    ElMessage.warning('请输入用户ID')
    return
  }
  
  loading.value = true
  try {
    const userId = parseInt(searchUserId.value)
    if (isNaN(userId)) {
      ElMessage.error('用户ID必须是数字')
      return
    }
    
    const params: any = {
      page: 1,
      page_size: 100,
    }
    if (searchType.value !== null) {
      params.type = searchType.value
    }
    if (searchIsRead.value !== null) {
      params.is_read = searchIsRead.value
    }
    
    const response = await getMessageList(userId, params)
    if (response.code === 0) {
      messageList.value = response.data.messages || []
    } else {
      ElMessage.error(response.message || '获取消息列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取消息列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchMessageList()
}

const handleAdd = () => {
  createFormData.value = {
    user_id: 0,
    type: 1,
    title: '',
    content: '',
    link: '',
  }
  createDialogVisible.value = true
}

const handleMarkAsRead = async (row: Message) => {
  try {
    await markAsRead(row.id)
    ElMessage.success('标记成功')
    fetchMessageList()
  } catch (error: any) {
    ElMessage.error(error.message || '标记失败')
  }
}

const handleBatchMarkAsRead = async () => {
  // TODO: 实现批量标记已读（需要表格选择功能）
  ElMessage.info('批量标记功能待实现')
}

const handleDelete = async (row: Message) => {
  try {
    await ElMessageBox.confirm('确定要删除该消息吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteMessage(row.id)
    ElMessage.success('删除成功')
    fetchMessageList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const handleCreateSubmit = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await sendMessage(createFormData.value)
      ElMessage.success('发送成功')
      createDialogVisible.value = false
      fetchMessageList()
    } catch (error: any) {
      ElMessage.error(error.message || '发送失败')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.message-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.batch-actions {
  display: flex;
  align-items: center;
}
</style>

