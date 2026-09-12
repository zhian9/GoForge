<template>
  <div class="detail-page" v-loading="loading">
    <div v-if="order" class="container">
      <h1 class="page-title">订单详情</h1>

      <!-- ======== 订单信息卡片 ======== -->
      <div class="info-card">
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">订单号</span>
            <span class="info-value mono">{{ order.order_no }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">订单状态</span>
            <span class="status-tag" :class="'status-'+order.status">{{ statusText(order.status) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">订单金额</span>
            <span class="info-value accent">¥{{ fmt(computedTotal) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">创建时间</span>
            <span class="info-value">{{ order.created_at }}</span>
          </div>
        </div>
      </div>

      <!-- ======== 操作 ======== -->
      <div class="actions-bar" v-if="order.status === 1 || order.status === 3">
        <button v-if="order.status === 1" class="btn-pay" @click="goPay">去支付</button>
        <button v-if="order.status === 3" class="btn-pay" @click="handleConfirmReceive">确认收货</button>
      </div>

      <!-- ======== 收货信息 ======== -->
      <div class="info-card" v-if="(order as any).receiver_name">
        <div class="info-grid cols-3">
          <div class="info-item">
            <span class="info-label">收货人</span>
            <span class="info-value">{{ (order as any).receiver_name }} {{ (order as any).receiver_phone }}</span>
          </div>
          <div class="info-item span-2">
            <span class="info-label">收货地址</span>
            <span class="info-value">{{ (order as any).receiver_address }}</span>
          </div>
        </div>
      </div>

      <!-- ======== 订单商品 ======== -->
      <div class="card">
        <h3 class="card-title">订单商品</h3>
        <div class="items-list">
          <div v-for="item in order.items" :key="item.id" class="item-row">
            <div class="item-img">
              <img :src="getProductImage(item)" :alt="getProductName(item)" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
            </div>
            <div class="item-info">
              <h4 class="item-name">{{ getProductName(item) }}</h4>
              <span class="item-price">¥{{ fmt(item.price) }}</span>
            </div>
            <div class="item-qty">×{{ item.quantity }}</div>
            <div class="item-sub">¥{{ fmt(item.total_price || item.total_amount || 0) }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getOrderDetail, confirmReceive } from '@/api/order'
import { getSkuDetail } from '@/api/sku'
import type { Order, OrderItem } from '@/api/order'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const order = ref<Order|null>(null)

const placeholderUri = 'data:image/svg+xml,'+encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const resolveUrl = (url:string) => {if(!url)return'';if(url.startsWith('http'))return url;if(url.startsWith('/'))return'http://localhost:8080'+url;return url}
const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const getField = (obj:any,...keys:string[])=>{for(const k of keys)if(obj&&k in obj&&obj[k]!==undefined&&obj[k]!==null)return obj[k];return undefined}

const statusMap:Record<number,string> = {0:'已取消',1:'待支付',2:'待发货',3:'待收货',4:'已完成',5:'已退款'}
const statusText = (s:number) => statusMap[s]||'未知'

const computedTotal = computed(() => {
  if (!order.value?.items?.length) return 0
  return order.value.items.reduce((sum:number, item:OrderItem) => {
    const p = typeof item.price === 'string' ? parseFloat(item.price) : (item.price || 0)
    return sum + p * (item.quantity || 0)
  }, 0)
})

const getProductImage = (item:OrderItem) => {
  // 优先 SKU 图片，回退产品图片
  const img = (item as any).sku_image || (item as any).skuImage || (item as any).product_image || (item as any).productImage || ''
  return img ? resolveUrl(img) : placeholderUri
}
const getProductName = (item:OrderItem) => {
  const pn = (item as any).product_name||(item as any).productName
  const sn = (item as any).sku_name||(item as any).skuName
  return pn||sn||`商品 ${item.product_id||(item as any).productId||'未知'}`
}

const loadOrderItemDetails = async () => {
  if(!order.value?.items?.length) return
  const promises = order.value.items.map(async (item:OrderItem) => {
    const hasName = !!(item.product_name||'')
    const priceVal = typeof item.price==='string'?parseFloat(item.price):(item.price||0)
    const hasPrice = priceVal>0
    const hasImage = !!((item as any).sku_image||(item as any).skuImage||'')
    if(hasName&&hasPrice&&hasImage) return
    const skuId = item.sku_id||getField(item as any,'skuId','SkuId')
    if(!skuId||skuId===0) return
    try{
      const r = await getSkuDetail(skuId)
      if(r.code===0&&r.data){
        const d = r.data as any
        const p = getField(d,'price','Price')||0
        const n = getField(d,'name','Name')||''
        const img = getField(d,'image','Image')||''
        if(!hasName&&n) item.product_name = n
        if(!hasPrice){const pn=typeof p==='number'?p:parseFloat(String(p||0));if(pn>0){item.price=String(pn.toFixed(2));const t=pn*(item.quantity||0);item.total_price=String(t.toFixed(2))}}
        if(img) (item as any).sku_image = img
      }
    }catch{}
  })
  await Promise.all(promises)
  if(order.value?.items?.length){
    let t=0;order.value.items.forEach((i:OrderItem)=>{const p=typeof i.price==='string'?parseFloat(i.price):(i.price||0);t+=p*(i.quantity||0)});order.value.total_amount=String(t.toFixed(2))
  }
}

const goPay = () => {
  if (order.value?.id) router.push(`/pay/${order.value.id}`)
}

const handleConfirmReceive = async () => {
  if (!order.value?.id) return
  try {
    await confirmReceive(order.value.id)
    ElMessage.success('已确认收货')
    fetchOrder()
  } catch (e: any) { ElMessage.error(e.message || '确认收货失败') }
}

const fetchOrder = async () => {
  loading.value=true
  try{
    const oid=Number(route.params.id);if(!oid||isNaN(oid))return
    const r = await getOrderDetail(oid)
    if(r.code===0&&r.data){
      const d = r.data as any
      order.value = {
        id:getField(d,'id','Id')||0,
        order_no:getField(d,'order_no','orderNo','OrderNo')||'',
        user_id:getField(d,'user_id','userId','UserId')||0,
        total_amount:String(getField(d,'total_amount','totalAmount','TotalAmount')||'0'),
        status:getField(d,'status','Status')||0,
        payment_status:getField(d,'payment_status','paymentStatus','PaymentStatus')||0,
        shipping_status:getField(d,'shipping_status','shippingStatus','ShippingStatus')||0,
        created_at:getField(d,'created_at','createdAt','CreatedAt')||'',
        updated_at:getField(d,'updated_at','updatedAt','UpdatedAt')||'',
        items:(d.items||[]).map((item:any)=>({
          id:getField(item,'id','Id')||0,order_id:getField(item,'order_id','orderId','OrderId')||0,
          product_id:getField(item,'product_id','productId','ProductId')||0,
          product_name:getField(item,'product_name','productName','ProductName')||'',
          sku_id:getField(item,'sku_id','skuId','SkuId')||0,
          sku_code:getField(item,'sku_code','skuCode','SkuCode')||'',
          sku_name:getField(item,'sku_name','skuName','SkuName')||'',
          sku_image:getField(item,'sku_image','skuImage','SkuImage')||'',
          sku_specs:getField(item,'sku_specs','skuSpecs','SkuSpecs')||{},
          quantity:getField(item,'quantity','Quantity')||0,
          price:String(getField(item,'price','Price')||'0'),
          total_price:String(getField(item,'total_price','totalAmount','TotalAmount')||'0'),
          total_amount:String(getField(item,'total_amount','totalAmount','TotalAmount')||'0'),
        } as OrderItem)),
      } as Order
      await loadOrderItemDetails()
    }
  }catch(e){console.error(e)}finally{loading.value=false}
}

watch(()=>route.params.id,(n,o)=>{if(n&&n!==o){order.value=null;fetchOrder()}})
onMounted(()=>fetchOrder())
</script>

<style scoped>
.detail-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.08); --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.02); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius:16px; --radius-sm:10px; min-height:calc(100vh-64px); padding:24px; font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.container { max-width:900px; margin:0 auto; }
.page-title { font-size:28px; font-weight:700; margin:0 0 32px; color:var(--text); letter-spacing:-.01em; }

/* Info Card */
.info-card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:24px; margin-bottom:20px; }
.info-grid { display:grid; grid-template-columns:repeat(3,1fr); gap:20px; }
.info-grid.cols-3 { grid-template-columns:repeat(3,1fr); }
.info-item { display:flex; flex-direction:column; gap:6px; }
.info-item.span-2 { grid-column:span 2; }
.info-label { font-size:12px; color:var(--text-dim); text-transform:uppercase; letter-spacing:.04em; }
.info-value { font-size:15px; color:var(--text); font-weight:500; word-break:break-all; }
.info-value.mono { font-family:'SF Mono','Fira Code',monospace; font-size:14px; }
.info-value.accent { font-size:20px; font-weight:700; color:var(--accent); letter-spacing:-.01em; }

/* Actions */
.actions-bar { display:flex; justify-content:flex-end; gap:12px; margin-bottom:20px; }
.btn-pay { padding:12px 40px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:15px; font-weight:700; cursor:pointer; transition:all .2s; }
.btn-pay:hover { box-shadow:0 0 24px rgba(0,245,255,.35); transform:translateY(-1px); }

/* Status Tags */
.status-tag { display:inline-block; padding:4px 14px; border-radius:100px; font-size:12px; font-weight:600; width:fit-content; }
.status-1,.pay-1,.ship-0 { color:#F59E0B; background:rgba(245,158,11,.1); }
.status-2,.pay-2,.ship-1 { color:#3B82F6; background:rgba(59,130,246,.1); }
.status-3,.pay-3,.ship-2 { color:var(--accent); background:var(--accent-dim); }
.status-4,.ship-3 { color:#10B981; background:rgba(16,185,129,.1); }
.status-5,.pay-4 { color:#F87171; background:rgba(248,113,113,.1); }

/* Items Card */
.card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:24px; margin-bottom:20px; }
.card-title { font-size:16px; font-weight:600; margin:0 0 20px; color:var(--text); }
.items-list { display:flex; flex-direction:column; }
.item-row { display:flex; align-items:center; gap:14px; padding:14px 0; border-bottom:1px solid var(--border); }
.item-row:last-child { border-bottom:none; }
.item-img { width:64px;height:64px;border-radius:var(--radius-sm);overflow:hidden;background:#111827;flex-shrink:0; }
.item-img img { width:100%;height:100%;object-fit:cover; }
.item-info { flex:1;min-width:0; }
.item-name { font-size:14px;font-weight:500;color:var(--text);margin:0 0 4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.item-price { font-size:13px;color:var(--accent);font-weight:600; }
.item-qty { font-size:14px;color:var(--text-dim);width:50px;text-align:center; }
.item-sub { font-size:15px;font-weight:600;color:var(--text);width:90px;text-align:right; }

@media(max-width:600px){ .info-grid { grid-template-columns:repeat(2,1fr); } .info-item.span-2 { grid-column:span 2; } }
</style>
