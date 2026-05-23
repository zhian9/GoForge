<template>
  <div class="banner-page">
    <PageHeader title="Banner 管理" desc="管理首页轮播和广告位">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增 Banner
        </button>
      </template>
    </PageHeader>

    <DarkCard v-loading="loading" no-pad>
      <el-table :data="bannerList" class="dark-table" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="封面" width="140">
          <template #default="{ row }">
            <div class="img-cell" v-if="getBannerImage(row)">
              <el-image :src="getBannerImage(row)" :preview-src-list="[getBannerImage(row)]" fit="cover" class="thumb-img" />
            </div>
            <span v-else style="color:#8890A5;font-size:12px">无图片</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="链接" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.link" class="link-text">{{ row.link }}</span>
            <span v-else style="color:#8890A5">无</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <span class="link-tag" :class="'lt-'+row.link_type">{{ typeText(row.link_type) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-dot" :class="row.status===1?'on':'off'">{{ row.status===1?'启用':'禁用' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn edit" @click="handleEdit(row)">编辑</button>
            <button class="tbl-btn del" @click="handleDelete(row)">删除</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <!-- ======== 编辑弹窗 ======== -->
    <EditModal
      v-model="dialogVisible"
      :title="dialogTitle"
      desc="管理首页轮播 Banner"
      size="lg"
      :loading="submitting"
      @confirm="handleSubmit"
      @cancel="handleDialogClose"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="标题">
          <el-input v-model="formData.title" placeholder="Banner 标题" maxlength="50" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="吸引用户的描述文字" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="封面图片" prop="image_local">
          <div class="upload-row">
            <div class="upload-area" :class="{ hasImage: getBannerImageUrl(formData) }" @click="triggerUpload">
              <img v-if="getBannerImageUrl(formData)" :src="getBannerImageUrl(formData)" />
              <div v-else class="upload-placeholder">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                <span>上传封面图</span>
                <em>1226 × 400px</em>
              </div>
              <div class="upload-overlay" v-if="getBannerImageUrl(formData)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                <span>更换图片</span>
              </div>
            </div>
            <el-upload ref="uploadRef" class="hidden-upload" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleImageUploadSuccess" :before-upload="beforeUpload" accept="image/*" />
          </div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="链接类型">
              <select v-model="formData.link_type" class="form-select">
                <option :value="1">商品详情</option>
                <option :value="2">分类页面</option>
                <option :value="3">外部链接</option>
                <option :value="4">无链接</option>
              </select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="跳转链接" v-if="formData.link_type!==4">
              <el-input v-model="formData.link" placeholder="https://..." />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="排序"><el-input-number v-model="formData.sort" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <label class="switch-wrap">
                <input type="checkbox" :checked="formData.status===1" @change="formData.status=formData.status===1?0:1" />
                <span class="switch-slider"></span>
                <span class="switch-label">{{ formData.status===1 ? '启用' : '禁用' }}</span>
              </label>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="开始时间"><el-date-picker v-model="formData.start_time" type="datetime" placeholder="选填" style="width:100%" format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束时间"><el-date-picker v-model="formData.end_time" type="datetime" placeholder="选填" style="width:100%" format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item></el-col>
        </el-row>
      </el-form>
    </EditModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormRules, type UploadProps } from 'element-plus'
import { getBannerList, createBanner, updateBanner, deleteBanner } from '@/api/banner'
import type { Banner, CreateBannerRequest } from '@/api/banner'
import { useUserStore } from '@/stores/user'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'
import EditModal from '@/components/EditModal.vue'

const userStore = useUserStore()
const loading = ref(false); const submitting = ref(false)
const bannerList = ref<Banner[]>([]); const uploadRef = ref()
const triggerUpload = () => { const el = document.querySelector('.hidden-upload input[type=file]') as HTMLInputElement|null; el?.click() }
const dialogVisible = ref(false); const dialogTitle = ref('新增 Banner')
const formRef = ref()
const formData = ref<CreateBannerRequest & { id?: number; start_time?: string; end_time?: string }>({
  title: '', description: '', image: '', image_local: '', link: '', link_type: 4, sort: 0, status: 1, start_time: '', end_time: '',
})
const formRules: FormRules = { image_local: [{ required: true, message: '请上传封面图片', trigger: 'change' }] }

const uploadAction = computed(() => 'http://localhost:8080/api/v1/files/upload')
const uploadHeaders = computed(() => ({ Authorization: userStore.token ? `Bearer ${userStore.token}` : '' }))

const resolveUrl = (url: string) => { if (!url) return ''; if (url.startsWith('http')) return url; if (url.startsWith('/')) return 'http://localhost:8080' + url; return url }
const getBannerImageUrl = (b: any) => resolveUrl(b.image_local || b.image || '')
const getBannerImage = (b: Banner) => getBannerImageUrl(b)

const typeText = (t: number) => ({ 1: '商品', 2: '分类', 3: '外链', 4: '无' }[t] || '未知')

const fetchList = async () => {
  loading.value = true
  try { const r = await getBannerList({ status: -1 }) as any; if (r.code === 0 && r.data) bannerList.value = r.data } catch { ElMessage.error('获取失败') } finally { loading.value = false }
}

const handleAdd = () => {
  dialogTitle.value = '新增 Banner'; formData.value = { title: '', description: '', image: '', image_local: '', link: '', link_type: 4, sort: 0, status: 1, start_time: '', end_time: '' }; dialogVisible.value = true
}
const handleEdit = (row: Banner) => {
  dialogTitle.value = '编辑 Banner'; formData.value = { id: row.id, title: row.title || '', description: row.description || '', image: row.image || '', image_local: row.image_local || '', link: row.link || '', link_type: row.link_type, sort: row.sort, status: row.status, start_time: row.start_time || '', end_time: row.end_time || '' }; dialogVisible.value = true
}
const handleDelete = async (row: Banner) => {
  try { await ElMessageBox.confirm('确定删除？', '提示', { type: 'warning' }); const r = await deleteBanner(row.id) as any; if (r.code === 0) { ElMessage.success('已删除'); fetchList() } else ElMessage.error(r.message || '失败') }
  catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '删除失败') }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return; submitting.value = true
    try {
      const data: CreateBannerRequest = { title: formData.value.title, description: formData.value.description, image: formData.value.image, image_local: formData.value.image_local, link: formData.value.link_type === 4 ? '' : formData.value.link, link_type: formData.value.link_type, sort: formData.value.sort, status: formData.value.status, start_time: formData.value.start_time, end_time: formData.value.end_time }
      if (formData.value.id) { const r = await updateBanner(formData.value.id, data) as any; if (r.code === 0) { ElMessage.success('已更新'); dialogVisible.value = false; fetchList() } else ElMessage.error(r.message || '失败') }
      else { const r = await createBanner(data) as any; if (r.code === 0) { ElMessage.success('已创建'); dialogVisible.value = false; fetchList() } else ElMessage.error(r.message || '失败') }
    } catch { ElMessage.error('操作失败') } finally { submitting.value = false }
  })
}

const beforeUpload: UploadProps['beforeUpload'] = (file) => {
  if (!file.type.startsWith('image/')) { ElMessage.error('仅支持图片'); return false }
  if (file.size > 10 * 1024 * 1024) { ElMessage.error('不超过10MB'); return false }
  return true
}
const handleImageUploadSuccess = (res: any, file: File) => {
  if (res.data || res.code === 0) {
    const data = res.data; let url = data.file_url || data.fileUrl || ''
    if (!url && (data.file_id || data.fileId)) { const fid = data.file_id || data.fileId; const ext = file.name.substring(file.name.lastIndexOf('.')); url = `/uploads/image/${fid}${ext}` }
    if (url) { formData.value.image_local = url; formData.value.image = url; formData.value = { ...formData.value }; ElMessage.success('上传成功') } else ElMessage.error('无法获取图片URL')
  }
}
const handleDialogClose = () => formRef.value?.resetFields()

onMounted(() => fetchList())
</script>

<style scoped>
.banner-page { padding: 0; }

.btn-add { display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-add:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); transform: translateY(-1px); }
.btn-add svg { width: 16px; height: 16px; }

/* Table extras */
.img-cell { width: 100px; height: 56px; border-radius: 8px; overflow: hidden; background: #111827; }
.thumb-img { width: 100%; height: 100%; }
.thumb-img :deep(img) { object-fit: cover; transition: transform .3s; }
.thumb-img:hover :deep(img) { transform: scale(1.08); }
.link-text { color: #3B82F6; font-size: 12px; }
.link-tag { font-size: 11px; padding: 2px 8px; border-radius: 100px; font-weight: 600; }
.lt-1 { background: rgba(16,185,129,.12); color: #10B981; }
.lt-2 { background: rgba(245,158,11,.12); color: #F59E0B; }
.lt-3 { background: rgba(59,130,246,.12); color: #3B82F6; }
.lt-4 { background: rgba(255,255,255,.06); color: #8890A5; }
.status-dot { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.status-dot.on { background: rgba(16,185,129,.12); color: #10B981; }
.status-dot.off { background: rgba(255,255,255,.06); color: #8890A5; }

/* Table action buttons */
.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; transition: all .15s; }
.tbl-btn.edit { color: #00F5FF; }
.tbl-btn.edit:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }

/* Upload */
.upload-row { position: relative; }
.hidden-upload { display: none; }
.upload-area {
  width: 100%; height: 180px; border: 1px dashed rgba(255,255,255,.1); border-radius: 14px;
  display: flex; align-items: center; justify-content: center; cursor: pointer;
  transition: all .25s; overflow: hidden; position: relative;
  background: rgba(255,255,255,0.015);
}
.upload-area:hover { border-color: #00F5FF; background: rgba(0,245,255,0.04); }
.upload-area img { width: 100%; height: 100%; object-fit: cover; }
.upload-placeholder { display: flex; flex-direction: column; align-items: center; gap: 8px; color: #8890A5; }
.upload-placeholder svg { width: 40px; height: 40px; }
.upload-placeholder span { font-size: 14px; }
.upload-placeholder em { font-size: 11px; font-style: normal; opacity: .5; }
.upload-overlay {
  position: absolute; inset: 0; background: rgba(0,0,0,.55);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px;
  color: #fff; opacity: 0; transition: opacity .2s;
}
.upload-overlay svg { width: 24px; height: 24px; }
.upload-area:hover .upload-overlay { opacity: 1; }
.form-select {
  appearance: none; -webkit-appearance: none; width: 100%;
  padding: 10px 32px 10px 14px; border-radius: 10px;
  background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06);
  color: #EDF0F5; font-size: 13px; cursor: pointer; outline: none; font-family: inherit;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat; background-position: right 8px center; background-size: 16px;
}
.form-select:focus { border-color: #00F5FF; }
.form-select option { background: #111827; color: #EDF0F5; }
.switch-wrap { display: inline-flex; align-items: center; gap: 10px; cursor: pointer; user-select: none; }
.switch-wrap input { display: none; }
.switch-slider { position: relative; width: 44px; height: 26px; border-radius: 26px; background: rgba(255,255,255,0.1); transition: background .2s; }
.switch-slider::after { content: ''; position: absolute; top: 3px; left: 3px; width: 20px; height: 20px; border-radius: 50%; background: #8890A5; transition: all .2s; }
.switch-wrap input:checked + .switch-slider { background: #00F5FF; }
.switch-wrap input:checked + .switch-slider::after { left: 21px; background: #0A0F1C; }
.switch-label { font-size: 13px; color: #8890A5; }
</style>
