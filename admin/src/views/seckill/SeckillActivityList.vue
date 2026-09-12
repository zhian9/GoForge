<template>
  <div class="seckill-page">
    <PageHeader title="秒杀活动" desc="管理限时秒杀活动">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增活动
        </button>
      </template>
    </PageHeader>

    <!-- Filter -->
    <div class="search-bar">
      <select v-model="statusFilter" class="filter-select" @change="fetchList">
        <option :value="-1">全部状态</option>
        <option :value="0">未开始</option>
        <option :value="1">进行中</option>
        <option :value="2">已结束</option>
      </select>
      <label class="switch-wrap">
        <input type="checkbox" :checked="includeDisabled" @change="includeDisabled=!includeDisabled;fetchList()" />
        <span class="switch-slider"></span>
        <span class="switch-label">包含禁用</span>
      </label>
    </div>

    <!-- Table -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="list" class="dark-table" stripe>
        <el-table-column prop="id" label="ID" width="65" />
        <el-table-column prop="name" label="活动名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="SKU" min-width="200">
          <template #default="{ row }">
            <div class="sku-cell">
              <div class="img-cell" v-if="row.sku_image"><el-image :src="getImageUrl(row.sku_image)" fit="cover" class="thumb-img" /></div>
              <div class="sku-text">
                <span class="sku-name">{{ row.sku_name || '-' }}</span>
                <span class="sku-id">ID: {{ row.sku_id }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="160">
          <template #default="{ row }"><span class="price-accent">¥{{ fmt(row.seckill_price) }}</span><span class="price-old">¥{{ fmt(row.original_price) }}</span></template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="70" />
        <el-table-column prop="sold" label="已售" width="70" />
        <el-table-column label="时间" min-width="200">
          <template #default="{ row }"><div class="time-text">{{ fmtTime(row.start_time) }}</div><div class="time-to">至 {{ fmtTime(row.end_time) }}</div></template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }"><span class="status-tag" :class="'s-'+row.status">{{ statusText(row.status) }}</span></template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }"><span class="status-dot" :class="row.enable_status===1?'on':'off'">{{ row.enable_status===1?'启用':'禁用' }}</span></template>
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
      <el-pagination background layout="prev,pager,next" :current-page="page" :page-size="pageSize" :total="total" @current-change="handlePageChange" />
    </div>

    <!-- Edit Modal -->
    <EditModal v-model="dialogVisible" :title="dialogTitle" desc="配置秒杀活动参数" size="md" :loading="submitting" @confirm="handleSubmit" @cancel="dialogVisible=false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="活动名称" prop="name"><el-input v-model="form.name" placeholder="例如：手机秒杀专场" /></el-form-item>
        <el-form-item label="选择SKU" prop="sku_id">
          <el-select v-model="form.sku_id" filterable clearable placeholder="请选择 SKU" style="width:100%"><el-option v-for="s in skuOptions" :key="s.id" :label="`${s.name} (ID:${s.id})`" :value="s.id" /></el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="秒杀价" prop="seckill_price"><el-input v-model="form.seckill_price" placeholder="1999.00" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="库存" prop="stock"><el-input-number v-model="form.stock" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="开始时间" prop="start_time"><el-date-picker v-model="form.start_time" type="datetime" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束时间" prop="end_time"><el-date-picker v-model="form.end_time" type="datetime" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="启用状态">
          <el-radio-group v-model="form.enable_status"><el-radio :value="1">启用</el-radio><el-radio :value="0">禁用</el-radio></el-radio-group>
        </el-form-item>
        <p class="hint">保存时会将 Redis 秒杀库存重置为该值。</p>
      </el-form>
    </EditModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getSkuList, type Sku } from '@/api/sku'
import { getSeckillActivityList, createSeckillActivity, updateSeckillActivity, deleteSeckillActivity, type SeckillActivity } from '@/api/seckill'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'
import EditModal from '@/components/EditModal.vue'

const loading = ref(false); const list = ref<SeckillActivity[]>([])
const page = ref(1); const pageSize = ref(10); const total = ref(0)
const statusFilter = ref(-1); const includeDisabled = ref(true)
const skuOptions = ref<Sku[]>([])
const dialogVisible = ref(false); const submitting = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>()
const form = ref({ id:0, name:'', sku_id:undefined as number|undefined, seckill_price:'', stock:0, start_time:null as Date|null, end_time:null as Date|null, enable_status:1 })
const dialogTitle = computed(() => isEdit.value ? '编辑秒杀活动' : '新增秒杀活动')
const rules: FormRules = {
  name: [{ required:true, message:'请输入活动名称', trigger:'blur' }],
  sku_id: [{ required:true, message:'请选择SKU', trigger:'change' }],
  seckill_price: [{ required:true, message:'请输入秒杀价', trigger:'blur' }],
  stock: [{ required:true, message:'请输入库存', trigger:'change' }],
  start_time: [{ required:true, message:'请选择开始时间', trigger:'change' }],
  end_time: [{ required:true, message:'请选择结束时间', trigger:'change' }],
}

const fmt = (v:any) => { const n=Number(v||0); return isNaN(n)?'0.00':n.toFixed(2) }
const fmtTime = (s?:number) => { const sec=Number(s||0); if(!sec)return'-'; return new Date(sec*1000).toLocaleString() }
const getImageUrl = (url:string) => { if(!url)return''; if(url.startsWith('http'))return url; if(url.startsWith('/'))return`http://localhost:8080${url}`; return`http://localhost:8080/${url}` }
const statusText = (v?:number) => ({0:'未开始',1:'进行中',2:'已结束'}[v??-1]||'-')

const fetchSkus = async () => { try{const r = await getSkuList({page:1,page_size:200,status:1}) as any; skuOptions.value=r.data?.list||[]}catch{} }
const fetchList = async () => {
  loading.value=true
  try{const r = await getSeckillActivityList({page:page.value,page_size:pageSize.value,status:statusFilter.value,include_disabled:includeDisabled.value}) as any; list.value=r.data?.list||[]; total.value=Number(r.data?.total||0)}
  finally{loading.value=false}
}
const handlePageChange = (p:number) => { page.value=p; fetchList() }

const resetForm = () => { form.value={id:0,name:'',sku_id:undefined,seckill_price:'',stock:0,start_time:null,end_time:null,enable_status:1} }
const handleAdd = () => { isEdit.value=false; resetForm(); dialogVisible.value=true }
const handleEdit = (row:SeckillActivity) => { isEdit.value=true; form.value={id:row.id,name:row.name,sku_id:row.sku_id,seckill_price:String(row.seckill_price||''),stock:Number(row.stock||0),start_time:row.start_time?new Date(Number(row.start_time)*1000):null,end_time:row.end_time?new Date(Number(row.end_time)*1000):null,enable_status:row.enable_status??1}; dialogVisible.value=true }
const handleDelete = async (row:SeckillActivity) => { try{await ElMessageBox.confirm(`确定删除「${row.name}」？`,'提示',{type:'warning'});await deleteSeckillActivity(row.id);ElMessage.success('已删除');fetchList()}catch{} }

const handleSubmit = async () => {
  if(!formRef.value)return; await formRef.value.validate(async(valid)=>{if(!valid)return;if(!form.value.sku_id||!form.value.start_time||!form.value.end_time)return
  const p={name:form.value.name,sku_id:form.value.sku_id,seckill_price:String(form.value.seckill_price||''),stock:Number(form.value.stock||0),start_time:Math.floor(form.value.start_time.getTime()/1000),end_time:Math.floor(form.value.end_time.getTime()/1000),enable_status:Number(form.value.enable_status ?? 1)}
  submitting.value=true; try{if(isEdit.value){await updateSeckillActivity(form.value.id,p);ElMessage.success('已更新')}else{await createSeckillActivity(p);ElMessage.success('已创建')}dialogVisible.value=false;fetchList()}catch(e:any){ElMessage.error(e.message||'保存失败')}finally{submitting.value=false}
})}

onMounted(async()=>{await fetchSkus();fetchList()})
</script>

<style scoped>
.seckill-page { padding: 0; }
.btn-add { display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-add:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); transform: translateY(-1px); }
.btn-add svg { width: 16px; height: 16px; }

.search-bar { display: flex; gap: 12px; align-items: center; margin-bottom: 20px; }
.filter-select { appearance: none; -webkit-appearance: none; padding: 9px 32px 9px 14px; border-radius: 10px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; cursor: pointer; outline: none; font-family: inherit; background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; background-size: 16px; }
.filter-select option { background: #111827; color: #EDF0F5; }

.switch-wrap { display: inline-flex; align-items: center; gap: 10px; cursor: pointer; user-select: none; }
.switch-wrap input { display: none; }
.switch-slider { position: relative; width: 44px; height: 26px; border-radius: 26px; background: rgba(255,255,255,0.1); transition: background .2s; }
.switch-slider::after { content: ''; position: absolute; top: 3px; left: 3px; width: 20px; height: 20px; border-radius: 50%; background: #8890A5; transition: all .2s; }
.switch-wrap input:checked + .switch-slider { background: #00F5FF; }
.switch-wrap input:checked + .switch-slider::after { left: 21px; background: #0A0F1C; }
.switch-label { font-size: 13px; color: #8890A5; }

.sku-cell { display: flex; align-items: center; gap: 10px; }
.img-cell { width: 44px; height: 44px; border-radius: 6px; overflow: hidden; background: #111827; flex-shrink: 0; }
.thumb-img :deep(img) { object-fit: cover; }
.sku-text { display: flex; flex-direction: column; gap: 2px; }
.sku-name { font-weight: 600; color: #EDF0F5; font-size: 13px; }
.sku-id { font-size: 11px; color: #8890A5; }
.price-accent { color: #00F5FF; font-weight: 700; font-size: 14px; }
.price-old { color: #8890A5; text-decoration: line-through; font-size: 11px; margin-left: 6px; }
.time-to { color: #8890A5; font-size: 12px; }
.status-tag { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.s-0 { background: rgba(245,158,11,.12); color: #F59E0B; }
.s-1 { background: rgba(248,113,113,.12); color: #F87171; }
.s-2 { background: rgba(255,255,255,.06); color: #8890A5; }
.status-dot { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.status-dot.on { background: rgba(16,185,129,.12); color: #10B981; }
.status-dot.off { background: rgba(255,255,255,.06); color: #8890A5; }
.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; transition: all .15s; }
.tbl-btn.edit { color: #00F5FF; }
.tbl-btn.edit:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }
.pagination { display: flex; justify-content: center; margin-top: 20px; }
.hint { margin: 0; color: #64748B; font-size: 12px; }
</style>
