<template>
  <div class="promotion-list">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="优惠券管理" name="coupons">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>优惠券列表</span>
              <el-button type="primary" @click="handleAddCoupon">新增优惠券</el-button>
            </div>
          </template>
          
          <div class="search-bar">
            <el-select
              v-model="couponStatus"
              placeholder="筛选状态"
              clearable
              style="width: 120px; margin-right: 10px;"
              @change="fetchCouponList"
            >
              <el-option label="全部" :value="null" />
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
            <el-button type="primary" @click="fetchCouponList">查询</el-button>
          </div>

          <el-table :data="couponList" v-loading="loading" border>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="优惠券名称" />
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                {{ row.type === 1 ? '满减券' : row.type === 2 ? '折扣券' : '其他' }}
              </template>
            </el-table-column>
            <el-table-column prop="discount_value" label="优惠值" width="120" />
            <el-table-column prop="min_amount" label="最低金额" width="120" />
            <el-table-column prop="valid_start_time" label="开始时间" width="180" />
            <el-table-column prop="valid_end_time" label="结束时间" width="180" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="handleEditCoupon(row)">编辑</el-button>
                <el-button type="danger" size="small" @click="handleDeleteCoupon(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="用户优惠券" name="user-coupons">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>用户优惠券</span>
            </div>
          </template>
          
          <div class="search-bar">
            <el-input
              v-model="searchUserId"
              placeholder="搜索用户ID"
              style="width: 200px; margin-right: 10px;"
              clearable
              @clear="fetchUserCouponList"
              @keyup.enter="fetchUserCouponList"
            />
            <el-button type="primary" @click="fetchUserCouponList">查询</el-button>
          </div>

          <el-table :data="userCouponList" v-loading="loading" border>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="user_id" label="用户ID" width="100" />
            <el-table-column prop="coupon_id" label="优惠券ID" width="100" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getUserCouponStatusType(row.status)">
                  {{ getUserCouponStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="expire_at" label="过期时间" width="180" />
            <el-table-column prop="created_at" label="领取时间" width="180" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 优惠券编辑对话框 -->
    <el-dialog
      v-model="couponDialogVisible"
      :title="couponDialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="couponFormRef"
        :model="couponFormData"
        :rules="couponFormRules"
        label-width="100px"
      >
        <el-form-item label="优惠券名称" prop="name">
          <el-input v-model="couponFormData.name" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="couponFormData.type" style="width: 100%">
            <el-option label="满减券" :value="1" />
            <el-option label="折扣券" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="优惠值" prop="discount_value">
          <el-input v-model="couponFormData.discount_value" />
        </el-form-item>
        <el-form-item label="最低金额" prop="min_amount">
          <el-input-number v-model="couponFormData.min_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="开始时间" prop="valid_start_time">
          <el-date-picker
            v-model="couponFormData.valid_start_time"
            type="datetime"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束时间" prop="valid_end_time">
          <el-date-picker
            v-model="couponFormData.valid_end_time"
            type="datetime"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="couponFormData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="couponDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCouponSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getCouponList,
  getUserCouponList,
  createCoupon,
  updateCoupon,
  deleteCoupon,
  type Coupon,
  type UserCoupon,
} from '@/api/promotion'

const activeTab = ref('coupons')
const loading = ref(false)
const couponList = ref<Coupon[]>([])
const userCouponList = ref<UserCoupon[]>([])
const couponStatus = ref<number | null>(null)
const searchUserId = ref('')

const couponDialogVisible = ref(false)
const couponDialogTitle = ref('新增优惠券')
const submitting = ref(false)
const couponFormRef = ref<FormInstance>()
const isEditCoupon = ref(false)
const couponFormData = ref({
  id: 0,
  name: '',
  type: 1,
  discount_type: 1,
  discount_value: '',
  min_amount: '',
  max_discount: '',
  valid_start_time: '',
  valid_end_time: '',
  status: 1,
})

const couponFormRules: FormRules = {
  name: [{ required: true, message: '请输入优惠券名称', trigger: 'blur' }],
  discount_value: [{ required: true, message: '请输入优惠值', trigger: 'blur' }],
  min_amount: [{ required: true, message: '请输入最低金额', trigger: 'blur' }],
  valid_start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  valid_end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

const fetchCouponList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: 1,
      page_size: 100,
    }
    if (couponStatus.value !== null) {
      params.status = couponStatus.value
    }
    const response = await getCouponList(params)
    if (response.code === 0) {
      couponList.value = response.data.coupons || []
    } else {
      ElMessage.error(response.message || '获取优惠券列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取优惠券列表失败')
  } finally {
    loading.value = false
  }
}

const fetchUserCouponList = async () => {
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
    const response = await getUserCouponList(userId)
    if (response.code === 0) {
      userCouponList.value = response.data.user_coupons || []
    } else {
      ElMessage.error(response.message || '获取用户优惠券列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取用户优惠券列表失败')
  } finally {
    loading.value = false
  }
}

const handleAddCoupon = () => {
  isEditCoupon.value = false
  couponDialogTitle.value = '新增优惠券'
  couponFormData.value = {
    id: 0,
    name: '',
    type: 1,
    discount_type: 1,
    discount_value: '',
    min_amount: '',
    max_discount: '',
    valid_start_time: '',
    valid_end_time: '',
    status: 1,
  }
  couponDialogVisible.value = true
}

const handleEditCoupon = (row: Coupon) => {
  isEditCoupon.value = true
  couponDialogTitle.value = '编辑优惠券'
  couponFormData.value = {
    id: row.id,
    name: row.name,
    type: row.type,
    discount_type: row.discount_type,
    discount_value: row.discount_value,
    min_amount: row.min_amount,
    max_discount: row.max_discount || '',
    valid_start_time: row.valid_start_time,
    valid_end_time: row.valid_end_time,
    status: row.status,
  }
  couponDialogVisible.value = true
}

const handleDeleteCoupon = async (row: Coupon) => {
  try {
    await ElMessageBox.confirm('确定要删除该优惠券吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteCoupon(row.id)
    ElMessage.success('删除成功')
    fetchCouponList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const handleCouponSubmit = async () => {
  if (!couponFormRef.value) return
  
  await couponFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      if (isEditCoupon.value) {
        await updateCoupon(couponFormData.value.id, couponFormData.value)
        ElMessage.success('更新成功')
      } else {
        await createCoupon(couponFormData.value as any)
        ElMessage.success('创建成功')
      }
      couponDialogVisible.value = false
      fetchCouponList()
    } catch (error: any) {
      ElMessage.error(error.message || (isEditCoupon.value ? '更新失败' : '创建失败'))
    } finally {
      submitting.value = false
    }
  })
}

const getUserCouponStatusType = (status: number) => {
  const statusMap: Record<number, string> = {
    1: 'success', // 未使用
    2: 'info',    // 已使用
    3: 'danger',  // 已过期
  }
  return statusMap[status] || ''
}

const getUserCouponStatusText = (status: number) => {
  const statusMap: Record<number, string> = {
    1: '未使用',
    2: '已使用',
    3: '已过期',
  }
  return statusMap[status] || '未知'
}

onMounted(() => {
  fetchCouponList()
})
</script>

<style scoped>
.promotion-list {
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

