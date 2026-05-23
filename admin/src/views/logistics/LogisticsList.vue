<template>
  <div class="logistics-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>物流管理</span>
          <el-button type="primary" @click="handleAdd">创建物流单</el-button>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchOrderNo"
          placeholder="搜索订单号"
          style="width: 200px; margin-right: 10px;"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="logisticsList" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_no" label="订单号" width="180" />
        <el-table-column prop="logistics_no" label="物流单号" width="180" />
        <el-table-column prop="company_name" label="物流公司" width="150" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收货人" width="120" />
        <el-table-column prop="receiver_phone" label="收货电话" width="150" />
        <el-table-column prop="receiver_address" label="收货地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
            <el-button type="success" size="small" @click="handleTracking(row)">物流轨迹</el-button>
            <el-button type="warning" size="small" @click="handleUpdateStatus(row)">更新状态</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建物流单对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建物流单"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createFormData"
        :rules="createFormRules"
        label-width="100px"
      >
        <el-form-item label="订单ID" prop="order_id">
          <el-input-number v-model="createFormData.order_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="订单号" prop="order_no">
          <el-input v-model="createFormData.order_no" />
        </el-form-item>
        <el-form-item label="物流公司" prop="company_code">
          <el-select v-model="createFormData.company_code" style="width: 100%">
            <el-option label="顺丰" value="SF" />
            <el-option label="圆通" value="YTO" />
            <el-option label="中通" value="ZTO" />
            <el-option label="申通" value="STO" />
            <el-option label="韵达" value="YD" />
          </el-select>
        </el-form-item>
        <el-form-item label="收货人" prop="receiver_name">
          <el-input v-model="createFormData.receiver_name" />
        </el-form-item>
        <el-form-item label="收货电话" prop="receiver_phone">
          <el-input v-model="createFormData.receiver_phone" />
        </el-form-item>
        <el-form-item label="收货地址" prop="receiver_address">
          <el-input v-model="createFormData.receiver_address" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">创建</el-button>
      </template>
    </el-dialog>

    <!-- 物流轨迹对话框 -->
    <el-dialog
      v-model="trackingDialogVisible"
      title="物流轨迹"
      width="700px"
    >
      <el-timeline>
        <el-timeline-item
          v-for="(item, index) in trackingList"
          :key="index"
          :timestamp="item.time"
        >
          {{ item.description }}
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getLogistics,
  createLogistics,
  updateLogisticsStatus,
  queryTracking,
  type Logistics,
  type TrackingInfo,
} from '@/api/logistics'

const loading = ref(false)
const logisticsList = ref<Logistics[]>([])
const searchOrderNo = ref('')

const createDialogVisible = ref(false)
const trackingDialogVisible = ref(false)
const submitting = ref(false)
const createFormRef = ref<FormInstance>()

const createFormData = ref({
  order_id: 0,
  order_no: '',
  company_code: '',
  receiver_name: '',
  receiver_phone: '',
  receiver_address: '',
})

const trackingList = ref<TrackingInfo[]>([])

const createFormRules: FormRules = {
  order_id: [{ required: true, message: '请输入订单ID', trigger: 'blur' }],
  order_no: [{ required: true, message: '请输入订单号', trigger: 'blur' }],
  company_code: [{ required: true, message: '请选择物流公司', trigger: 'change' }],
  receiver_name: [{ required: true, message: '请输入收货人', trigger: 'blur' }],
  receiver_phone: [{ required: true, message: '请输入收货电话', trigger: 'blur' }],
  receiver_address: [{ required: true, message: '请输入收货地址', trigger: 'blur' }],
}

const getStatusType = (status: number) => {
  const statusMap: Record<number, string> = {
    0: 'info',    // 待发货
    1: 'warning', // 已发货
    2: 'success', // 运输中
    3: 'success', // 已送达
    4: 'danger',  // 异常
  }
  return statusMap[status] || ''
}

const getStatusText = (status: number) => {
  const statusMap: Record<number, string> = {
    0: '待发货',
    1: '已发货',
    2: '运输中',
    3: '已送达',
    4: '异常',
  }
  return statusMap[status] || '未知'
}

const fetchLogistics = async () => {
  if (!searchOrderNo.value) {
    ElMessage.warning('请输入订单号')
    return
  }
  
  loading.value = true
  try {
    // TODO: 需要后端支持按订单号查询，当前先按订单ID查询
    const response = await getLogistics(parseInt(searchOrderNo.value))
    if (response.code === 0) {
      logisticsList.value = [response.data]
    } else {
      ElMessage.error(response.message || '获取物流信息失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取物流信息失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchLogistics()
}

const handleAdd = () => {
  createFormData.value = {
    order_id: 0,
    order_no: '',
    company_code: '',
    receiver_name: '',
    receiver_phone: '',
    receiver_address: '',
  }
  createDialogVisible.value = true
}

const handleView = async (row: Logistics) => {
  try {
    const response = await getLogistics(row.order_id)
    if (response.code === 0) {
      ElMessageBox.alert(JSON.stringify(response.data, null, 2), '物流详情', {
        confirmButtonText: '确定',
      })
    } else {
      ElMessage.error(response.message || '获取物流详情失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取物流详情失败')
  }
}

const handleTracking = async (row: Logistics) => {
  try {
    const response = await queryTracking(row.logistics_no, row.company_code)
    if (response.code === 0) {
      trackingList.value = response.data.tracking || []
      trackingDialogVisible.value = true
    } else {
      ElMessage.error(response.message || '查询物流轨迹失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '查询物流轨迹失败')
  }
}

const handleUpdateStatus = async (row: Logistics) => {
  try {
    const { value: status } = await ElMessageBox.prompt('请输入新状态（0-待发货，1-已发货，2-运输中，3-已送达，4-异常）', '更新物流状态', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^[0-4]$/,
      inputErrorMessage: '请输入0-4之间的数字',
    })
    
    await updateLogisticsStatus(row.order_id, parseInt(status))
    ElMessage.success('更新成功')
    fetchLogistics()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '更新失败')
    }
  }
}

const handleCreateSubmit = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await createLogistics(createFormData.value)
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      fetchLogistics()
    } catch (error: any) {
      ElMessage.error(error.message || '创建失败')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.logistics-list {
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

