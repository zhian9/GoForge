<template>
  <div class="order-page">
    <PageHeader title="订单管理" desc="管理平台所有订单" />

    <!-- Filter -->
    <div class="search-bar">
      <select v-model="searchStatus" class="filter-select" @change="handleSearch">
        <option :value="-1">全部状态</option>
        <option :value="1">待支付</option>
        <option :value="2">待发货</option>
        <option :value="3">待收货</option>
        <option :value="4">已完成</option>
        <option :value="0">已取消</option>
      </select>
    </div>

    <!-- Table -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="orderList" class="dark-table" stripe>
        <el-table-column label="订单号" min-width="180">
          <template #default="{ row }"><span class="mono">{{ row.orderNo || row.order_no }}</span></template>
        </el-table-column>
        <el-table-column label="用户ID" width="80">
          <template #default="{ row }">{{ row.userId ?? row.user_id }}</template>
        </el-table-column>
        <el-table-column label="金额" width="110">
          <template #default="{ row }"><span class="price-accent">¥{{ fmt(calcRowTotal(row)) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-tag" :class="'s-'+ (row.status ?? 0)">{{ statusText(row.status ?? 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="收货人" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.receiverName || row.receiver_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="地址" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.receiverAddress || row.receiver_address || '-' }}</template>
        </el-table-column>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ fmtDate(row.createdAt ?? row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn" @click="handleView(row)">查看</button>
            <button v-if="(row.status??0)===2" class="tbl-btn ship" @click="handleShip(row)">发货</button>
            <button v-if="(row.status??0)===1" class="tbl-btn del" @click="handleCancel(row)">取消</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <div class="pagination" v-if="total>0">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @size-change="fetchOrderList" @current-change="fetchOrderList" />
    </div>

    <!-- Detail Modal -->
    <EditModal v-model="detailDialogVisible" title="订单详情" :desc="'订单号：'+(currentOrder?.orderNo||currentOrder?.order_no||'')" size="lg" hide-footer>
      <div v-if="currentOrder">
        <div class="info-grid">
          <div class="info-item"><span class="info-label">用户ID</span><span>{{ currentOrder.userId ?? currentOrder.user_id }}</span></div>
          <div class="info-item"><span class="info-label">状态</span><span class="status-tag" :class="'s-'+(currentOrder.status??0)">{{ statusText(currentOrder.status??0) }}</span></div>
          <div class="info-item"><span class="info-label">金额</span><span class="price-accent">¥{{ fmt(currentOrder.totalAmount ?? currentOrder.total_amount ?? 0) }}</span></div>
          <div class="info-item"><span class="info-label">收货人</span><span>{{ currentOrder.receiverName || currentOrder.receiver_name || '-' }}</span></div>
          <div class="info-item" style="grid-column:span 2"><span class="info-label">地址</span><span>{{ currentOrder.receiverAddress || currentOrder.receiver_address || '-' }}</span></div>
        </div>

        <h4 class="section-subtitle">商品明细</h4>
        <el-table :data="currentOrder.items||[]" class="dark-table" size="small">
          <el-table-column label="商品" min-width="180">
            <template #default="{ row }">{{ row.productName || row.product_name || '商品'+row.productId || '-' }}</template>
          </el-table-column>
          <el-table-column label="SKU" width="200" show-overflow-tooltip>
            <template #default="{ row }">{{ row.skuName || row.sku_name || 'SKU'+row.skuId || '-' }}</template>
          </el-table-column>
          <el-table-column label="单价" width="100">
            <template #default="{ row }"><span class="price-accent">¥{{ fmt(row.price ?? 0) }}</span></template>
          </el-table-column>
          <el-table-column label="数量" width="70">
            <template #default="{ row }">{{ row.quantity ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="小计" width="110">
            <template #default="{ row }">¥{{ fmt((row.price??0)*(row.quantity??0)) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </EditModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, getOrderDetail, cancelOrder, shipOrder, type Order } from '@/api/order'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'
import EditModal from '@/components/EditModal.vue'

const loading = ref(false); const orderList = ref<Order[]>([])
const currentPage = ref(1); const pageSize = ref(10); const total = ref(0)
const searchStatus = ref<number>(-1)
const detailDialogVisible = ref(false); const currentOrder = ref<Order|null>(null)

const fmt = (v:any) => { const n=Number(v||0); return isNaN(n)?'0.00':n.toFixed(2) }
const calcRowTotal = (row:any) => { if (!row?.items?.length) return 0; return row.items.reduce((s:number,i:any)=>s+(Number(i.price)||0)*(i.quantity||0),0) }
const fmtDate = (s:string) => { if(!s) return'-'; try{ return new Date(s).toLocaleString('zh-CN',{year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'}) }catch{return s} }
const statusMap: Record<number,string> = { 0:'已取消',1:'待支付',2:'待发货',3:'待收货',4:'已完成',5:'已退款' }
const statusText = (s:number) => statusMap[s]||'未知'

const fetchOrderList = async () => {
  loading.value = true
  try {
    const params: any = { page: currentPage.value, page_size: pageSize.value }
    params.status = searchStatus.value // -1 means all, otherwise specific status
    const r = await getOrderList(params) as any
    if (r.code===0) { orderList.value = r.data?.list || r.data?.orders || []; total.value = Number(r.data?.total || 0) }
  } catch { ElMessage.error('获取失败') } finally { loading.value = false }
}

const handleSearch = () => { currentPage.value = 1; fetchOrderList() }
const handleView = async (row: Order) => {
  try { const r = await getOrderDetail(row.id) as any; if (r.code===0) { currentOrder.value = r.data || r; detailDialogVisible.value = true } }
  catch { ElMessage.error('获取详情失败') }
}
const handleCancel = async (row: Order) => {
  try { await ElMessageBox.confirm('确定取消？', '提示', { type: 'warning' }); await cancelOrder(row.id); ElMessage.success('已取消'); fetchOrderList() }
  catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message) }
}
const handleShip = async (row: Order) => {
  try {
    const { value } = await ElMessageBox.prompt('物流公司编码（SF=顺丰 YTO=圆通 ZTO=中通 STO=申通 YD=韵达）', '发货', {
      confirmButtonText: '发货', cancelButtonText: '取消', inputValue: 'SF',
      inputPattern: /^[A-Za-z0-9]+$/, inputErrorMessage: '请输入物流公司编码',
    })
    await shipOrder(row.id, { company_code: value })
    ElMessage.success('发货成功')
    fetchOrderList()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '发货失败') }
}

watch(searchStatus, () => handleSearch())
onMounted(() => fetchOrderList())
</script>

<style scoped>
.order-page { padding: 0; }
.search-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.filter-select { appearance: none; -webkit-appearance: none; padding: 9px 32px 9px 14px; border-radius: 10px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.06); color: #EDF0F5; font-size: 13px; cursor: pointer; outline: none; font-family: inherit; background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; background-size: 16px; }
.filter-select option { background: #111827; color: #EDF0F5; }

.mono { font-family: 'SF Mono','Fira Code',monospace; font-size: 12px; }
.price-accent { color: #00F5FF; font-weight: 700; }
.status-tag { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.s-0 { background: rgba(255,255,255,.06); color: #8890A5; }
.s-1 { background: rgba(245,158,11,.12); color: #F59E0B; }
.s-2 { background: rgba(59,130,246,.12); color: #3B82F6; }
.s-3 { background: rgba(0,245,255,.12); color: #00F5FF; }
.s-4 { background: rgba(16,185,129,.12); color: #10B981; }
.s-5 { background: rgba(248,113,113,.12); color: #F87171; }
.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; color: #00F5FF; transition: all .15s; }
.tbl-btn:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }
.tbl-btn.ship { color: #10B981; }
.tbl-btn.ship:hover { background: rgba(16,185,129,.1); border-color: rgba(16,185,129,.2); }
.pagination { margin-top: 20px; display: flex; justify-content: flex-end; }

.info-grid { display: grid; grid-template-columns: repeat(2,1fr); gap: 14px; margin-bottom: 20px; }
.info-item { display: flex; flex-direction: column; gap: 4px; padding: 10px 14px; background: rgba(255,255,255,0.02); border-radius: 10px; }
.info-label { font-size: 11px; color: #8890A5; text-transform: uppercase; letter-spacing: .04em; }
.section-subtitle { font-size: 14px; font-weight: 600; color: #EDF0F5; margin: 0 0 12px; }
</style>
