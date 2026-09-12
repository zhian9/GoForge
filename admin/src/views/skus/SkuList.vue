<template>
  <div class="sku-page">
    <PageHeader title="SKU 管理" desc="管理商品规格与库存">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增 SKU
        </button>
      </template>
    </PageHeader>

    <!-- 搜索 -->
    <div class="search-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input v-model="searchKeyword" type="text" placeholder="搜索 SKU 编码或名称..." class="search-input" @keyup.enter="handleSearch" />
      </div>
      <select v-model="searchProductId" class="filter-select" @change="handleSearch">
        <option :value="undefined">全部商品</option>
        <option v-for="p in productList" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
      <select v-model="searchStatus" class="filter-select narrow" @change="handleSearch">
        <option :value="-1">全部状态</option>
        <option :value="1">上架</option>
        <option :value="0">下架</option>
      </select>
    </div>

    <!-- 表格 -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="skuList" class="dark-table" stripe>
        <el-table-column prop="id" label="ID" width="65" />
        <el-table-column label="商品" width="140" show-overflow-tooltip>
          <template #default="{ row }"><span class="text-dim">{{ getProductName(row.product_id) }}</span></template>
        </el-table-column>
        <el-table-column prop="sku_code" label="SKU编码" width="130" />
        <el-table-column prop="name" label="SKU名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="规格" min-width="200">
          <template #default="{ row }">
            <div class="spec-tags">
              <span v-for="(v,k) in (row.specs||{})" :key="k" class="spec-tag">{{ k }}: {{ v }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="图片" width="80">
          <template #default="{ row }">
            <div class="img-cell" v-if="row.image"><el-image :src="getImageUrl(row.image)" :preview-src-list="[getImageUrl(row.image)]" fit="cover" class="thumb-img" /></div>
            <span v-else class="text-dim">-</span>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="110">
          <template #default="{ row }"><span class="price-accent">¥{{ fmt(row.price) }}</span><span v-if="row.original_price && row.original_price>row.price" class="price-old">¥{{ fmt(row.original_price) }}</span></template>
        </el-table-column>
        <el-table-column label="库存" width="80">
          <template #default="{ row }">
            <span class="stock-val" :class="row.stock<=10?'low':row.stock<=50?'mid':'ok'">{{ row.stock }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status" :active-value="1" :inactive-value="0" @change="(v:number)=>handleStatusChange(row,v)" :loading="row.statusUpdating" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn edit" @click="handleEdit(row)">编辑</button>
            <button class="tbl-btn del" @click="handleDelete(row)">删除</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <!-- 分页 -->
    <div class="pagination" v-if="pagination.total>0">
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @size-change="(s:number)=>{pagination.pageSize=s;pagination.page=1;loadSkuList()}" @current-change="(p:number)=>{pagination.page=p;loadSkuList()}" />
    </div>

    <!-- 对话框（逻辑不变） -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="760px" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="商品" prop="product_id">
          <el-select v-model="formData.product_id" placeholder="选择商品" style="width:100%" filterable :disabled="isEdit">
            <el-option v-for="p in productList" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="SKU编码" prop="sku_code">
          <el-input v-model="formData.sku_code" placeholder="唯一编码（自动生成）" :disabled="isEdit" @input="skuCodeTouched=true" />
        </el-form-item>
        <el-form-item label="SKU名称" prop="name"><el-input v-model="formData.name" placeholder="SKU名称" /></el-form-item>

        <!-- 规格编辑 -->
        <el-form-item label="规格" prop="specs">
          <div class="specs-editor">
            <div v-for="(k,i) in specKeys" :key="i" class="spec-row">
              <input v-model="specKeys[i]" placeholder="规格名" class="spec-input" @input="onSpecChange(i)" />
              <input v-model="specValues[i]" placeholder="规格值" class="spec-input" @input="onSpecChange(i)" />
              <button type="button" class="spec-del" @click="removeSpec(i)">×</button>
            </div>
            <button type="button" class="spec-add" @click="addSpec">+ 添加规格</button>
          </div>
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="价格" prop="price"><el-input-number v-model="formData.price" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="原价"><el-input-number v-model="formData.original_price" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="库存" prop="stock"><el-input-number v-model="formData.stock" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="重量(kg)"><el-input-number v-model="formData.weight" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="体积(m³)"><el-input-number v-model="formData.volume" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="状态"><el-radio-group v-model="formData.status"><el-radio :value="1">上架</el-radio><el-radio :value="0">下架</el-radio></el-radio-group></el-form-item></el-col>
        </el-row>

        <el-form-item label="SKU图片">
          <div class="upload-row">
            <el-upload class="uploader" :action="uploadAction" :headers="uploadHeaders" :show-file-list="false" :on-success="handleImageUploadSuccess" :before-upload="beforeUpload" accept="image/*">
              <img v-if="formData.image" :src="getImageUrl(formData.image)" class="upload-preview" />
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="upload-icon"><rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
            </el-upload>
            <el-button v-if="formData.image" link type="danger" @click="formData.image=''">删除图片</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getSkuList, createSku, updateSku, deleteSku, type Sku, type CreateSkuRequest } from '@/api/sku'
import { getProductList, type Product } from '@/api/product'
import { getToken } from '@/utils/auth'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'

const loading = ref(false)
const skuList = ref<(Sku & { statusUpdating?: boolean })[]>([])
const productList = ref<Product[]>([])
const searchKeyword = ref('')
const searchProductId = ref<number|undefined>()
const searchStatus = ref<number>(-1)
const pagination = reactive({ page:1, pageSize:20, total:0 })
const dialogVisible = ref(false); const dialogTitle = ref('新增SKU'); const isEdit = ref(false)
const submitting = ref(false); const formRef = ref<FormInstance>()
const formData = reactive<CreateSkuRequest & { id?: number }>({ product_id:0, sku_code:'', name:'', specs:{}, price:0, original_price:undefined, stock:0, image:'', weight:undefined, volume:undefined, status:1 })
const specKeys = ref<string[]>([]); const specValues = ref<string[]>([]); const skuCodeTouched = ref(false)

const uploadAction = `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/api/v1/files/upload`
const uploadHeaders = { Authorization: `Bearer ${getToken()}` }

const fmt = (v:any) => { const n = Number(v||0); return isNaN(n) ? '0.00' : n.toFixed(2) }
const getImageUrl = (url:string) => { if(!url)return''; if(url.startsWith('http'))return url; return `http://localhost:8080${url.startsWith('/')?'':'/'}${url}` }
const getProductName = (pid:number|undefined) => { const p = productList.value.find(x=>x.id===pid); return p?p.name:`ID:${pid}` }

const buildAutoSkuCode = () => {
  const pid = Number(formData.product_id||0)
  const entries = Object.entries(formData.specs||{}).filter(([k,v])=>k&&v).sort(([a],[b])=>a.localeCompare(b)).map(([k,v])=>`${k}=${v}`).join('|')
  let hash=5381; for(let i=0;i<entries.length;i++) hash=(hash*33)^entries.charCodeAt(i)
  return `P${pid}-${(hash>>>0).toString(16).slice(0,8)}`
}

const formRules: FormRules = {
  product_id: [{ required:true, message:'请选择商品', trigger:'change' }],
  sku_code: [{ required:true, message:'请输入SKU编码', trigger:'blur' }],
  name: [{ required:true, message:'请输入SKU名称', trigger:'blur' }],
  specs: [{ validator: (_r, v, cb) => { if (!v||Object.keys(v).length===0) cb(new Error('请至少添加一个规格')); else cb() }, trigger:'change' }],
  price: [{ required:true, message:'请输入价格', trigger:'blur' }],
  stock: [{ required:true, message:'请输入库存', trigger:'blur' }],
}

const loadProducts = async () => { try { const r = await getProductList({page:1,page_size:1000,status:-1}) as any; if(r.code===0&&r.data) productList.value=r.data.list||[] } catch {} }
const loadSkuList = async () => {
  loading.value=true; try {
    const params:any={page:pagination.page||1,page_size:pagination.pageSize||20}
    if(searchProductId.value) params.product_id=searchProductId.value; if(searchStatus.value>=-1) params.status=searchStatus.value
    const r = await getSkuList(params) as any
    if(r.code===0&&r.data){ skuList.value=(r.data.list||[]).map((s:any)=>({...s,product_id:s.product_id||s.productId,sku_code:s.sku_code||s.skuCode,specs:s.specs||{},original_price:s.original_price||s.originalPrice,image:s.image||'',statusUpdating:false})); pagination.total=Number(r.data.total||0) }
  } catch { ElMessage.error('加载失败') } finally { loading.value=false }
}

const handleSearch = () => { pagination.page=1; loadSkuList() }
const handleStatusChange = async (row:any, v:number) => { if(row.statusUpdating)return; row.statusUpdating=true; try{await updateSku(row.id,{status:v});Object.assign(row,{status:v});ElMessage.success('已更新')}catch(e:any){ElMessage.error(e.message)}finally{row.statusUpdating=false} }

const resetForm = () => { Object.assign(formData,{id:undefined,product_id:0,sku_code:'',name:'',specs:{},price:0,original_price:undefined,stock:0,image:'',weight:undefined,volume:undefined,status:1}); specKeys.value=[]; specValues.value=[]; skuCodeTouched.value=false; formRef.value?.clearValidate() }
const handleAdd = () => { dialogTitle.value='新增SKU'; isEdit.value=false; resetForm(); dialogVisible.value=true }
const handleEdit = (row:any) => { dialogTitle.value='编辑SKU'; isEdit.value=true; Object.assign(formData,{id:row.id,product_id:row.product_id,sku_code:row.sku_code,name:row.name,specs:row.specs||{},price:row.price,original_price:row.original_price,stock:row.stock,image:row.image||'',weight:row.weight,volume:row.volume,status:row.status}); specKeys.value=Object.keys(formData.specs); specValues.value=Object.values(formData.specs); dialogVisible.value=true }
const handleDelete = async (row:Sku) => { try{await ElMessageBox.confirm('确定删除？','提示',{type:'warning'});await deleteSku(row.id);ElMessage.success('已删除');loadSkuList()}catch(e:any){if(e!=='cancel')ElMessage.error(e.message)} }
const handleDialogClose = () => resetForm()

const addSpec = () => { specKeys.value.push(''); specValues.value.push('') }
const removeSpec = (i:number) => { specKeys.value.splice(i,1); specValues.value.splice(i,1); updateFormDataSpecs() }
const onSpecChange = (i:number) => updateFormDataSpecs()
const updateFormDataSpecs = () => {
  const s:Record<string,string>={}; specKeys.value.forEach((k,i)=>{if(k&&specValues.value[i])s[k]=specValues.value[i]}); formData.specs=s
  if(!isEdit.value&&!skuCodeTouched.value&&formData.product_id) formData.sku_code = buildAutoSkuCode()
}

const handleSubmit = async () => { if(!formRef.value)return; await formRef.value.validate(async(valid)=>{if(!valid)return;submitting.value=true;try{if(isEdit.value){await updateSku(formData.id!,formData);ElMessage.success('已更新')}else{const r=await createSku(formData);if(r.code===0){ElMessage.success('已创建')}else{ElMessage.error(r.message||'创建失败');return}}dialogVisible.value=false;if(!isEdit.value){pagination.page=1;searchProductId.value=undefined;searchStatus.value=-1}await loadSkuList()}catch(e:any){ElMessage.error(e.message)}finally{submitting.value=false}}) }

const beforeUpload = (f:File) => { if(!f.type.startsWith('image/')){ElMessage.error('仅图片');return false}if(f.size>10*1024*1024){ElMessage.error('不超过10MB');return false}return true }
const handleImageUploadSuccess = (res:any) => { if(res.code===0&&res.data){formData.image=res.data.file_url||res.data.file_path||'';ElMessage.success('上传成功')} }

onMounted(()=>{loadProducts();loadSkuList()})
</script>

<style scoped>
.sku-page { padding: 0; }
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
.filter-select.narrow { width: 120px; }
.filter-select option { background: #111827; color: #EDF0F5; }

.text-dim { color: #8890A5; font-size: 12px; }
.spec-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.spec-tag { padding: 2px 8px; border-radius: 6px; background: rgba(0,245,255,.08); color: #00F5FF; font-size: 11px; font-weight: 500; }
.img-cell { width: 48px; height: 48px; border-radius: 6px; overflow: hidden; background: #111827; }
.thumb-img :deep(img) { object-fit: cover; transition: transform .3s; }
.thumb-img:hover :deep(img) { transform: scale(1.1); }
.price-accent { color: #00F5FF; font-weight: 700; font-size: 14px; }
.price-old { color: #8890A5; text-decoration: line-through; font-size: 11px; margin-left: 6px; }
.stock-val { font-weight: 600; font-size: 14px; }
.stock-val.low { color: #F87171; }
.stock-val.mid { color: #F59E0B; }
.stock-val.ok { color: #10B981; }

.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; transition: all .15s; }
.tbl-btn.edit { color: #00F5FF; }
.tbl-btn.edit:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }

.pagination { margin-top: 20px; display: flex; justify-content: flex-end; }

.specs-editor { display: flex; flex-direction: column; gap: 8px; }
.spec-row { display: flex; gap: 8px; align-items: center; }
.spec-input { padding: 8px 12px; border-radius: 8px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; outline: none; width: 160px; font-family: inherit; transition: border-color .2s; }
.spec-input:focus { border-color: #00F5FF; }
.spec-del { width: 28px; height: 28px; border-radius: 6px; border: none; background: transparent; color: #F87171; font-size: 16px; cursor: pointer; }
.spec-add { padding: 6px 14px; border-radius: 8px; border: 1px solid rgba(0,245,255,.3); background: transparent; color: #00F5FF; font-size: 12px; cursor: pointer; width: fit-content; transition: all .15s; }
.spec-add:hover { background: rgba(0,245,255,.08); }

.upload-row { display: flex; gap: 12px; align-items: center; }
.uploader { width: 100px; height: 100px; border: 1px dashed rgba(255,255,255,.1); border-radius: 10px; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: border-color .2s; overflow: hidden; }
.uploader:hover { border-color: #00F5FF; }
.upload-preview { width: 100%; height: 100%; object-fit: cover; }
.upload-icon { width: 32px; height: 32px; color: #8890A5; }
</style>
