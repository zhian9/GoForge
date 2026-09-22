<template>
  <div class="product-page">
    <PageHeader title="商品管理" desc="管理平台商品">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增商品
        </button>
      </template>
    </PageHeader>

    <!-- 搜索 -->
    <div class="search-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input v-model="searchKeyword" type="text" placeholder="搜索商品名称..." class="search-input" @keyup.enter="handleSearch" />
      </div>
      <select v-model="searchCategory" class="filter-select" @change="handleSearch">
        <option :value="null">全部分类</option>
        <option v-for="cat in flatCategories" :key="cat.id" :value="cat.id">{{ '  '.repeat(Math.max(0,(cat.level||1)-1)) + cat.name }}</option>
      </select>
    </div>

    <!-- 表格 -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="productList" class="dark-table" stripe>
        <el-table-column type="selection" width="45" />
        <el-table-column prop="id" label="ID" width="65" />
        <el-table-column label="图片" width="100">
          <template #default="{ row }">
            <div class="img-cell" v-if="row.main_image || row.local_main_image">
              <el-image :src="getImageUrl(row.local_main_image || row.main_image)" :preview-src-list="[getImageUrl(row.local_main_image || row.main_image)]" fit="cover" class="thumb-img" />
            </div>
            <span v-else class="text-dim">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="商品名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="分类" width="110">
          <template #default="{ row }"><span class="text-dim">{{ getCategoryNameById(row.category_id) }}</span></template>
        </el-table-column>
        <el-table-column label="价格" width="110" sortable>
          <template #default="{ row }"><span class="price-accent">¥{{ fmt(row.price) }}</span></template>
        </el-table-column>
        <el-table-column label="原价" width="90">
          <template #default="{ row }"><span class="text-original" v-if="row.original_price">¥{{ fmt(row.original_price) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status" :active-value="1" :inactive-value="0" @change="(v:number)=>handleStatusChange(row,v)" :loading="row.statusUpdating" inline-prompt active-text="上架" inactive-text="下架" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="热门" width="80">
          <template #default="{ row }">
            <span class="hot-dot" :class="row.is_hot?'on':'off'">{{ row.is_hot?'热':'- ' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建" width="100">
          <template #default="{ row }">{{ (row.created_at||'').slice(0,10) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn edit" @click="handleEdit(row)">编辑</button>
            <button class="tbl-btn del" @click="handleDelete(row)">删除</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <div class="pagination" v-if="total>0">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @size-change="fetchProductList" @current-change="fetchProductList" />
    </div>

    <!-- 对话框（保持原有逻辑不变） -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="760px" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="商品名称" prop="name"><el-input v-model="formData.name" maxlength="100" /></el-form-item>
        <el-form-item label="商品描述"><el-input v-model="formData.description" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="价格" prop="price"><el-input-number v-model="formData.price" :min="0.01" :precision="2" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="原价"><el-input-number v-model="formData.original_price" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分类" prop="category_id">
              <el-tree-select v-model="formData.category_id" :data="categoryTree" :props="{value:'id',label:'name',children:'children'}" placeholder="选择分类" check-strictly node-key="id" style="width:100%" clearable @change="handleCategoryChange" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态"><el-radio-group v-model="formData.status"><el-radio :value="1">上架</el-radio><el-radio :value="0">下架</el-radio></el-radio-group></el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="主图">
          <div class="upload-row">
            <el-upload class="uploader" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleMainImageUploadSuccess" :before-upload="beforeUpload" accept="image/*">
              <img v-if="getImageUrl(formData.local_main_image||formData.main_image)" :src="getImageUrl(formData.local_main_image||formData.main_image)" class="upload-preview" />
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="upload-icon"><rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
            </el-upload>
            <el-input v-model="formData.main_image" placeholder="或输入主图URL" style="flex:1" />
          </div>
        </el-form-item>
        <el-form-item label="更多图片">
          <div class="img-grid">
            <div v-for="(img,i) in (formData.local_images||formData.images||[])" :key="i" class="img-item">
              <el-image :src="getImageUrl(img)" fit="cover" class="grid-thumb" :preview-src-list="(formData.local_images||formData.images||[]).map((u:string)=>getImageUrl(u))" />
              <button class="img-remove" @click="removeImage(i)">×</button>
            </div>
            <el-upload class="uploader grid-upload" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleImageUploadSuccess" :before-upload="beforeUpload" accept="image/*" multiple>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </el-upload>
          </div>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadProps } from 'element-plus'
import { resolveAssetUrl, uploadUrl } from '@/utils/api'
import { getProductList, createProduct, updateProduct, deleteProduct, type Product, type CreateProductRequest } from '@/api/product'
import { getCategoryTree, type Category } from '@/api/category'
import { useUserStore } from '@/stores/user'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'

const loading = ref(false); const productList = ref<Product[]>([]); const currentPage = ref(1); const pageSize = ref(10); const total = ref(0)
const searchKeyword = ref(''); const searchCategory = ref<number|null>(null)
const categoryTree = ref<Category[]>([]); const userStore = useUserStore()
const dialogVisible = ref(false); const dialogTitle = ref('新增商品'); const isEdit = ref(false)
const submitting = ref(false); const currentProductId = ref<number>(0); const formRef = ref<FormInstance>()
const formData = ref<CreateProductRequest & { images?: string[]; local_images?: string[]; description?: string }>({ name:'',description:'',price:0,original_price:0,category_id:0,status:1,is_hot:0,images:[],local_images:[] })

const uploadAction = computed(() => uploadUrl)
const uploadHeaders = computed(() => ({ Authorization: userStore.token ? `Bearer ${userStore.token}` : '' }))
const flatCategories = computed(() => { const r: Category[] = []; const f = (cats: Category[]) => cats.forEach(c => { r.push(c); if (c.children?.length) f(c.children) }); f(categoryTree.value); return r })

const fmt = (v:any) => { const n = Number(v||0); return isNaN(n) ? '0.00' : n.toFixed(2) }
const getImageUrl = resolveAssetUrl

const getCategoryNameById = (id: number|null|undefined) => {
  const nid = Number(id||0); if (!nid || !categoryTree.value.length) return '未分类'
  const find = (cats: Category[], tid: number): Category|null => { for (const c of cats) { if (Number(c.id)===tid) return c; if (c.children) { const f = find(c.children, tid); if (f) return f } } return null }
  const cat = find(categoryTree.value, nid); return cat ? cat.name : `ID:${nid}`
}

const formRules: FormRules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择分类', trigger: 'change' }, { validator: (_r:any, v:any, cb:any) => { if (!v || v===0 || v==='0') cb(new Error('请选择分类')); else cb() }, trigger: 'change' }],
}

const fetchProductList = async () => {
  loading.value=true
  try { const p:any = { page:currentPage.value, page_size:pageSize.value }; if (searchKeyword.value) p.keyword = searchKeyword.value; if (searchCategory.value) p.category_id = searchCategory.value
    const r = await getProductList(p) as any
    if (r.code===0) {
      productList.value = (r.data.list||[]).map((p:any) => ({ ...p, category_id: Number(p.category_id||p.categoryId||0), main_image: p.main_image||p.mainImage||'', local_main_image: p.local_main_image||p.localMainImage||'', hotUpdating:false, statusUpdating:false }))
      total.value = Number(r.data.total||0)
    }
  } catch { ElMessage.error('获取失败') } finally { loading.value=false }
}

const fetchCategoryTree = async () => { try { const r = await getCategoryTree({status:-1}) as any; if (r.code===0) categoryTree.value = r.data || [] } catch {} }

const handleSearch = () => { currentPage.value=1; fetchProductList() }
const handleCategoryChange = (v:any) => { formData.value.category_id = v ? Number(v) : 0 }

const beforeUpload: UploadProps['beforeUpload'] = (f) => { if (!f.type.startsWith('image/')) { ElMessage.error('仅图片'); return false } if (f.size>5*1024*1024) { ElMessage.error('不超过5MB'); return false } return true }
const uploadPic = (res:any, file:File, set: (url:string)=>void) => {
  if (res.data||res.code===0) { const d=res.data; let url=d.file_url||d.fileUrl||''; if (!url&&(d.file_id||d.fileId)) { const ext=file.name.substring(file.name.lastIndexOf('.')); url=`/uploads/image/${d.file_id||d.fileId}${ext}` } if (url) { set(url); ElMessage.success('上传成功') } }
}
const handleMainImageUploadSuccess = (res:any, file:File) => uploadPic(res, file, (url) => { formData.value.local_main_image=url; formData.value.main_image=url; formData.value={...formData.value} })
const handleImageUploadSuccess = (res:any, file:File) => uploadPic(res, file, (url) => { if(!formData.value.local_images) formData.value.local_images=[]; if(!formData.value.images) formData.value.images=[]; formData.value.local_images=[...formData.value.local_images,url]; formData.value.images=[...formData.value.images,url] })

const removeImage = (i:number) => { formData.value.images?.splice(i,1); formData.value.local_images?.splice(i,1); formData.value.images=[...(formData.value.images||[])]; formData.value.local_images=[...(formData.value.local_images||[])] }

const handleAdd = () => { isEdit.value=false; dialogTitle.value='新增商品'; currentProductId.value=0; formData.value={ name:'',description:'',price:0,original_price:0,category_id:0,status:1,is_hot:0,images:[],local_images:[] }; dialogVisible.value=true }
const handleEdit = (row: any) => {
  isEdit.value=true; dialogTitle.value='编辑商品'; currentProductId.value=row.id
  const imgs = Array.isArray(row.images) ? row.images : (typeof row.images==='string'?JSON.parse(row.images||'[]'):[])
  const limgs = Array.isArray(row.local_images) ? row.local_images : (typeof row.local_images==='string'?JSON.parse(row.local_images||'[]'):[])
  formData.value = { name:row.name, description:row.description||'', detail:row.detail||'', price:row.price, original_price:row.original_price||0, category_id:Number(row.category_id||0), status:row.status, is_hot:row.is_hot||0, main_image:row.main_image||'', local_main_image:row.local_main_image||'', images:imgs, local_images:limgs }
  dialogVisible.value=true
}
const handleDelete = async (row: Product) => {
  try { await ElMessageBox.confirm('确定删除？', '提示', { type:'warning' }); await deleteProduct(row.id); ElMessage.success('已删除'); fetchProductList() } catch(e:any) { if (e!=='cancel') ElMessage.error(e.message||'删除失败') }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => { if (!valid) return
    if (formData.value.price<=0) { ElMessage.error('价格必须大于0'); return }
    const cid = Number(formData.value.category_id); if (!cid||cid<=0) { ElMessage.error('请选择分类'); return }
    submitting.value=true
    try {
      const data: any = { name:formData.value.name, price:formData.value.price, category_id:cid, status:formData.value.status||1, is_hot:formData.value.is_hot||0, main_image:formData.value.main_image||'', local_main_image:formData.value.local_main_image||'', images:Array.isArray(formData.value.images)?formData.value.images:[], local_images:Array.isArray(formData.value.local_images)?formData.value.local_images:[] }
      if (formData.value.description) { data.detail = formData.value.description; data.description = formData.value.description }
      if ((formData.value.original_price ?? 0) > 0) data.original_price = formData.value.original_price
      if (isEdit.value) { await updateProduct(currentProductId.value, data); ElMessage.success('已更新') }
      else { await createProduct(data); ElMessage.success('已创建') }
      dialogVisible.value=false; await fetchCategoryTree(); fetchProductList()
    } catch(e:any) { ElMessage.error(e.message||'操作失败') } finally { submitting.value=false }
  })
}

const handleStatusChange = async (row: any, v: number) => {
  if (row.statusUpdating) return; row.statusUpdating=true
  try { await updateProduct(row.id, { category_id: row.category_id, status: v, is_hot: -1 } as any); Object.assign(row, { status: v }); ElMessage.success(v===1?'已上架':'已下架') }
  catch(e:any) { ElMessage.error(e.message||'更新失败') } finally { row.statusUpdating=false }
}

const handleDialogClose = () => formRef.value?.resetFields()

onMounted(() => { fetchCategoryTree().then(()=>fetchProductList()).catch(()=>fetchProductList()) })
</script>

<style scoped>
.product-page { padding: 0; }
.btn-add { display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-add:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); transform: translateY(-1px); }
.btn-add svg { width: 16px; height: 16px; }

.search-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-box { position: relative; display: flex; align-items: center; flex: 1; }
.search-icon { position: absolute; left: 12px; width: 16px; height: 16px; color: #8890A5; pointer-events: none; }
.search-input { width: 100%; padding: 10px 16px 10px 36px; border-radius: 12px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; outline: none; transition: all .25s; font-family: inherit; }
.search-input::placeholder { color: #8890A5; }
.search-input:focus { border-color: #00F5FF; box-shadow: 0 0 0 3px rgba(0,245,255,.08); }
.filter-select { appearance: none; -webkit-appearance: none; padding: 10px 32px 10px 14px; border-radius: 12px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; cursor: pointer; outline: none; background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; background-size: 16px; }
.filter-select option { background: #111827; color: #EDF0F5; }

.img-cell { width: 60px; height: 60px; border-radius: 8px; overflow: hidden; background: #111827; }
.thumb-img :deep(img) { object-fit: cover; transition: transform .3s; }
.thumb-img:hover :deep(img) { transform: scale(1.1); }
.price-accent { color: #00F5FF; font-weight: 700; font-size: 14px; }
.text-original { color: #8890A5; text-decoration: line-through; font-size: 12px; }
.text-dim { color: #8890A5; font-size: 12px; }
.hot-dot { font-size: 11px; padding: 2px 8px; border-radius: 100px; font-weight: 600; }
.hot-dot.on { background: rgba(245,158,11,.12); color: #F59E0B; }
.hot-dot.off { color: #8890A5; opacity: .4; }

.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; transition: all .15s; }
.tbl-btn.edit { color: #00F5FF; }
.tbl-btn.edit:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }

.pagination { margin-top: 20px; display: flex; justify-content: flex-end; }

.upload-row { display: flex; gap: 12px; align-items: flex-start; }
.uploader { width: 120px; height: 120px; border: 1px dashed rgba(255,255,255,.1); border-radius: 12px; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: border-color .2s; overflow: hidden; flex-shrink: 0; }
.uploader:hover { border-color: #00F5FF; }
.upload-preview { width: 100%; height: 100%; object-fit: cover; }
.upload-icon { width: 36px; height: 36px; color: #8890A5; }
.img-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.img-item { position: relative; width: 100px; height: 100px; border-radius: 8px; overflow: hidden; background: #111827; }
.grid-thumb { width: 100%; height: 100%; }
.grid-thumb :deep(img) { object-fit: cover; }
.img-remove { position: absolute; top: 2px; right: 2px; width: 22px; height: 22px; border-radius: 50%; border: none; background: rgba(0,0,0,.7); color: #F87171; font-size: 14px; cursor: pointer; display: flex; align-items: center; justify-content: center; opacity: 0; transition: opacity .15s; }
.img-item:hover .img-remove { opacity: 1; }
.grid-upload { width: 100px; height: 100px; background: rgba(255,255,255,.02); border: 1px dashed rgba(255,255,255,.1); border-radius: 8px; color: #8890A5; }
.grid-upload svg { width: 24px; height: 24px; }
.stock-hint { margin: 0 0 16px; color: #8890A5; font-size: 12px; line-height: 1.6; }
</style>
