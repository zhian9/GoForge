<template>
  <div class="payment-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>支付管理</span>
          <el-button type="primary" @click="handleAdd">创建支付</el-button>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchPaymentNo"
          placeholder="搜索支付单号"
          style="width: 200px; margin-right: 10px;"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchStatus"
          placeholder="筛选状态"
          clearable
          style="width: 120px; margin-right: 10px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="null" />
          <el-option label="待支付" :value="1" />
          <el-option label="已支付" :value="2" />
          <el-option label="已退款" :value="3" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="paymentList" v-loading="loading" border>
        <el-table-column prop="payment_no" label="支付单号" width="200" />
        <el-table-column prop="order_no" label="订单号" width="200" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }">
            ¥{{ row.amount }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="payment_method" label="支付方式" width="120">
          <template #default="{ row }">
            {{ getPaymentMethodText(row.payment_method) }}
          </template>
        </el-table-column>
        <el-table-column prop="payment_time" label="支付时间" width="180" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
            <el-button
              v-if="row.status === 2"
              type="warning"
              size="small"
              @click="handleRefund(row)"
            >
              退款
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建支付对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建支付"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createFormData"
        :rules="createFormRules"
        label-width="100px"
      >
        <el-form-item label="订单号" prop="order_no">
          <el-input v-model="createFormData.order_no" />
        </el-form-item>
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="createFormData.user_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number v-model="createFormData.amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="支付方式" prop="payment_method">
          <el-select v-model="createFormData.payment_method" style="width: 100%">
            <el-option label="微信" :value="1" />
            <el-option label="支付宝" :value="2" />
            <el-option label="银联" :value="3" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">创建</el-button>
      </template>
    </el-dialog>

    <!-- 退款对话框 -->
    <el-dialog
      v-model="refundDialogVisible"
      title="退款"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="refundFormRef"
        :model="refundFormData"
        :rules="refundFormRules"
        label-width="100px"
      >
        <el-form-item label="支付单号">
          <el-input v-model="refundFormData.payment_no" disabled />
        </el-form-item>
        <el-form-item label="退款金额" prop="refund_amount">
          <el-input-number v-model="refundFormData.refund_amount" :min="0.01" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="退款原因" prop="reason">
          <el-input v-model="refundFormData.reason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRefundSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createPayment, refundPayment, getPayment, type Payment, type CreatePaymentRequest, type RefundRequest } from '@/api/payment'

const loading = ref(false)
const paymentList = ref<Payment[]>([])
const searchPaymentNo = ref('')
const searchStatus = ref<number | null>(null)

const createDialogVisible = ref(false)
const refundDialogVisible = ref(false)
const submitting = ref(false)
const createFormRef = ref<FormInstance>()
const refundFormRef = ref<FormInstance>()

const createFormData = ref<CreatePaymentRequest>({
  order_no: '',
  user_id: 0,
  amount: 0,
  payment_method: 1,
})

const refundFormData = ref<RefundRequest>({
  payment_no: '',
  refund_amount: 0,
  reason: '',
})

const createFormRules: FormRules = {
  order_no: [{ required: true, message: '请输入订单号', trigger: 'blur' }],
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  payment_method: [{ required: true, message: '请选择支付方式', trigger: 'change' }],
}

const refundFormRules: FormRules = {
  refund_amount: [{ required: true, message: '请输入退款金额', trigger: 'blur' }],
  reason: [{ required: true, message: '请输入退款原因', trigger: 'blur' }],
}

const getStatusType = (status: number) => {
  const statusMap: Record<number, string> = {
    1: 'warning', // 待支付
    2: 'success', // 已支付
    3: 'danger',  // 已退款
  }
  return statusMap[status] || ''
}

const getStatusText = (status: number) => {
  const statusMap: Record<number, string> = {
    1: '待支付',
    2: '已支付',
    3: '已退款',
  }
  return statusMap[status] || '未知'
}

const getPaymentMethodText = (method: number) => {
  const methodMap: Record<number, string> = {
    1: '微信',
    2: '支付宝',
    3: '银联',
  }
  return methodMap[method] || '未知'
}

const handleSearch = () => {
  // TODO: 实现搜索功能（需要后端支持）
  ElMessage.info('搜索功能待实现')
}

const handleAdd = () => {
  createFormData.value = {
    order_no: '',
    user_id: 0,
    amount: 0,
    payment_method: 1,
  }
  createDialogVisible.value = true
}

const handleView = async (row: Payment) => {
  try {
    const response = await getPayment(row.payment_no)
    if (response.code === 0) {
      ElMessageBox.alert(JSON.stringify(response.data, null, 2), '支付详情', {
        confirmButtonText: '确定',
      })
    } else {
      ElMessage.error(response.message || '获取支付详情失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取支付详情失败')
  }
}

const handleRefund = (row: Payment) => {
  refundFormData.value = {
    payment_no: row.payment_no,
    refund_amount: row.amount,
    reason: '',
  }
  refundDialogVisible.value = true
}

const handleCreateSubmit = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await createPayment(createFormData.value)
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      // TODO: 刷新列表
    } catch (error: any) {
      ElMessage.error(error.message || '创建失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleRefundSubmit = async () => {
  if (!refundFormRef.value) return
  
  await refundFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await refundPayment(refundFormData.value.payment_no, refundFormData.value)
      ElMessage.success('退款成功')
      refundDialogVisible.value = false
      // TODO: 刷新列表
    } catch (error: any) {
      ElMessage.error(error.message || '退款失败')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.payment-list {
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

