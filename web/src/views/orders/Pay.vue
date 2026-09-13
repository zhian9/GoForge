<template>
  <div class="pay-page">
    <h1 class="page-title">收银台</h1>

    <div class="pay-card" v-if="!paid && !failed">
      <div class="amount-block">
        <span class="amount-label">支付金额</span>
        <span class="amount-value">¥{{ fmt(amount) }}</span>
        <span class="balance-tip">账户余额 ¥{{ fmt((userStore.userInfo as any)?.balance ?? 0) }}</span>
      </div>

      <div class="method-block">
        <span class="block-label">支付方式</span>
        <div class="method-list">
          <button
            v-for="m in methods"
            :key="m.value"
            class="method-btn"
            :class="{ active: selectedMethod === m.value }"
            @click="switchMethod(m.value)"
          >{{ m.label }}</button>
        </div>
      </div>

      <div class="mock-tip">当前为 Mock 支付渠道，用于联调跑通流程，不会真实扣款。</div>

      <div class="pay-actions">
        <button class="btn-pay" @click="doPay(1)">{{ selectedMethod === 4 ? '确认支付' : '模拟支付成功' }}</button>
        <button v-if="selectedMethod !== 4" class="btn-fail" @click="doPay(2)">模拟支付失败</button>
      </div>
    </div>

    <div class="pay-card result" v-else-if="paid">
      <div class="result-icon ok">✓</div>
      <p class="result-title">支付成功</p>
      <p class="result-sub">订单号 {{ orderNo }}</p>
      <button class="btn-primary" @click="goOrder">查看订单</button>
    </div>

    <div class="pay-card result" v-else>
      <div class="result-icon fail">✕</div>
      <p class="result-title">支付失败</p>
      <p class="result-sub">订单号 {{ orderNo }}</p>
      <button class="btn-primary" @click="goOrder">返回订单</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getOrderDetail } from '@/api/order'
import { createPayment, mockPayCallback, queryPaymentStatus } from '@/api/payment'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const orderId = Number(route.params.id || 0)
const orderNo = ref('')
const amount = ref(0)
const paymentNo = ref('')
const selectedMethod = ref(1)
const paid = ref(false)
const failed = ref(false)
const creating = ref(false)

const methods = [
  { value: 1, label: '微信支付' },
  { value: 2, label: '支付宝' },
  { value: 3, label: '银联' },
  { value: 4, label: '余额支付' },
]

const fmt = (v: any) => { const n = Number(v || 0); return isNaN(n) ? '0.00' : n.toFixed(2) }

const loadOrder = async () => {
  try {
    const r = await getOrderDetail(orderId)
    if (r.code === 0 && r.data) {
      const d = r.data as any
      orderNo.value = d.order_no || d.orderNo || ''
      const amt = d.total_amount ?? d.totalAmount ?? d.pay_amount ?? 0
      amount.value = typeof amt === 'string' ? parseFloat(amt) : Number(amt || 0)
    }
  } catch {}
}

const createPay = async () => {
  if (!orderId || creating.value) return
  creating.value = true
  try {
    const r = await createPayment({
      order_id: orderId,
      order_no: orderNo.value,
      amount: amount.value.toFixed(2),
      payment_method: selectedMethod.value,
    })
    if (r.code === 0 && r.data) {
      paymentNo.value = r.data.payment_no || r.data.paymentNo || ''
    }
  } catch {}
  finally { creating.value = false }
}

const switchMethod = async (m: number) => {
  if (m === selectedMethod.value) return
  selectedMethod.value = m
  await createPay()
}

const doPay = async (status: number) => {
  if (!paymentNo.value) { ElMessage.warning('支付单未创建'); return }
  try {
    await mockPayCallback({
      payment_no: paymentNo.value,
      status,
      third_party_no: `MOCK-${Date.now()}`,
    })
    if (status === 1) {
      // 轮询确认支付状态已落库
      let tries = 0
      while (tries < 5) {
        const s = await queryPaymentStatus(paymentNo.value)
        if (s.status === 1) break
        tries++
        await new Promise(r => setTimeout(r, 500))
      }
      paid.value = true
    } else {
      failed.value = true
    }
  } catch (e: any) {
    ElMessage.error(e.message || '支付回调失败')
  }
}

const goOrder = () => router.push(`/orders/${orderId}`)

onMounted(async () => {
  await loadOrder()
  await createPay()
})
</script>

<style scoped>
.pay-page { --accent:#00F5FF; --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.02); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); min-height:calc(100vh-64px); padding:32px 24px; font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.page-title { max-width:480px; margin:0 auto 24px; font-size:26px; font-weight:700; color:var(--text); }
.pay-card { max-width:480px; margin:0 auto; background:var(--card-bg); border:1px solid var(--border); border-radius:18px; padding:32px; }
.amount-block { text-align:center; padding-bottom:24px; border-bottom:1px solid var(--border); }
.amount-label { display:block; font-size:13px; color:var(--text-dim); margin-bottom:8px; }
.amount-value { font-size:40px; font-weight:800; color:var(--accent); letter-spacing:-.02em; }
.balance-tip { display:block; margin-top:8px; font-size:12px; color:var(--text-dim); }
.method-block { padding:24px 0; }
.block-label { font-size:13px; color:var(--text-dim); display:block; margin-bottom:12px; }
.method-list { display:flex; gap:10px; }
.method-btn { flex:1; padding:12px; border-radius:10px; border:1px solid var(--border); background:transparent; color:var(--text); font-size:14px; cursor:pointer; transition:all .15s; }
.method-btn.active { border-color:var(--accent); background:rgba(0,245,255,.08); color:var(--accent); }
.mock-tip { font-size:12px; color:var(--text-dim); background:rgba(245,158,11,.08); border:1px solid rgba(245,158,11,.2); border-radius:8px; padding:10px 14px; margin-bottom:24px; }
.pay-actions { display:flex; gap:12px; }
.btn-pay { flex:1; padding:15px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:15px; font-weight:700; cursor:pointer; }
.btn-fail { flex:1; padding:15px; border-radius:100px; border:1px solid rgba(248,113,113,.4); background:transparent; color:#F87171; font-size:15px; font-weight:600; cursor:pointer; }
.result { text-align:center; }
.result-icon { width:64px; height:64px; border-radius:50%; display:inline-flex; align-items:center; justify-content:center; font-size:30px; margin-bottom:16px; }
.result-icon.ok { background:rgba(16,185,129,.15); color:#10B981; }
.result-icon.fail { background:rgba(248,113,113,.15); color:#F87171; }
.result-title { font-size:22px; font-weight:700; color:var(--text); margin:0 0 8px; }
.result-sub { font-size:14px; color:var(--text-dim); margin:0 0 24px; }
.btn-primary { padding:14px 40px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:15px; font-weight:700; cursor:pointer; }
</style>
