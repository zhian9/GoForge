<template>
  <div class="cart-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>购物车管理</span>
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
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </div>

      <el-table :data="cartList" v-loading="loading" border>
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="sku_id" label="SKU ID" width="100" />
        <el-table-column prop="quantity" label="数量" width="100" />
        <el-table-column prop="is_selected" label="是否选中" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_selected === 1 ? 'success' : 'info'">
              {{ row.is_selected === 1 ? '已选中' : '未选中' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column prop="updated_at" label="更新时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="编辑购物车"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="formData.user_id" :min="1" disabled style="width: 100%" />
        </el-form-item>
        <el-form-item label="SKU ID" prop="sku_id">
          <el-input-number v-model="formData.sku_id" :min="1" disabled style="width: 100%" />
        </el-form-item>
        <el-form-item label="数量" prop="quantity">
          <el-input-number v-model="formData.quantity" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="是否选中" prop="is_selected">
          <el-radio-group v-model="formData.is_selected">
            <el-radio :value="0">未选中</el-radio>
            <el-radio :value="1">已选中</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getCart, updateCartQuantity, removeCartItem, selectCartItem, type CartItem } from '@/api/cart'

const loading = ref(false)
const cartList = ref<CartItem[]>([])
const searchUserId = ref('')

const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const formData = ref({
  id: 0,
  user_id: 0,
  sku_id: 0,
  quantity: 1,
  is_selected: 0,
})

const formRules: FormRules = {
  quantity: [
    { required: true, message: '请输入数量', trigger: 'blur' },
    { type: 'number', min: 1, message: '数量必须大于0', trigger: 'blur' },
  ],
}

const fetchCartList = async () => {
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
    
    const response = await getCart(userId)
    if (response.code === 0) {
      cartList.value = response.data || []
    } else {
      ElMessage.error(response.message || '获取购物车列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取购物车列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchCartList()
}

const handleEdit = (row: CartItem) => {
  formData.value = {
    id: row.id,
    user_id: row.user_id,
    sku_id: row.sku_id,
    quantity: row.quantity,
    is_selected: row.is_selected,
  }
  dialogVisible.value = true
}

const handleDelete = async (row: CartItem) => {
  try {
    await ElMessageBox.confirm('确定要删除该购物车商品吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await removeCartItem(row.sku_id, {
      user_id: row.user_id,
      sku_ids: [row.sku_id],
    })
    ElMessage.success('删除成功')
    fetchCartList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await updateCartQuantity(formData.value.sku_id, {
        user_id: formData.value.user_id,
        sku_id: formData.value.sku_id,
        quantity: formData.value.quantity,
      })
      
      // 更新选中状态
      await selectCartItem({
        user_id: formData.value.user_id,
        sku_id: formData.value.sku_id,
        is_selected: formData.value.is_selected,
      })
      
      ElMessage.success('更新成功')
      dialogVisible.value = false
      fetchCartList()
    } catch (error: any) {
      ElMessage.error(error.message || '更新失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  // 不自动加载，需要输入用户ID后查询
})
</script>

<style scoped>
.cart-list {
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

