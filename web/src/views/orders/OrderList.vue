<template>
  <div class="list-page" v-loading="loading">
    <h1 class="page-title">我的订单</h1>

    <!-- ======== 状态筛选 Tab ======== -->
    <div class="filter-tabs">
      <button v-for="f in filters" :key="f.value" class="tab-btn" :class="{active:statusFilter===f.value}" @click="statusFilter=f.value;fetchOrders()">{{ f.label }}</button>
    </div>

    <!-- ======== 空状态 ======== -->
    <div v-if="!loading && orders.length===0" class="empty-block">
      <svg viewBox="0 0 100 100" fill="none" class="empty-icon"><rect x="20" y="25" width="60" height="55" rx="8" stroke="currentColor" stroke-width="2.5"/><path d="M35 50h30M35 60h20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"/></svg>
      <p class="empty-title">暂无订单</p>
      <button class="btn-go" @click="$router.push('/products')">去逛逛</button>
    </div>

    <!-- ======== 订单列表 ======== -->
    <div v-else class="orders-list">
      <div v-for="order in orders" :key="order.id" class="order-card">
        <!-- 订单头部 -->
        <div class="order-head">
          <div class="order-head-left">
            <span class="order-no">订单号 {{ order.orderNo }}</span>
            <span class="order-time">{{ fmtTime(order.createdAt) }}</span>
          </div>
          <span class="status-tag" :class="'s-'+order.status">{{ statusText(order.status) }}</span>
        </div>

        <!-- 订单商品 -->
        <div class="order-items">
          <div v-for="item in order.items" :key="item.id" class="order-item">
            <div class="item-img">
              <img :src="getItemImage(item)" :alt="item.productName||'商品'" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
            </div>
            <div class="item-info">
              <h4 class="item-name">{{ item.productName || `商品 ${item.productId}` }}</h4>
              <span class="item-meta">¥{{ fmt(item.price) }} × {{ item.quantity }}</span>
            </div>
            <button class="btn-review" :class="{done:reviewedItemIds.has(Number(item.id))}" @click="openReview(order,item)" :disabled="reviewedItemIds.has(Number(item.id))">
              {{ reviewedItemIds.has(Number(item.id)) ? '已评价' : '去评价' }}
            </button>
          </div>
        </div>

        <!-- 订单底部 -->
        <div class="order-foot">
          <div class="order-total">合计 <span class="total-amount">¥{{ fmt(calcTotal(order)) }}</span></div>
          <div class="order-actions">
            <button v-if="order.status===1" class="btn-pay" @click="handlePay(order)">去支付</button>
            <button class="btn-detail" @click="$router.push(`/orders/${order.id}`)">查看详情</button>
            <button v-if="order.status===1" class="btn-cancel" @click="handleCancel(order)">取消订单</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ======== 分页 ======== -->
    <div v-if="total>pageSize" class="pagination">
      <button class="page-btn" :disabled="currentPage<=1" @click="currentPage--;fetchOrders()">←</button>
      <span class="page-info">{{ currentPage }} / {{ Math.ceil(total/pageSize) }}</span>
      <button class="page-btn" :disabled="currentPage>=Math.ceil(total/pageSize)" @click="currentPage++;fetchOrders()">→</button>
    </div>

    <!-- ======== 评价弹窗 ======== -->
    <div v-if="reviewDialogVisible" class="modal-overlay" @click.self="reviewDialogVisible=false">
      <div class="modal-card">
        <h3 class="modal-title">发表评价</h3>
        <div class="modal-body">
          <div class="star-row"><span class="star-label">评分</span><span v-for="i in 5" :key="i" class="star" :class="{filled:i<=reviewForm.rating}" @click="reviewForm.rating=i">★</span></div>
          <textarea v-model="reviewForm.content" class="modal-textarea" rows="4" maxlength="300" placeholder="写下你的真实体验吧～"></textarea>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="reviewDialogVisible=false">取消</button>
          <button class="btn-save" :disabled="reviewSubmitting" @click="submitReview">{{ reviewSubmitting?'提交中...':'提交评价' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, cancelOrder } from '@/api/order'
import { useUserStore } from '@/stores/user'
import { getSkuDetail } from '@/api/sku'
import type { Order } from '@/api/order'
import { createReview } from '@/api/review'

const router = useRouter()
const loading = ref(false)
const orders = ref<Order[]>([])
const statusFilter = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const filters = [
  {value:0,label:'全部'},{value:1,label:'待支付'},{value:2,label:'待发货'},
  {value:3,label:'待收货'},{value:4,label:'已完成'},
]

const reviewDialogVisible = ref(false); const reviewSubmitting = ref(false)
const reviewedItemIds = ref<Set<number>>(new Set())
const reviewForm = ref({order_id:0,order_item_id:0,product_id:0,sku_id:0,rating:5,content:''})

const placeholderUri = 'data:image/svg+xml,'+encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const resolveUrl = (url:string) => {if(!url)return'';if(url.startsWith('http'))return url;if(url.startsWith('/'))return'http://localhost:8080'+url;return url}
const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const fmtTime = (s:string) => {if(!s)return'';try{return new Date(s).toLocaleString('zh-CN',{year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})}catch{return s}}
const getItemImage = (item:any) => {
  const img = item?.skuImage || item?.sku_image || item?.sku_img
  return img ? resolveUrl(img) : placeholderUri
}

const calcTotal = (order:any) => {
  if (!order?.items?.length) return 0
  return order.items.reduce((s:number, i:any) => s + (Number(i.price)||0) * (i.quantity||0), 0)
}
const statusMap:Record<number,string> = {0:'已取消',1:'待支付',2:'待发货',3:'待收货',4:'已完成',5:'已退款'}
const statusText = (s:number) => statusMap[s]||'未知'

const fetchOrders = async () => {
  loading.value=true
  try{
    const us = useUserStore()
    const params:any = {user_id:us.userId,page:currentPage.value,page_size:pageSize.value}
    params.status = statusFilter.value===0?-1:statusFilter.value
    const r = await getOrderList(params) as any
    if(r.code===0){orders.value=r.data.list||[];total.value=r.data.total||0;enrichImages()}
  }catch(e){console.error(e)}finally{loading.value=false}
}

const enrichImages = async () => {
  for (const order of orders.value) {
    if (!order.items) continue
    for (const item of order.items as any[]) {
      if (item.skuImage || item.sku_image) continue
      try {
        const skuRes = await getSkuDetail(item.skuId || item.sku_id || 0)
        if (skuRes.code === 0 && skuRes.data) {
          const img = (skuRes.data as any).image || ''
          if (img) { item.sku_img = img }
        }
      } catch {}
    }
  }
}

const openReview = (order:any,item:any) => {
  const oid=Number(order?.id??0),iid=Number(item?.id??0),pid=Number(item?.productId??item?.product_id??0),sid=Number(item?.skuId??item?.sku_id??0)
  if(!oid||!iid||!pid||!sid){ElMessage.error('订单商品信息不完整');return}
  reviewForm.value={order_id:oid,order_item_id:iid,product_id:pid,sku_id:sid,rating:5,content:''}
  reviewDialogVisible.value=true
}

const submitReview = async () => {
  reviewSubmitting.value=true
  try{
    const us=useUserStore()
    await createReview({user_id:Number(us.userId),order_id:reviewForm.value.order_id,order_item_id:reviewForm.value.order_item_id,product_id:reviewForm.value.product_id,sku_id:reviewForm.value.sku_id,rating:Number(reviewForm.value.rating),content:reviewForm.value.content,images:[],videos:[]})
    reviewedItemIds.value.add(Number(reviewForm.value.order_item_id));ElMessage.success('评价已提交');reviewDialogVisible.value=false
  }catch(e:any){ElMessage.error(e?.message||'提交失败')}
  finally{reviewSubmitting.value=false}
}

const handlePay = (order:Order) => router.push(`/pay/${order.id}`)
const handleCancel = async (order:Order) => {
  try{await ElMessageBox.confirm('确定取消该订单？','提示',{type:'warning'});await cancelOrder(order.id);ElMessage.success('已取消');fetchOrders()}
  catch(e:any){if(e!=='cancel')ElMessage.error(e.message||'取消失败')}
}

onMounted(()=>fetchOrders())
</script>

<style scoped>
.list-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.08); --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.025); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius:16px; --radius-sm:10px; max-width:960px; margin:0 auto; padding:32px 24px; min-height:calc(100vh-64px); font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.page-title { font-size:28px; font-weight:700; margin:0 0 28px; color:var(--text); }

/* Filter Tabs */
.filter-tabs { display:flex; gap:6px; margin-bottom:28px; }
.tab-btn { padding:9px 22px; border-radius:100px; border:1px solid var(--border); background:transparent; color:var(--text-dim); font-size:13px; font-weight:500; cursor:pointer; transition:all .15s; }
.tab-btn:hover { color:var(--text); border-color:rgba(255,255,255,.15); }
.tab-btn.active { background:var(--accent); border-color:var(--accent); color:#0A0F1C; font-weight:600; }

/* Empty */
.empty-block { text-align:center; padding:80px 0; }
.empty-icon { width:100px;height:100px;color:var(--text-dim);opacity:.3;margin-bottom:20px; }
.empty-title { font-size:18px;color:var(--text-dim);margin:0 0 20px; }
.btn-go { padding:12px 32px;border-radius:100px;border:none;background:var(--accent);color:var(--bg);font-size:15px;font-weight:600;cursor:pointer; }
.btn-go:hover { box-shadow:0 0 24px rgba(0,245,255,.3); }

/* Order Cards */
.orders-list { display:flex; flex-direction:column; gap:16px; }
.order-card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:20px 24px; transition:border-color .2s; }
.order-card:hover { border-color:rgba(255,255,255,.1); }

/* Head */
.order-head { display:flex; justify-content:space-between; align-items:center; padding-bottom:16px; border-bottom:1px solid var(--border); margin-bottom:8px; }
.order-head-left { display:flex; align-items:center; gap:16px; }
.order-no { font-size:13px; color:var(--text); font-family:'SF Mono','Fira Code',monospace; }
.order-time { font-size:12px; color:var(--text-dim); }

/* Status tags */
.status-tag { display:inline-block; padding:4px 14px; border-radius:100px; font-size:12px; font-weight:600; }
.s-0 { color:#F87171; background:rgba(248,113,113,.1); }
.s-1 { color:#F59E0B; background:rgba(245,158,11,.1); }
.s-2 { color:#3B82F6; background:rgba(59,130,246,.1); }
.s-3 { color:var(--accent); background:var(--accent-dim); }
.s-4 { color:#10B981; background:rgba(16,185,129,.1); }
.s-5 { color:#F87171; background:rgba(248,113,113,.1); }

/* Items */
.order-items { padding:8px 0; }
.order-item { display:flex; align-items:center; gap:14px; padding:10px 0; border-bottom:1px solid rgba(255,255,255,.03); }
.order-item:last-child { border-bottom:none; }
.item-img { width:64px;height:64px;border-radius:var(--radius-sm);overflow:hidden;background:#111827;flex-shrink:0; }
.item-img img { width:100%;height:100%;object-fit:cover; }
.item-info { flex:1;min-width:0; }
.item-name { font-size:14px;font-weight:500;color:var(--text);margin:0 0 4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.item-meta { font-size:12px;color:var(--text-dim); }
.btn-review { padding:6px 14px;border-radius:100px;border:1px solid var(--accent);background:transparent;color:var(--accent);font-size:12px;cursor:pointer;transition:all .15s;flex-shrink:0; }
.btn-review:hover:not(:disabled) { background:var(--accent-dim); }
.btn-review.done { border-color:var(--border);color:var(--text-dim);cursor:default; }

/* Footer */
.order-foot { display:flex; justify-content:space-between; align-items:center; padding-top:16px; border-top:1px solid var(--border); }
.order-total { font-size:15px;color:var(--text-dim); }
.total-amount { font-size:22px;font-weight:700;color:var(--accent);margin-left:8px; }
.order-actions { display:flex; gap:10px; }
.btn-pay,.btn-detail,.btn-cancel { padding:9px 20px;border-radius:100px;font-size:13px;font-weight:500;cursor:pointer;transition:all .15s; }
.btn-pay { border:none;background:var(--accent);color:#0A0F1C;font-weight:600; }
.btn-pay:hover { box-shadow:0 0 20px rgba(0,245,255,.3); }
.btn-detail { border:1px solid var(--border);background:transparent;color:var(--text); }
.btn-detail:hover { border-color:rgba(255,255,255,.2);background:rgba(255,255,255,.04); }
.btn-cancel { border:1px solid transparent;background:transparent;color:var(--text-dim); }
.btn-cancel:hover { color:#F87171; }

/* Pagination */
.pagination { display:flex; justify-content:center; align-items:center; gap:16px; margin-top:32px; }
.page-btn { width:40px;height:38px;border-radius:10px;border:1px solid var(--border);background:var(--card-bg);color:var(--text);font-size:16px;cursor:pointer;transition:all .15s; }
.page-btn:hover:not(:disabled) { border-color:var(--accent);color:var(--accent); }
.page-btn:disabled { opacity:.3;cursor:not-allowed; }
.page-info { font-size:14px;color:var(--text-dim);min-width:80px;text-align:center; }

/* Modal */
.modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.6);backdrop-filter:blur(4px);z-index:200;display:flex;align-items:center;justify-content:center; }
.modal-card { background:#111827;border:1px solid var(--border);border-radius:var(--radius);padding:28px;width:480px; }
.modal-title { font-size:18px;font-weight:600;color:var(--text);margin:0 0 20px; }
.modal-body { display:flex;flex-direction:column;gap:16px; }
.star-row { display:flex;align-items:center;gap:8px; }
.star-label { font-size:13px;color:var(--text-dim);margin-right:8px; }
.star { font-size:26px;color:rgba(255,255,255,.08);cursor:pointer;transition:all .1s; }
.star:hover { transform:scale(1.15); }
.star.filled { color:#FBBF24; }
.modal-textarea { width:100%;padding:12px;border-radius:var(--radius-sm);background:rgba(255,255,255,.04);border:1px solid var(--border);color:var(--text);font-size:13px;resize:vertical;outline:none;font-family:inherit;box-sizing:border-box; }
.modal-textarea:focus { border-color:var(--accent); }
.modal-textarea::placeholder { color:var(--text-dim); }
.modal-footer { display:flex;justify-content:flex-end;gap:12px;margin-top:20px; }
.btn-cancel { padding:10px 24px;border-radius:100px;border:1px solid var(--border);background:transparent;color:var(--text);font-size:14px;cursor:pointer; }
.btn-cancel:hover { background:rgba(255,255,255,.04); }
.btn-save { padding:10px 28px;border-radius:100px;border:none;background:var(--accent);color:#0A0F1C;font-size:14px;font-weight:600;cursor:pointer; }
.btn-save:hover:not(:disabled) { box-shadow:0 0 20px rgba(0,245,255,.3); }
.btn-save:disabled { opacity:.5;cursor:not-allowed; }
</style>
