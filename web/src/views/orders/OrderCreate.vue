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
                <img :src="getItemImage(item)" :alt="item.productName" @error="e=>(e.target as HTMLImageElement).src=thumbPlaceholderUri" />
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
          <div class="checkout-row">
            <span>优惠券</span>
            <el-select
              v-model="selectedUserCouponId"
              size="small"
              placeholder="不使用优惠券"
              style="width:180px"
              @change="handleCouponChange"
            >
              <el-option label="不使用优惠券" :value="0" />
              <el-option
                v-for="uc in usableCoupons"
                :key="uc.id"
                :label="couponLabel(uc)"
                :value="uc.id"
              />
            </el-select>
          </div>
          <div class="checkout-row" v-if="couponDiscount > 0">
            <span>优惠</span><span class="discount-amount">-¥{{ fmt(couponDiscount) }}</span>
          </div>
          <div class="checkout-divider"></div>
          <div class="checkout-total">
            <span>应付总额</span>
            <span class="total-price">¥{{ fmt(payableAmount) }}</span>
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
import { resolveAssetUrl as resolveUrl } from '@/utils/api'
import { ElMessage } from 'element-plus'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder } from '@/api/order'
import { removeItem } from '@/api/cart'
import { getAddressList, addAddress, updateAddress } from '@/api/user'
import type { CartItem } from '@/api/cart'
import type { Address } from '@/api/user'
import {
  getCouponList,
  getUserCoupons,
  calculateDiscount,
  couponFaceText,
  couponThresholdText,
  type Coupon,
  type UserCoupon,
} from '@/api/promotion'
import { placeholderImage } from '@/utils/placeholder'
import { ensureLogin } from '@/utils/auth'

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

// 订单条目里的图只有 64px，用不带文字的小尺寸占位图
const thumbPlaceholderUri = placeholderImage(160, '')

const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const getItemImage = (item:CartItem) => item.productImage?resolveUrl(item.productImage):thumbPlaceholderUri
const displayAddress = (addr:any) => [addr.province,addr.city,addr.district,addr.detail].filter(Boolean).join('')||addr.receiver_address||''
const totalAmount = computed(() => orderItems.value.reduce((s,i)=>s+(i.price||0)*i.quantity,0))
const freightAmount = computed(() => totalAmount.value>=69?0:10)

// ===== 优惠券 =====
// 注意：传给下单接口的 coupon_id 是「用户券 ID」（user_coupon.id），
// 而试算接口用的是「券模板 ID」，两者别混。
const userCoupons = ref<UserCoupon[]>([])
const couponTemplates = ref<Coupon[]>([])
const selectedUserCouponId = ref(0)
const couponDiscount = ref(0)

const couponMap = computed(() => {
  const map = new Map<number, Coupon>()
  couponTemplates.value.forEach(c => map.set(Number(c.id), c))
  return map
})
const couponTemplateOf = (uc: UserCoupon) => couponMap.value.get(Number(uc.coupon_id))
/** 未使用、未过期且达到门槛的券才可选 */
const usableCoupons = computed(() =>
  userCoupons.value.filter(uc => {
    if (uc.status !== 0) return false
    if (uc.expire_at && new Date(uc.expire_at).getTime() < Date.now()) return false
    const tpl = couponTemplateOf(uc)
    return !tpl || Number(tpl.min_amount) <= totalAmount.value
  })
)
const couponLabel = (uc: UserCoupon) => {
  const tpl = couponTemplateOf(uc)
  const name = tpl?.name || '优惠券'
  return `${name}（${couponFaceText(tpl)} · ${couponThresholdText(tpl)}）`
}
const payableAmount = computed(() => Math.max(totalAmount.value + freightAmount.value - couponDiscount.value, 0))

const loadCoupons = async () => {
  try {
    const tplRes: any = await getCouponList({ status: 1, page: 1, page_size: 100 })
    couponTemplates.value = tplRes?.data || []
    if (userStore.token) {
      const mineRes: any = await getUserCoupons(Number(userStore.userId || 0), 0)
      userCoupons.value = mineRes?.data || []
    }
  } catch {
    // 券信息失败不影响下单流程
  }
}

const handleCouponChange = async () => {
  couponDiscount.value = 0
  if (!selectedUserCouponId.value) return
  const uc = userCoupons.value.find(c => c.id === selectedUserCouponId.value)
  const tpl = uc && couponTemplateOf(uc)
  if (!tpl) return
  try {
    const res: any = await calculateDiscount(Number(tpl.id), totalAmount.value)
    couponDiscount.value = Number(res?.data?.discount_amount || 0) || 0
  } catch {
    ElMessage.warning('优惠券试算失败，可稍后重试')
  }
}

const confirmAddress = () => {
  const addr = addresses.value.find(a=>a.id===selectedAddressId.value)
  if(addr){ selectedAddress.value=addr; showAddressDialog.value=false }
}

const handleSubmit = async () => {
  if(!selectedAddress.value){ElMessage.warning('请选择收货地址');return}
  if(orderItems.value.length===0){ElMessage.warning('订单商品不能为空');return}
  if(!ensureLogin(router, '请先登录后再提交订单'))return
  submitting.value=true
  try{
    const addr = selectedAddress.value
    const addrStr = displayAddress(addr)
    const r = await createOrder({
      user_id:userStore.userId,
      items:orderItems.value.map(i=>({sku_id:i.skuId,quantity:i.quantity,product_name:i.productName||'',price:String(i.price||0)})),
      address_id:addr.id,
      receiver_name:addr.receiver_name,
      receiver_phone:addr.receiver_phone,
      receiver_address:addrStr,
      clear_cart:false,
      order_type:1,
      // 用户券 ID（不是券模板 ID）；不选则不带，后端按 0 处理
      coupon_id:selectedUserCouponId.value > 0 ? selectedUserCouponId.value : undefined,
    })
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

onMounted(async () => {loading.value=true;try{await Promise.all([loadAddresses(),loadOrderItems(),loadCoupons()])}finally{loading.value=false}})
</script>

<style scoped>
.order-page { max-width:1100px; margin:0 auto; padding:32px 24px; min-height:calc(100vh-64px); font-family:var(--gf-font); }
.page-title { font-size:28px; font-weight:700; margin:0 0 32px; color:var(--text); letter-spacing:-.01em; }

.order-layout { display:flex; gap:24px; align-items:flex-start; }
.order-main { flex:1; min-width:0; }

/* Card */
.card { background:var(--gf-glass-2); -webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate)); backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate)); border:1px solid var(--gf-stroke); border-radius:var(--radius); box-shadow:var(--gf-inner-shadow); padding:24px; margin-bottom:20px; }
.card-head { display:flex; justify-content:space-between; align-items:center; margin-bottom:16px; }
.card-title { font-size:16px; font-weight:600; margin:0; color:var(--text); }

/* Address */
.address-info p { margin:4px 0; line-height:1.5; }
.addr-name { font-size:15px; font-weight:600; color:var(--text); margin:0 0 6px; }
.addr-phone { font-weight:400; color:var(--text-dim); margin-left:8px; }
.addr-detail { font-size:13px; color:var(--text-dim); margin:0; }
.btn-link { background:none;border:none;color:var(--accent);font-size:13px;font-weight:500;cursor:pointer;padding:0; }
.btn-link:hover { opacity:.8; }
.btn-outline { padding:10px 24px;border-radius:var(--radius-pill);border:1px solid rgba(79,216,255,.4);background:var(--gf-glass-1);color:var(--accent);font-size:13px;font-weight:500;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft);transition:all .2s; }
.btn-outline:hover { background:var(--accent-dim);border-color:rgba(79,216,255,.6); }

/* Order Items */
.order-items { display:flex; flex-direction:column; }
.item-row { display:flex; align-items:center; gap:14px; padding:14px 0; border-bottom:1px solid var(--gf-stroke); }
.item-row:last-child { border-bottom:none; }
.item-img { width:64px;height:64px;border-radius:var(--radius-sm);overflow:hidden;background:linear-gradient(150deg,rgba(255,255,255,.05),rgba(255,255,255,.01));flex-shrink:0; }
.item-img img { width:100%;height:100%;object-fit:cover; }
.item-info { flex:1;min-width:0; }
.item-name { font-size:14px;font-weight:500;color:var(--text);margin:0 0 4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.item-price { font-size:13px;color:var(--price);font-weight:600; }
.item-qty { font-size:14px;color:var(--text-dim);width:50px;text-align:center; }
.item-sub { font-size:15px;font-weight:600;color:var(--text);width:80px;text-align:right; }

/* Checkout Sidebar */
.order-sidebar { width:340px; flex-shrink:0; position:sticky; top:88px; }
.checkout-card { background:var(--gf-glass-2); -webkit-backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); border:1px solid var(--gf-stroke-strong); border-radius:var(--radius); box-shadow:var(--gf-shadow-3),var(--gf-inner-shadow); padding:24px; }
.checkout-row { display:flex; justify-content:space-between; padding:8px 0; font-size:14px; color:var(--text-dim); }
.checkout-row .discount-amount { color:#7CE38B; font-weight:600; }
.checkout-divider { height:1px; background:var(--gf-stroke); margin:12px 0; }
.checkout-total { display:flex; justify-content:space-between; align-items:center; padding:8px 0; }
.checkout-total span:first-child { font-size:16px; font-weight:600; color:var(--text); }
.total-price { font-size:28px; font-weight:800; color:var(--price); letter-spacing:-.02em; }
.btn-submit { width:100%; padding:16px; border-radius:var(--radius-pill); border:none; background:var(--gf-gradient); color:#04121a; font-size:16px; font-weight:700; cursor:pointer; box-shadow:var(--gf-inner-shadow-soft),0 12px 32px -14px var(--accent-glow); transition:filter .2s,box-shadow .2s,transform .2s; margin-top:20px; }
.btn-submit:hover:not(:disabled) { filter:brightness(1.06); box-shadow:var(--gf-inner-shadow-soft),0 20px 46px -16px var(--accent-glow); transform:translateY(-2px); }
.btn-submit:disabled { opacity:.35; cursor:not-allowed; }
.btn-submit.loading { opacity:.7; }
.freight-hint { font-size:12px; color:var(--text-dim); text-align:center; margin:10px 0 0; }

/* Empty */
.empty-block { text-align:center; padding:24px 0; color:var(--text-dim); font-size:14px; cursor:pointer; }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(4,6,12,.62); -webkit-backdrop-filter:blur(6px); backdrop-filter:blur(6px); z-index:200; display:flex; align-items:center; justify-content:center; }
.modal-card { background:var(--gf-glass-deep); -webkit-backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); border:1px solid var(--gf-stroke-strong); border-radius:var(--radius); box-shadow:var(--gf-shadow-3),var(--gf-inner-shadow); padding:28px; width:520px; max-height:80vh; overflow-y:auto; }
.modal-title { font-size:18px; font-weight:600; color:var(--text); margin:0 0 20px; }
.modal-footer { display:flex; justify-content:flex-end; gap:12px; margin-top:20px; }
.address-list { display:flex; flex-direction:column; gap:10px; }
.addr-option { display:flex; gap:12px; padding:14px; border:1px solid var(--gf-stroke); background:var(--gf-glass-1); border-radius:var(--radius-sm); box-shadow:var(--gf-inner-shadow-soft); cursor:pointer; transition:all .2s; }
.addr-option:hover { border-color:var(--gf-stroke-strong); background:var(--gf-glass-2); }
.addr-option.selected { border-color:rgba(79,216,255,.5); background:var(--accent-dim); }
.radio-dot { width:18px;height:18px;border-radius:50%;border:1px solid var(--gf-stroke-strong);flex-shrink:0;margin-top:2px;transition:all .15s; }
.radio-dot.checked { border-color:transparent;background:var(--gf-gradient);box-shadow:0 0 0 4px rgba(79,216,255,.15); }
.addr-content p { margin:0 0 4px; }
.addr-content .addr-name { font-size:14px;font-weight:600;color:var(--text); }
.addr-content .addr-phone { font-weight:400;color:var(--text-dim);margin-left:6px;font-size:13px; }
.addr-content .addr-detail { font-size:13px;color:var(--text-dim); }
.btn-cancel { padding:10px 24px;border-radius:var(--radius-pill);border:1px solid var(--gf-stroke);background:var(--gf-glass-1);color:var(--text);font-size:14px;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft);transition:all .2s; }
.btn-cancel:hover { background:var(--gf-glass-2);border-color:var(--gf-stroke-strong); }
.btn-save { padding:10px 28px;border-radius:var(--radius-pill);border:none;background:var(--gf-gradient);color:#04121a;font-size:14px;font-weight:700;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft),0 10px 26px -12px var(--accent-glow); }
.btn-save:hover { filter:brightness(1.06);box-shadow:var(--gf-inner-shadow-soft),0 16px 34px -14px var(--accent-glow); }
.addr-edit { flex-shrink:0; }
.address-form { display:flex; flex-direction:column; gap:12px; }
.form-row { display:flex; align-items:center; gap:10px; }
.form-row label { width:70px; font-size:13px; color:var(--text-dim); flex-shrink:0; }
.form-row input { flex:1; padding:9px 12px; border-radius:var(--radius-sm); background:var(--gf-glass-1); border:1px solid var(--gf-stroke); color:var(--text); font-size:13px; outline:none; box-shadow:var(--gf-inner-shadow-soft); transition:border-color .2s,box-shadow .2s; font-family:inherit; }
.form-row input:focus { border-color:rgba(79,216,255,.5); box-shadow:var(--gf-inner-shadow-soft),0 0 0 3px var(--accent-dim); }
.default-check { display:flex; align-items:center; gap:6px; font-size:13px; color:var(--text-dim); cursor:pointer; }
.footer-right { display:flex; gap:12px; }

@media(max-width:768px) { .order-layout { flex-direction:column; } .order-sidebar { width:100%;position:static; } }
</style>
