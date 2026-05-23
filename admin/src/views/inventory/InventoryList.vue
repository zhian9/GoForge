<template>
  <div class="inventory-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>库存管理</span>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchSkuId"
          placeholder="搜索SKU ID"
          style="width: 200px; margin-right: 10px;"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="inventoryList" v-loading="loading" border>
        <el-table-column prop="sku_id" label="SKU ID" width="120" />
        <el-table-column prop="total_stock" label="总库存" width="120" />
        <el-table-column prop="available_stock" label="可用库存" width="120">
          <template #default="{ row }">
            <el-tag :type="row.available_stock > 0 ? 'success' : 'danger'">
              {{ row.available_stock }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="locked_stock" label="锁定库存" width="120" />
        <el-table-column prop="reserved_stock" label="预留库存" width="120" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDeduct(row)">扣减库存</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 扣减库存对话框 -->
    <el-dialog
      v-model="deductDialogVisible"
      title="扣减库存"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="deductFormRef"
        :model="deductFormData"
        :rules="deductFormRules"
        label-width="100px"
      >
        <el-form-item label="SKU ID">
          <el-input-number v-model="deductFormData.sku_id" :min="1" disabled style="width: 100%" />
        </el-form-item>
        <el-form-item label="扣减数量" prop="quantity">
          <el-input-number v-model="deductFormData.quantity" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="订单号" prop="order_no">
          <el-input v-model="deductFormData.order_no" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deductDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleDeductSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getInventory, deductStock, type Inventory, type DeductStockRequest } from '@/api/inventory'

const loading = ref(false)
const inventoryList = ref<Inventory[]>([])
const searchSkuId = ref('')

const deductDialogVisible = ref(false)
const submitting = ref(false)
const deductFormRef = ref<FormInstance>()

const deductFormData = ref<DeductStockRequest>({
  sku_id: 0,
  quantity: 1,
  order_no: '',
})

const deductFormRules: FormRules = {
  quantity: [
    { required: true, message: '请输入扣减数量', trigger: 'blur' },
    { type: 'number', min: 1, message: '数量必须大于0', trigger: 'blur' },
  ],
  order_no: [{ required: true, message: '请输入订单号', trigger: 'blur' }],
}

const fetchInventory = async () => {
  if (!searchSkuId.value) {
    ElMessage.warning('请输入SKU ID')
    return
  }
  
  loading.value = true
  try {
    const skuId = parseInt(searchSkuId.value)
    if (isNaN(skuId)) {
      ElMessage.error('SKU ID必须是数字')
      return
    }
    
    const response = await getInventory(skuId)
    if (response.code === 0) {
      inventoryList.value = [response.data]
    } else {
      ElMessage.error(response.message || '获取库存信息失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取库存信息失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchInventory()
}

const handleDeduct = (row: Inventory) => {
  deductFormData.value = {
    sku_id: row.sku_id,
    quantity: 1,
    order_no: '',
  }
  deductDialogVisible.value = true
}

const handleDeductSubmit = async () => {
  if (!deductFormRef.value) return
  
  await deductFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await deductStock(deductFormData.value)
      ElMessage.success('扣减成功')
      deductDialogVisible.value = false
      fetchInventory()
    } catch (error: any) {
      ElMessage.error(error.message || '扣减失败')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.inventory-list {
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

