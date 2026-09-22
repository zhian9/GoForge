<template>
  <div class="review-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>评价管理</span>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchProductId"
          placeholder="商品ID（留空查全部）"
          style="width: 200px; margin-right: 10px;"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchRating"
          placeholder="筛选评分"
          clearable
          style="width: 120px; margin-right: 10px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="null" />
          <el-option label="5星" :value="5" />
          <el-option label="4星" :value="4" />
          <el-option label="3星" :value="3" />
          <el-option label="2星" :value="2" />
          <el-option label="1星" :value="1" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="reviewList" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="product_id" label="商品ID" width="100" />
        <el-table-column prop="rating" label="评分" width="100">
          <template #default="{ row }">
            <el-rate v-model="row.rating" disabled show-score />
          </template>
        </el-table-column>
        <el-table-column prop="content" label="评价内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '已审核' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reply_content" label="回复内容" min-width="150" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleReply(row)">回复</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 回复对话框 -->
    <el-dialog
      v-model="replyDialogVisible"
      title="回复评价"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="replyFormRef"
        :model="replyFormData"
        :rules="replyFormRules"
        label-width="100px"
      >
        <el-form-item label="评价内容">
          <el-input v-model="replyFormData.content" type="textarea" :rows="3" disabled />
        </el-form-item>
        <el-form-item label="回复内容" prop="reply_content">
          <el-input v-model="replyFormData.reply_content" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleReplySubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getProductReviews,
  replyReview,
  type Review,
} from '@/api/review'

const loading = ref(false)
const reviewList = ref<Review[]>([])
const searchProductId = ref('')
const searchRating = ref<number | null>(null)

const replyDialogVisible = ref(false)
const submitting = ref(false)
const replyFormRef = ref<FormInstance>()
const replyFormData = ref({
  id: 0,
  content: '',
  reply_content: '',
})

const replyFormRules: FormRules = {
  reply_content: [{ required: true, message: '请输入回复内容', trigger: 'blur' }],
}

const fetchReviewList = async () => {
  loading.value = true
  try {
    // 留空 = 查询全部评价（product_id=0）
    const productId = searchProductId.value ? parseInt(searchProductId.value) : 0
    if (searchProductId.value && isNaN(productId)) {
      ElMessage.error('商品ID必须是数字')
      return
    }

    const params: any = { page: 1, page_size: 100 }
    // proto: rating=0 表示全部
    params.rating = searchRating.value === null ? 0 : searchRating.value

    const response = await getProductReviews(productId, params)
    if (response.code === 0) {
      // 后端返回：data 是数组
      reviewList.value = response.data || []
    } else {
      ElMessage.error(response.message || '获取评价列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取评价列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchReviewList()
}

const handleReply = (row: Review) => {
  replyFormData.value = {
    id: row.id,
    content: row.content,
    reply_content: row.reply_content || '',
  }
  replyDialogVisible.value = true
}

const handleReplySubmit = async () => {
  if (!replyFormRef.value) return
  
  await replyFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await replyReview(replyFormData.value.id, replyFormData.value.reply_content)
      ElMessage.success('回复成功')
      replyDialogVisible.value = false
      fetchReviewList()
    } catch (error: any) {
      ElMessage.error(error.message || '回复失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  fetchReviewList()
})
</script>

<style scoped>
.review-list {
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
</style>

