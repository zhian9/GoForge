<template>
  <div class="category-page">
    <PageHeader title="分类管理" desc="管理商品分类树">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增分类
        </button>
      </template>
    </PageHeader>

    <DarkCard v-loading="loading" no-pad>
      <el-tree
        :data="categoryTree"
        :props="{ children: 'children', label: 'name' }"
        default-expand-all
        node-key="id"
        :expand-on-click-node="false"
        class="dark-tree"
      >
        <template #default="{ data }">
          <div class="tree-row" :class="'level-'+data.level">
            <div class="tree-main">
              <span class="node-name">{{ data.name }}</span>
              <span class="node-meta">
                <span class="level-badge">L{{ data.level }}</span>
                <span class="status-tag" :class="data.status===1?'on':'off'">{{ data.status===1?'启用':'禁用' }}</span>
              </span>
            </div>
            <div class="tree-actions">
              <button class="act-btn add" @click.stop="handleAddChild(data)" title="添加子分类">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              </button>
              <button class="act-btn edit" @click.stop="handleEdit(data)" title="编辑">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button class="act-btn del" @click.stop="handleDelete(data)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
              </button>
            </div>
          </div>
        </template>
      </el-tree>
    </DarkCard>

    <!-- ======== 对话框 ======== -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="父分类">
          <el-tree-select v-model="formData.parentId" :data="categoryTree" :props="{children:'children',label:'name',value:'id'}" placeholder="不选即为顶级分类" check-strictly node-key="id" style="width:100%" clearable @change="handleParentChange" />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="formData.name" placeholder="输入分类名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="分类描述">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="输入描述（选填）" maxlength="200" show-word-limit />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="排序">
              <el-input-number v-model="formData.sort" :min="0" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="formData.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 图标上传 -->
        <el-form-item label="图标">
          <div class="upload-row">
            <el-upload class="uploader" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleIconUploadSuccess" :before-upload="beforeUpload" accept="image/*">
              <el-image v-if="formData.iconLocal||formData.icon" :src="formData.iconLocal||formData.icon" class="upload-preview" fit="cover" />
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="upload-icon"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </el-upload>
            <el-input v-model="formData.icon" placeholder="或输入图标URL" style="flex:1" />
          </div>
        </el-form-item>
        <el-form-item label="图片">
          <div class="upload-row">
            <el-upload class="uploader" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleImageUploadSuccess" :before-upload="beforeUpload" accept="image/*">
              <el-image v-if="formData.imageLocal||formData.image" :src="formData.imageLocal||formData.image" class="upload-preview" fit="cover" />
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="upload-icon"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </el-upload>
            <el-input v-model="formData.image" placeholder="或输入图片URL" style="flex:1" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadProps } from 'element-plus'
import { uploadUrl } from '@/utils/api'
import { getCategoryTree, createCategory, updateCategory, deleteCategory, type Category, type CreateCategoryRequest } from '@/api/category'
import { useUserStore } from '@/stores/user'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'

const loading = ref(false)
const categoryTree = ref<Category[]>([])
const userStore = useUserStore()
const uploadAction = computed(() => uploadUrl)
const uploadHeaders = computed(() => ({ Authorization: userStore.token ? `Bearer ${userStore.token}` : '' }))

const dialogVisible = ref(false); const dialogTitle = ref('新增分类'); const isEdit = ref(false)
const submitting = ref(false); const currentCategoryId = ref<number>(0)
const formRef = ref<FormInstance>()
const formData = ref<CreateCategoryRequest>({ parentId:0, name:'', description:'', sort:0, status:1 })
const formRules: FormRules = { name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }] }

const fetchCategoryTree = async () => {
  loading.value=true
  try { const r = await getCategoryTree({ status: -1 }) as any; if (r.code===0) categoryTree.value = r.data || [] }
  catch (e:any) { ElMessage.error(e.message||'获取失败') } finally { loading.value=false }
}

const handleAdd = () => { isEdit.value=false; dialogTitle.value='新增分类'; currentCategoryId.value=0; formData.value={ parentId:0, name:'', description:'', sort:0, status:1 }; dialogVisible.value=true }
const handleAddChild = (parent: Category) => { isEdit.value=false; dialogTitle.value='添加子分类'; currentCategoryId.value=0; formData.value={ parentId:parent.id, name:'', description:'', level:parent.level+1, sort:0, status:1 }; dialogVisible.value=true }
const handleEdit = (row: Category) => { isEdit.value=true; dialogTitle.value='编辑分类'; currentCategoryId.value=row.id; formData.value={ parentId:row.parentId, name:row.name, description:row.description||'', level:row.level, sort:row.sort, icon:row.icon, iconLocal:row.iconLocal, image:row.image, imageLocal:row.imageLocal, status:row.status }; dialogVisible.value=true }
const handleDelete = async (row: Category) => {
  try { await ElMessageBox.confirm('确定删除？不可恢复', '提示', { type: 'warning' }); await deleteCategory(row.id); ElMessage.success('已删除'); fetchCategoryTree() }
  catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '删除失败') }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value=true
    try {
      const pid = formData.value.parentId ? Number(formData.value.parentId) : 0
      const data: any = { parentId: pid, name: formData.value.name, status: formData.value.status || 1 }
      if (formData.value.description) data.description = formData.value.description
      if (formData.value.sort !== undefined) data.sort = formData.value.sort
      if (formData.value.icon) data.icon = formData.value.icon
      if ((formData as any).iconLocal) data.iconLocal = (formData as any).iconLocal
      if (formData.value.image) data.image = formData.value.image
      if ((formData as any).imageLocal) data.imageLocal = (formData as any).imageLocal
      if ((formData as any).level) data.level = (formData as any).level
      if (isEdit.value) { await updateCategory(currentCategoryId.value, data); ElMessage.success('已更新') }
      else { await createCategory(data); ElMessage.success('已创建') }
      dialogVisible.value=false; fetchCategoryTree()
    } catch (e: any) { ElMessage.error(e.message || '操作失败') }
    finally { submitting.value=false }
  })
}

const handleDialogClose = () => formRef.value?.resetFields()
const handleParentChange = (v: any) => { formData.value.parentId = v ? Number(v) : 0 }

const beforeUpload: UploadProps['beforeUpload'] = (file) => {
  if (!file.type.startsWith('image/')) { ElMessage.error('仅支持图片'); return false }
  if (file.size > 5*1024*1024) { ElMessage.error('不超过5MB'); return false }
  return true
}
const handleIconUploadSuccess = (res: any) => {
  if (res.code===0 && res.data) { (formData as any).iconLocal = res.data.file_url||res.data.file_id; if (!formData.value.icon) formData.value.icon = res.data.file_url||''; ElMessage.success('上传成功') }
}
const handleImageUploadSuccess = (res: any) => {
  if (res.code===0 && res.data) { (formData as any).imageLocal = res.data.file_url||res.data.file_id; if (!formData.value.image) formData.value.image = res.data.file_url||''; ElMessage.success('上传成功') }
}

onMounted(() => fetchCategoryTree())
</script>

<style scoped>
.category-page { padding: 0; }

.btn-add { display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-add:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); transform: translateY(-1px); }
.btn-add svg { width: 16px; height: 16px; }

/* Dark Tree Overrides */
.dark-tree { background: transparent !important; color: #EDF0F5; }
.dark-tree :deep(.el-tree-node__content) { height: auto; padding: 0; background: transparent !important; }
.dark-tree :deep(.el-tree-node__content:hover) { background: rgba(0,245,255,.04) !important; }
.dark-tree :deep(.el-tree-node__expand-icon) { color: #8890A5; padding: 6px; }
.dark-tree :deep(.el-tree-node__expand-icon:hover) { color: #00F5FF; }

/* Tree Row */
.tree-row { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px 10px 0; width: 100%; }
.tree-main { display: flex; align-items: center; gap: 12px; flex: 1; min-width: 0; }
.node-name { font-size: 14px; font-weight: 500; color: #EDF0F5; }
.node-meta { display: flex; align-items: center; gap: 8px; }
.level-badge { font-size: 10px; padding: 1px 6px; border-radius: 4px; background: rgba(0,245,255,.1); color: #00F5FF; font-weight: 600; }
.status-tag { font-size: 10px; padding: 2px 8px; border-radius: 100px; font-weight: 600; }
.status-tag.on { background: rgba(16,185,129,.12); color: #10B981; }
.status-tag.off { background: rgba(255,255,255,.06); color: #8890A5; }

/* Action Buttons */
.tree-actions { display: flex; gap: 4px; opacity: 0; transition: opacity .15s; }
.tree-row:hover .tree-actions { opacity: 1; }
.act-btn { width: 32px; height: 32px; border-radius: 8px; border: 1px solid transparent; background: transparent; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all .15s; }
.act-btn svg { width: 15px; height: 15px; }
.act-btn.add { color: #00F5FF; }
.act-btn.add:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.act-btn.edit { color: #8890A5; }
.act-btn.edit:hover { color: #F59E0B; background: rgba(245,158,11,.1); border-color: rgba(245,158,11,.2); }
.act-btn.del { color: #8890A5; }
.act-btn.del:hover { color: #F87171; background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }

/* Upload */
.upload-row { display: flex; gap: 12px; align-items: center; }
.uploader { width: 80px; height: 80px; border: 1px dashed rgba(255,255,255,.1); border-radius: 10px; display: flex; align-items: center; justify-content: center; cursor: pointer; flex-shrink: 0; transition: border-color .2s; }
.uploader:hover { border-color: #00F5FF; }
.upload-preview { width: 80px; height: 80px; border-radius: 8px; object-fit: cover; }
.upload-icon { width: 24px; height: 24px; color: #8890A5; }
</style>
