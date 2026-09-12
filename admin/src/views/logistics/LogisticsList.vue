<template>
  <div class="logistics-page">
    <PageHeader title="物流管理" desc="查看订单物流与轨迹" />

    <!-- 搜索 -->
    <div class="search-bar">
      <div class="search-box">
        <input v-model="searchOrderNo" type="text" placeholder="搜索订单号..." class="search-input" @keyup.enter="handleSearch" />
      </div>
      <button class="btn-search" @click="handleSearch">查询</button>
    </div>

    <!-- 表格 -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="logisticsList" class="dark-table" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column prop="logistics_no" label="物流单号" min-width="180" />
        <el-table-column prop="company_name" label="物流公司" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <span class="status-tag" :class="'ls-' + row.status">{{ statusText(row.status) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收货人" width="100" />
        <el-table-column label="当前位置" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.current_location || '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ fmtDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn" @click="handleTracking(row)">轨迹</button>
            <button class="tbl-btn ship" @click="handleUpdateStatus(row)">更新状态</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <div class="pagination" v-if="total > 0">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total,sizes,prev,pager,next" @size-change="fetchList" @current-change="fetchList" />
    </div>

    <!-- 物流轨迹弹窗 -->
    <el-dialog v-model="trackingVisible" title="物流轨迹" width="640px">
      <el-timeline v-if="trackingList.length">
        <el-timeline-item v-for="(t, i) in trackingList" :key="i" :timestamp="fmtDate(t.time)" :type="i === 0 ? 'primary' : undefined">
          <div class="track-node">
            <div class="track-status">{{ t.status }}</div>
            <div class="track-loc" v-if="t.location">{{ t.location }}</div>
            <div class="track-remark" v-if="t.remark">{{ t.remark }}</div>
          </div>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无轨迹" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listLogistics, queryTracking, updateLogisticsStatus, type Logistics, type TrackingNode } from '@/api/logistics'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'

const loading = ref(false)
const logisticsList = ref<Logistics[]>([])
const searchOrderNo = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const trackingVisible = ref(false)
const trackingList = ref<TrackingNode[]>([])

const statusMap: Record<number, string> = { 0: '待发货', 1: '已发货', 2: '运输中', 3: '已送达', 4: '异常' }
const statusText = (s: number) => statusMap[s] || '未知'

const fmtDate = (s: string) => {
  if (!s) return '-'
  try { return new Date(s).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) } catch { return s }
}

const fetchList = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (searchOrderNo.value) params.order_no = searchOrderNo.value
    const r = await listLogistics(params) as any
    if (r.code === 0) {
      logisticsList.value = r.data || []
      total.value = Number(r.total || 0)
    }
  } catch { ElMessage.error('获取物流列表失败') } finally { loading.value = false }
}

const handleSearch = () => { page.value = 1; fetchList() }

const handleTracking = async (row: Logistics) => {
  try {
    const r = await queryTracking(row.logistics_no) as any
    if (r.code === 0) {
      trackingList.value = Array.isArray(r.data) ? r.data : []
      trackingVisible.value = true
    } else {
      ElMessage.error(r.message || '查询物流轨迹失败')
    }
  } catch (e: any) { ElMessage.error(e.message || '查询物流轨迹失败') }
}

const handleUpdateStatus = async (row: Logistics) => {
  try {
    const { value } = await ElMessageBox.prompt('新状态（0待发货 1已发货 2运输中 3已送达 4异常）', '更新物流状态', {
      confirmButtonText: '确定', cancelButtonText: '取消',
      inputPattern: /^[0-4]$/, inputErrorMessage: '请输入 0-4 之间的数字',
    })
    await updateLogisticsStatus(row.logistics_no, parseInt(value))
    ElMessage.success('更新成功')
    fetchList()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '更新失败') }
}

onMounted(() => fetchList())
</script>

<style scoped>
.logistics-page { padding: 0; }
.search-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-box { position: relative; display: flex; align-items: center; flex: 1; max-width: 360px; }
.search-input { width: 100%; padding: 10px 16px; border-radius: 12px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; outline: none; transition: all .25s; font-family: inherit; }
.search-input:focus { border-color: #00F5FF; box-shadow: 0 0 0 3px rgba(0,245,255,.08); }
.btn-search { padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-search:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); }

.status-tag { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.ls-0 { background: rgba(255,255,255,.06); color: #8890A5; }
.ls-1 { background: rgba(245,158,11,.12); color: #F59E0B; }
.ls-2 { background: rgba(59,130,246,.12); color: #3B82F6; }
.ls-3 { background: rgba(16,185,129,.12); color: #10B981; }
.ls-4 { background: rgba(248,113,113,.12); color: #F87171; }

.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; color: #00F5FF; transition: all .15s; }
.tbl-btn:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.ship { color: #10B981; }
.tbl-btn.ship:hover { background: rgba(16,185,129,.1); border-color: rgba(16,185,129,.2); }

.pagination { margin-top: 20px; display: flex; justify-content: flex-end; }

.track-node { display: flex; flex-direction: column; gap: 2px; }
.track-status { font-size: 13px; font-weight: 600; color: #EDF0F5; }
.track-loc { font-size: 12px; color: #8890A5; }
.track-remark { font-size: 12px; color: #8890A5; }
</style>
