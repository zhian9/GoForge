<template>
  <div class="cart-page">
    <h1 class="page-title">购物车</h1>

    <!-- ======== 空购物车 ======== -->
    <div v-if="cartStore.cartItems.length === 0 && !loading" class="empty-block">
      <svg viewBox="0 0 120 120" fill="none" class="empty-icon">
        <circle cx="35" cy="85" r="6" stroke="currentColor" stroke-width="2.5"/><circle cx="85" cy="85" r="6" stroke="currentColor" stroke-width="2.5"/>
        <path d="M15 35h10l14 55h42l16-42H40" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M30 35 L42 10 H78 L90 35" stroke="currentColor" stroke-width="2" stroke-dasharray="4 3"/>
      </svg>
      <p class="empty-title">购物车是空的</p>
      <p class="empty-desc">快去挑选心仪的商品吧</p>
      <button class="btn-go" @click="$router.push('/products')">去逛逛</button>
    </div>

    <!-- ======== 有商品 ======== -->
    <div v-else class="cart-layout">
      <!-- 商品列表 -->
      <div class="cart-items">
        <!-- 全选栏 -->
        <div class="select-bar">
          <label class="check-all" @click="toggleSelectAll">
            <span class="checkbox" :class="{ checked: allSelected }">
              <svg v-if="allSelected" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
            </span>
            全选
          </label>
          <button class="btn-text danger" @click="handleClearAll" :disabled="cartStore.cartItems.length===0">清空购物车</button>
        </div>

        <!-- 商品卡片列表 -->
        <div v-for="item in cartStore.cartItems" :key="item.skuId" class="cart-item-card">
          <!-- 选择框 -->
          <span class="checkbox" :class="{ checked: item.isSelected===1 }" @click="toggleItem(item)"></span>
          <!-- 图片 -->
          <div class="item-img" @click="$router.push(`/products/${item.skuId}`)">
            <img :src="getItemImage(item)" :alt="item.productName" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
          </div>
          <!-- 信息 -->
          <div class="item-info">
            <h4 class="item-name" @click="$router.push(`/products/${item.skuId}`)">{{ item.productName || '商品 '+item.skuId }}</h4>
            <span class="item-price">¥{{ fmt(item.price) }}</span>
          </div>
          <!-- 数量 -->
          <div class="item-qty">
            <button class="qty-btn" @click="changeQty(item, -1)" :disabled="item.quantity<=1">−</button>
            <span class="qty-val">{{ item.quantity }}</span>
            <button class="qty-btn" @click="changeQty(item, 1)">+</button>
          </div>
          <!-- 小计 -->
          <div class="item-subtotal">
            <span class="subtotal-val">¥{{ fmt(item.price * item.quantity) }}</span>
          </div>
          <!-- 删除 -->
          <button class="btn-delete" @click="handleRemove(item)" title="删除">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
          </button>
        </div>
      </div>

      <!-- ======== 结算栏 ======== -->
      <div class="checkout-bar">
        <div class="checkout-card">
          <h3 class="checkout-title">订单摘要</h3>
          <div class="checkout-row">
            <span>商品数量</span><span>{{ selectedCount }} 件</span>
          </div>
          <div class="checkout-row" v-if="selectedCount !== cartStore.cartItems.reduce((s,i)=>s+i.quantity,0)">
            <span>总数量</span><span>{{ cartStore.cartItems.reduce((s,i)=>s+i.quantity,0) }} 件</span>
          </div>
          <div class="checkout-row">
            <span>商品金额</span><span class="accent">¥{{ fmt(selectedTotal) }}</span>
          </div>
          <div class="checkout-row">
            <span>运费</span><span>{{ selectedTotal>=69?'免运费':'¥10.00' }}</span>
          </div>
          <div class="checkout-divider"></div>
          <div class="checkout-total">
            <span>应付总额</span>
            <span class="total-price">¥{{ fmt(selectedTotal + (selectedTotal>=69?0:10)) }}</span>
          </div>
          <button class="btn-checkout" @click="handleCheckout" :disabled="selectedCount===0">
            去结算（{{ selectedCount }}件）
          </button>
          <p class="freight-hint" v-if="selectedTotal<69 && selectedCount>0">还差 ¥{{ fmt(69-selectedTotal) }} 免运费</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCartStore } from '@/stores/cart'
import { updateQuantity, removeItem, clearCart, selectItem, batchSelect } from '@/api/cart'
import type { CartItem } from '@/api/cart'

const router = useRouter()
const cartStore = useCartStore()
const loading = ref(false)

const placeholderUri = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const resolveUrl = (url: string) => {if(!url)return'';if(url.startsWith('http'))return url;if(url.startsWith('/'))return'http://localhost:8080'+url;return url}
const fmt = (n: any) => { const v = Number(n||0); return isNaN(v)?'0.00':v.toFixed(2) }
const getItemImage = (item: CartItem) => item.productImage ? resolveUrl(item.productImage) : placeholderUri

const selectedCount = computed(() => cartStore.cartItems.filter(i=>i.isSelected===1).reduce((s,i)=>s+i.quantity,0))
const selectedTotal = computed(() => cartStore.cartItems.filter(i=>i.isSelected===1).reduce((s,i)=>s+i.price*i.quantity,0))
const allSelected = computed(() => cartStore.cartItems.length>0 && cartStore.cartItems.every(i=>i.isSelected===1))

const toggleSelectAll = async () => {
  const newVal = allSelected.value ? 0 : 1
  cartStore.cartItems.forEach(i => i.isSelected = newVal)
  try { await batchSelect(cartStore.cartItems.map(i => i.skuId), newVal) } catch { ElMessage.error('操作失败') }
}

const toggleItem = async (item: CartItem) => {
  const newVal = item.isSelected === 1 ? 0 : 1
  item.isSelected = newVal
  try { await selectItem(item.skuId, newVal) } catch { item.isSelected = newVal === 1 ? 0 : 1; ElMessage.error('操作失败') }
}

const changeQty = async (item: CartItem, delta: number) => {
  const newQty = Math.max(1, item.quantity + delta)
  if (newQty === item.quantity) return
  try {
    await updateQuantity(item.skuId, { skuId: item.skuId, quantity: newQty })
    cartStore.updateItemQuantity(item.skuId, newQty)
  } catch { ElMessage.error('更新失败') }
}

const handleRemove = async (item: CartItem) => {
  try { await ElMessageBox.confirm('确定删除？','提示',{confirmButtonText:'确定',cancelButtonText:'取消',type:'warning'}); await removeItem(item.skuId); cartStore.removeItem(item.skuId); ElMessage.success('已删除') }
  catch(e:any) { if(e!=='cancel') ElMessage.error(e.message||'删除失败') }
}

const handleClearAll = async () => {
  if(cartStore.cartItems.length===0) return
  try { await ElMessageBox.confirm('确定清空购物车？','提示',{type:'warning'}); await clearCart(cartStore.cartItems[0]?.userId||0); cartStore.clearCart(); ElMessage.success('已清空') }
  catch(e:any) { if(e!=='cancel') ElMessage.error(e.message||'操作失败') }
}

const handleCheckout = () => {
  if (selectedCount.value===0) { ElMessage.warning('请先选择商品'); return }
  router.push('/orders/create')
}

onMounted(async () => { loading.value = true; try { await cartStore.fetchCart() } finally { loading.value = false } })
</script>

<style scoped>
.cart-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.08); --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.02); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius:14px; --radius-sm:10px; max-width:1100px; margin:0 auto; padding:32px 24px; min-height:calc(100vh-64px); font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.page-title { font-size:28px; font-weight:700; margin:0 0 32px; color:var(--text); letter-spacing:-.01em; }

/* Empty */
.empty-block { text-align:center; padding:100px 0; }
.empty-icon { width:120px; height:120px; color:var(--text-dim); opacity:.35; margin-bottom:24px; }
.empty-title { font-size:20px; font-weight:600; color:var(--text); margin:0 0 8px; }
.empty-desc { font-size:14px; color:var(--text-dim); margin:0 0 24px; }
.btn-go { padding:12px 36px; border-radius:100px; border:none; background:var(--accent); color:var(--bg); font-size:15px; font-weight:600; cursor:pointer; transition:all .2s; }
.btn-go:hover { box-shadow:0 0 24px rgba(0,245,255,.3); transform:translateY(-1px); }

/* Layout */
.cart-layout { display:flex; gap:24px; align-items:flex-start; }
.cart-items { flex:1; min-width:0; }

/* Select bar */
.select-bar { display:flex; justify-content:space-between; align-items:center; padding:0 0 16px; margin-bottom:16px; border-bottom:1px solid var(--border); }
.check-all { display:inline-flex; align-items:center; gap:10px; cursor:pointer; font-size:14px; color:var(--text); user-select:none; }
.btn-text { background:none; border:none; cursor:pointer; font-size:13px; padding:0; }
.btn-text.danger { color:#F87171; }
.btn-text:hover { opacity:.8; }
.btn-text:disabled { opacity:.3; cursor:not-allowed; }
.checkbox { width:22px; height:22px; border-radius:6px; border:2px solid var(--border); display:inline-flex; align-items:center; justify-content:center; cursor:pointer; transition:all .15s; flex-shrink:0; }
.checkbox:hover { border-color:var(--accent); }
.checkbox.checked { background:var(--accent); border-color:var(--accent); }
.checkbox svg { width:14px; height:14px; color:var(--bg); }
.checkbox.checked svg { display:block; }
.checkbox:not(.checked) svg { display:none; }

/* Item card */
.cart-item-card { display:flex; align-items:center; gap:14px; padding:16px; background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); margin-bottom:10px; transition:all .2s; }
.cart-item-card:hover { border-color:rgba(255,255,255,.1); }
.item-img { width:80px; height:80px; border-radius:var(--radius-sm); overflow:hidden; background:#111827; flex-shrink:0; cursor:pointer; }
.item-img img { width:100%;height:100%;object-fit:cover;transition:transform .4s; }
.item-img:hover img { transform:scale(1.08); }
.item-info { flex:1; min-width:0; }
.item-name { font-size:14px; font-weight:500; color:var(--text); margin:0 0 6px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; cursor:pointer; }
.item-name:hover { color:var(--accent); }
.item-price { font-size:14px; color:var(--accent); font-weight:600; }

/* Quantity */
.item-qty { display:flex; align-items:center; gap:0; background:rgba(255,255,255,0.04); border:1px solid var(--border); border-radius:100px; overflow:hidden; }
.qty-btn { width:34px; height:34px; border:none; background:transparent; color:var(--text); font-size:16px; cursor:pointer; transition:all .15s; display:flex;align-items:center;justify-content:center; }
.qty-btn:hover:not(:disabled) { background:var(--accent-dim); color:var(--accent); }
.qty-btn:disabled { opacity:.3;cursor:not-allowed; }
.qty-val { width:36px; text-align:center; font-size:14px; font-weight:600; color:var(--text); }

/* Subtotal */
.item-subtotal { width:90px; text-align:right; }
.subtotal-val { font-size:15px; font-weight:600; color:var(--text); }

/* Delete */
.btn-delete { width:36px; height:36px; border:none; background:transparent; color:var(--text-dim); cursor:pointer; border-radius:8px; transition:all .15s; display:flex;align-items:center;justify-content:center; flex-shrink:0; }
.btn-delete svg { width:18px; height:18px; }
.btn-delete:hover { color:#F87171; background:rgba(248,113,113,.1); }

/* ======== Checkout Bar ======== */
.checkout-bar { width:320px; flex-shrink:0; position:sticky; top:88px; }
.checkout-card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:24px; }
.checkout-title { font-size:16px; font-weight:600; color:var(--text); margin:0 0 20px; }
.checkout-row { display:flex; justify-content:space-between; align-items:center; padding:8px 0; font-size:14px; color:var(--text-dim); }
.checkout-row .accent { color:var(--accent); font-weight:600; }
.checkout-divider { height:1px; background:var(--border); margin:12px 0; }
.checkout-total { display:flex; justify-content:space-between; align-items:center; padding:8px 0; font-size:16px; font-weight:600; color:var(--text); }
.total-price { font-size:24px; color:var(--accent); font-weight:800; letter-spacing:-.01em; }
.btn-checkout { width:100%; padding:16px; border-radius:100px; border:none; background:var(--accent); color:var(--bg); font-size:16px; font-weight:700; cursor:pointer; transition:all .2s; margin-top:20px; }
.btn-checkout:hover:not(:disabled) { box-shadow:0 0 32px rgba(0,245,255,.35); transform:translateY(-2px); }
.btn-checkout:disabled { opacity:.35; cursor:not-allowed; }
.freight-hint { font-size:12px; color:var(--text-dim); text-align:center; margin:10px 0 0; }

@media(max-width:768px) { .cart-layout { flex-direction:column; } .checkout-bar { width:100%;position:static; } .item-subtotal { display:none; } .cart-item-card { flex-wrap:wrap; } }
</style>
