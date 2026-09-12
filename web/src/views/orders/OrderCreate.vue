<template>
  <div class="order-page">
    <h1 class="page-title">确认订单</h1>

    <div class="order-layout">
      <!-- ======== 左侧：地址 + 商品 ======== -->
      <div class="order-main">
        <!-- 收货地址 -->
        <div class="card">
          <div class="card-head">
            <h3 class="card-title">收货地址</h3>
            <button class="btn-link" @click="showAddressDialog=true">选择地址</button>
          </div>
          <div v-if="selectedAddress" class="address-info">
            <p class="addr-name">{{ selectedAddress.receiver_name }} <span class="addr-phone">{{ selectedAddress.receiver_phone }}</span></p>
            <p class="addr-detail">{{ displayAddress(selectedAddress) }}</p>
          </div>
          <div v-else class="empty-block" @click="showAddressDialog=true">
            <p>请选择收货地址</p>
            <button class="btn-outline">选择地址</button>
          </div>
        </div>

        <!-- 订单商品 -->
        <div class="card">
          <h3 class="card-title">订单商品</h3>
          <div class="order-items">
            <div v-for="item in orderItems" :key="item.skuId" class="item-row">
              <div class="item-img">
                <img :src="getItemImage(item)" :alt="item.productName" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
              </div>
              <div class="item-info">
                <h4 class="item-name">{{ item.productName || '商品 '+item.skuId }}</h4>
                <span class="item-price">¥{{ fmt(item.price) }}</span>
              </div>
              <div class="item-qty">×{{ item.quantity }}</div>
              <div class="item-sub">¥{{ fmt((item.price||0)*item.quantity) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ======== 右侧：结算 ======== -->
      <div class="order-sidebar">
        <div class="checkout-card">
          <h3 class="card-title">订单摘要</h3>
          <div class="checkout-row"><span>商品合计</span><span>¥{{ fmt(totalAmount) }}</span></div>
          <div class="checkout-row"><span>运费</span><span>{{ freightAmount===0?'免运费':'¥'+fmt(freightAmount) }}</span></div>
          <div class="checkout-divider"></div>
          <div class="checkout-total">
            <span>应付总额</span>
            <span class="total-price">¥{{ fmt(totalAmount + freightAmount) }}</span>
          </div>
          <button
            class="btn-submit"
            :disabled="!selectedAddress || orderItems.length===0"
            :class="{loading:submitting}"
            @click="handleSubmit"
          >{{ submitting ? '提交中...' : '提交订单' }}</button>
          <p class="freight-hint" v-if="totalAmount<69 && orderItems.length>0">还差 ¥{{ fmt(69-totalAmount) }} 免运费</p>
        </div>
      </div>
    </div>

    <!-- ======== 地址选择弹窗 ======== -->
    <div v-if="showAddressDialog" class="modal-overlay" @click.self="showAddressDialog=false">
      <div class="modal-card">
        <h3 class="modal-title">{{ addressFormVisible ? (editingAddressId ? '编辑收货地址' : '新增收货地址') : '选择收货地址' }}</h3>

        <!-- 地址表单 -->
        <div v-if="addressFormVisible" class="address-form">
          <div class="form-row"><label>收货人</label><input v-model="addressForm.receiver_name" placeholder="姓名" /></div>
          <div class="form-row"><label>手机号</label><input v-model="addressForm.receiver_phone" placeholder="手机号" /></div>
          <div class="form-row"><label>省份</label><input v-model="addressForm.province" placeholder="省份" /></div>
          <div class="form-row"><label>城市</label><input v-model="addressForm.city" placeholder="城市" /></div>
          <div class="form-row"><label>区县</label><input v-model="addressForm.district" placeholder="区县" /></div>
          <div class="form-row"><label>详细地址</label><input v-model="addressForm.detail" placeholder="街道、门牌号等" /></div>
          <label class="default-check"><input type="checkbox" v-model="addressForm.is_default" /> 设为默认地址</label>
          <div class="modal-footer">
            <button class="btn-cancel" @click="cancelAddressForm">取消</button>
            <button class="btn-save" :class="{loading:addressSaving}" @click="saveAddress">{{ addressSaving ? '保存中...' : '保存' }}</button>
          </div>
        </div>

        <!-- 地址列表 -->
        <template v-else>
          <div v-if="addresses.length===0" class="empty-block" style="padding:40px 0">暂无收货地址，点下方「新增地址」添加</div>
          <div v-else class="address-list">
            <div
              v-for="addr in addresses" :key="addr.id"
              class="addr-option"
              :class="{selected:selectedAddressId===addr.id}"
              @click="selectedAddressId=addr.id"
            >
              <span class="radio-dot" :class="{checked:selectedAddressId===addr.id}"></span>
              <div class="addr-content">
                <p class="addr-name">{{ addr.receiver_name }} <span class="addr-phone">{{ addr.receiver_phone }}</span></p>
                <p class="addr-detail">{{ displayAddress(addr) }}</p>
              </div>
              <button class="btn-link addr-edit" @click.stop="openEditAddress(addr)">编辑</button>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-link" @click="openAddAddress">+ 新增地址</button>
            <div class="footer-right">
              <button class="btn-cancel" @click="showAddressDialog=false">取消</button>
              <button class="btn-save" @click="confirmAddress">确定</button>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder } from '@/api/order'
import { removeItem } from '@/api/cart'
import { getAddressList, addAddress, updateAddress } from '@/api/user'
import type { CartItem } from '@/api/cart'
import type { Address } from '@/api/user'

const route = useRoute(); const router = useRouter()
const cartStore = useCartStore(); const userStore = useUserStore()

const loading = ref(false); const submitting = ref(false)
const showAddressDialog = ref(false)
const selectedAddressId = ref<number|null>(null)
const selectedAddress = ref<Address|null>(null)
const addresses = ref<Address[]>([])
const orderItems = ref<CartItem[]>([])
const addressFormVisible = ref(false)
const editingAddressId = ref<number|null>(null)
const addressSaving = ref(false)
const addressForm = reactive({ receiver_name:'', receiver_phone:'', province:'', city:'', district:'', detail:'', is_default:false })

const placeholderUri = 'data:image/svg+xml,'+encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const resolveUrl = (url:string) => {if(!url)return'';if(url.startsWith('http'))return url;if(url.startsWith('/'))return'http://localhost:8080'+url;return url}
const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const getItemImage = (item:CartItem) => item.productImage?resolveUrl(item.productImage):placeholderUri
const displayAddress = (addr:any) => [addr.province,addr.city,addr.district,addr.detail].filter(Boolean).join('')||addr.receiver_address||''
const totalAmount = computed(() => orderItems.value.reduce((s,i)=>s+(i.price||0)*i.quantity,0))
const freightAmount = computed(() => totalAmount.value>=69?0:10)

const confirmAddress = () => {
  const addr = addresses.value.find(a=>a.id===selectedAddressId.value)
  if(addr){ selectedAddress.value=addr; showAddressDialog.value=false }
}

const handleSubmit = async () => {
  if(!selectedAddress.value){ElMessage.warning('请选择收货地址');return}
  if(orderItems.value.length===0){ElMessage.warning('订单商品不能为空');return}
  if(!userStore.userId){ElMessage.warning('请先登录');router.push('/login');return}
  submitting.value=true
  try{
    const addr = selectedAddress.value
    const addrStr = displayAddress(addr)
    const r = await createOrder({user_id:userStore.userId,items:orderItems.value.map(i=>({sku_id:i.skuId,quantity:i.quantity,product_name:i.productName||'',price:String(i.price||0)})),address_id:addr.id,receiver_name:addr.receiver_name,receiver_phone:addr.receiver_phone,receiver_address:addrStr,clear_cart:false,order_type:1})
    if(r.code===0&&r.data){
      ElMessage.success('订单创建成功')
      // 只移除已下单的商品，未选中的商品保留在购物车
      const orderedSkuIds = orderItems.value.map(i => i.skuId)
      try { await Promise.all(orderedSkuIds.map(skuId => removeItem(skuId))) } catch {}
      orderedSkuIds.forEach(skuId => cartStore.removeItem(skuId))
      router.push(`/orders/${r.data.id}`)
    }
    else ElMessage.error(r.message||'创建订单失败')
  }catch(e:any){ElMessage.error(e.message||'创建订单失败')}
  finally{submitting.value=false}
}

const loadAddresses = async (preferId?: number) => {
  try{
    const r = await getAddressList(userStore.userId)
    if(r.code===0 && r.data){
      addresses.value = r.data
      const d = preferId ? r.data.find((a:Address)=>a.id===preferId) : (r.data.find((a:Address)=>a.is_default===1) || r.data[0])
      if(d){ selectedAddressId.value = d.id; selectedAddress.value = d }
    }
  }catch{}
}

const openAddAddress = () => {
  editingAddressId.value = null
  Object.assign(addressForm, { receiver_name:'', receiver_phone:'', province:'', city:'', district:'', detail:'', is_default:false })
  addressFormVisible.value = true
}

const openEditAddress = (addr: Address) => {
  editingAddressId.value = addr.id
  Object.assign(addressForm, { receiver_name:addr.receiver_name, receiver_phone:addr.receiver_phone, province:addr.province, city:addr.city, district:addr.district, detail:addr.detail, is_default:addr.is_default===1 })
  addressFormVisible.value = true
}

const cancelAddressForm = () => {
  addressFormVisible.value = false
  editingAddressId.value = null
}

const saveAddress = async () => {
  if(!addressForm.receiver_name || !addressForm.receiver_phone || !addressForm.province || !addressForm.detail){
    ElMessage.warning('请填写收货人、手机号、省份和详细地址'); return
  }
  addressSaving.value = true
  try {
    const data:any = {
      user_id: userStore.userId,
      receiver_name: addressForm.receiver_name,
      receiver_phone: addressForm.receiver_phone,
      province: addressForm.province,
      city: addressForm.city,
      district: addressForm.district,
      detail: addressForm.detail,
      is_default: addressForm.is_default ? 1 : 0,
    }
    if (editingAddressId.value) {
      data.id = editingAddressId.value
      await updateAddress(editingAddressId.value, data)
      ElMessage.success('地址已更新')
      await loadAddresses(editingAddressId.value)
    } else {
      const r:any = await addAddress(data)
      if (r.code === 0) {
        ElMessage.success('地址已添加')
        await loadAddresses(r.data?.id)
      } else {
        ElMessage.error(r.message || '添加失败')
      }
    }
    addressFormVisible.value = false
    editingAddressId.value = null
  } catch(e:any) { ElMessage.error(e.message || '保存失败') }
  finally { addressSaving.value = false }
}
const loadOrderItems = async () => {
  // 立即购买：路由带 sku_id，只结算这一件（需要拉最新购物车拿到刚加购的这件）
  const buyNowSkuId = Number(route.query.sku_id || 0)
  if (buyNowSkuId) {
    await cartStore.fetchCart()
    const item = cartStore.cartItems.find(i => i.skuId === buyNowSkuId)
    if (item) {
      const qty = Number(route.query.quantity || 0)
      orderItems.value = [{ ...item, quantity: qty > 0 ? qty : item.quantity }]
      return
    }
  }
  // 购物车结算：尊重用户在购物车页的勾选状态（本地已有就不再重新拉取，避免勾选被后端默认值覆盖）
  if (cartStore.cartItems.length === 0) await cartStore.fetchCart()
  orderItems.value = cartStore.cartItems.filter(i => i.isSelected === 1).map(i => ({...i}))
}

onMounted(async () => {loading.value=true;try{await Promise.all([loadAddresses(),loadOrderItems()])}finally{loading.value=false}})
</script>

<style scoped>
.order-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.08); --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.02); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius:16px; --radius-sm:10px; max-width:1100px; margin:0 auto; padding:32px 24px; min-height:calc(100vh-64px); font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.page-title { font-size:28px; font-weight:700; margin:0 0 32px; color:var(--text); letter-spacing:-.01em; }

.order-layout { display:flex; gap:24px; align-items:flex-start; }
.order-main { flex:1; min-width:0; }

/* Card */
.card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:24px; margin-bottom:20px; }
.card-head { display:flex; justify-content:space-between; align-items:center; margin-bottom:16px; }
.card-title { font-size:16px; font-weight:600; margin:0; color:var(--text); }

/* Address */
.address-info p { margin:4px 0; line-height:1.5; }
.addr-name { font-size:15px; font-weight:600; color:var(--text); margin:0 0 6px; }
.addr-phone { font-weight:400; color:var(--text-dim); margin-left:8px; }
.addr-detail { font-size:13px; color:var(--text-dim); margin:0; }
.btn-link { background:none;border:none;color:var(--accent);font-size:13px;font-weight:500;cursor:pointer;padding:0; }
.btn-link:hover { opacity:.8; }
.btn-outline { padding:10px 24px;border-radius:100px;border:1px solid var(--accent);background:transparent;color:var(--accent);font-size:13px;font-weight:500;cursor:pointer;transition:all .2s; }
.btn-outline:hover { background:var(--accent-dim); }

/* Order Items */
.order-items { display:flex; flex-direction:column; }
.item-row { display:flex; align-items:center; gap:14px; padding:14px 0; border-bottom:1px solid var(--border); }
.item-row:last-child { border-bottom:none; }
.item-img { width:64px;height:64px;border-radius:var(--radius-sm);overflow:hidden;background:#111827;flex-shrink:0; }
.item-img img { width:100%;height:100%;object-fit:cover; }
.item-info { flex:1;min-width:0; }
.item-name { font-size:14px;font-weight:500;color:var(--text);margin:0 0 4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.item-price { font-size:13px;color:var(--accent);font-weight:600; }
.item-qty { font-size:14px;color:var(--text-dim);width:50px;text-align:center; }
.item-sub { font-size:15px;font-weight:600;color:var(--text);width:80px;text-align:right; }

/* Checkout Sidebar */
.order-sidebar { width:340px; flex-shrink:0; position:sticky; top:88px; }
.checkout-card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:24px; }
.checkout-row { display:flex; justify-content:space-between; padding:8px 0; font-size:14px; color:var(--text-dim); }
.checkout-divider { height:1px; background:var(--border); margin:12px 0; }
.checkout-total { display:flex; justify-content:space-between; align-items:center; padding:8px 0; }
.checkout-total span:first-child { font-size:16px; font-weight:600; color:var(--text); }
.total-price { font-size:28px; font-weight:800; color:var(--accent); letter-spacing:-.02em; }
.btn-submit { width:100%; padding:16px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:16px; font-weight:700; cursor:pointer; transition:all .2s; margin-top:20px; }
.btn-submit:hover:not(:disabled) { box-shadow:0 0 32px rgba(0,245,255,.35); transform:translateY(-2px); }
.btn-submit:disabled { opacity:.35; cursor:not-allowed; }
.btn-submit.loading { opacity:.7; }
.freight-hint { font-size:12px; color:var(--text-dim); text-align:center; margin:10px 0 0; }

/* Empty */
.empty-block { text-align:center; padding:24px 0; color:var(--text-dim); font-size:14px; cursor:pointer; }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.6); backdrop-filter:blur(4px); z-index:200; display:flex; align-items:center; justify-content:center; }
.modal-card { background:#111827; border:1px solid var(--border); border-radius:var(--radius); padding:28px; width:520px; max-height:80vh; overflow-y:auto; }
.modal-title { font-size:18px; font-weight:600; color:var(--text); margin:0 0 20px; }
.modal-footer { display:flex; justify-content:flex-end; gap:12px; margin-top:20px; }
.address-list { display:flex; flex-direction:column; gap:10px; }
.addr-option { display:flex; gap:12px; padding:14px; border:1px solid var(--border); border-radius:var(--radius-sm); cursor:pointer; transition:all .15s; }
.addr-option:hover { border-color:rgba(255,255,255,.15); }
.addr-option.selected { border-color:var(--accent); background:var(--accent-dim); }
.radio-dot { width:18px;height:18px;border-radius:50%;border:2px solid var(--border);flex-shrink:0;margin-top:2px;transition:all .15s; }
.radio-dot.checked { border-color:var(--accent);background:var(--accent);box-shadow:0 0 0 4px rgba(0,245,255,.15); }
.addr-content p { margin:0 0 4px; }
.addr-content .addr-name { font-size:14px;font-weight:600;color:var(--text); }
.addr-content .addr-phone { font-weight:400;color:var(--text-dim);margin-left:6px;font-size:13px; }
.addr-content .addr-detail { font-size:13px;color:var(--text-dim); }
.btn-cancel { padding:10px 24px;border-radius:100px;border:1px solid var(--border);background:transparent;color:var(--text);font-size:14px;cursor:pointer;transition:all .2s; }
.btn-cancel:hover { background:rgba(255,255,255,.04); }
.btn-save { padding:10px 28px;border-radius:100px;border:none;background:var(--accent);color:#0A0F1C;font-size:14px;font-weight:600;cursor:pointer; }
.btn-save:hover { box-shadow:0 0 20px rgba(0,245,255,.3); }
.addr-edit { flex-shrink:0; }
.address-form { display:flex; flex-direction:column; gap:12px; }
.form-row { display:flex; align-items:center; gap:10px; }
.form-row label { width:70px; font-size:13px; color:var(--text-dim); flex-shrink:0; }
.form-row input { flex:1; padding:9px 12px; border-radius:8px; background:rgba(255,255,255,.04); border:1px solid var(--border); color:var(--text); font-size:13px; outline:none; transition:border-color .2s; font-family:inherit; }
.form-row input:focus { border-color:var(--accent); }
.default-check { display:flex; align-items:center; gap:6px; font-size:13px; color:var(--text-dim); cursor:pointer; }
.footer-right { display:flex; gap:12px; }

@media(max-width:768px) { .order-layout { flex-direction:column; } .order-sidebar { width:100%;position:static; } }
</style>
