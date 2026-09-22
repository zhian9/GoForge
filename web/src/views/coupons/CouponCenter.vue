<template>
  <div class="coupon-page">
    <div class="page-head">
      <h1>优惠券</h1>
      <p class="sub">先领券，下单时自动可选，抵扣后的金额才是实付金额</p>
    </div>

    <div class="tabs">
      <button class="tab" :class="{ active: tab === 'center' }" @click="switchTab('center')">领券中心</button>
      <button class="tab" :class="{ active: tab === 'mine' }" @click="switchTab('mine')">我的优惠券</button>
    </div>

    <!-- ===== 领券中心 ===== -->
    <div v-if="tab === 'center'" class="list">
      <el-skeleton v-if="loading" :rows="4" animated />
      <el-empty v-else-if="claimable.length === 0" description="暂无可领取的优惠券" />
      <div v-for="c in claimable" v-else :key="c.id" class="coupon-card">
        <div class="face">
          <span class="face-value">{{ couponFaceText(c) }}</span>
          <span class="face-threshold">{{ couponThresholdText(c) }}</span>
        </div>
        <div class="meta">
          <h3>{{ c.name }}</h3>
          <p class="dim">有效期至 {{ formatDate(c.valid_end_time) }}</p>
          <p class="dim">{{ stockText(c) }}</p>
        </div>
        <div class="action">
          <el-button type="primary" :loading="receiving === c.id" @click="handleReceive(c)">立即领取</el-button>
        </div>
      </div>
    </div>

    <!-- ===== 我的优惠券 ===== -->
    <div v-else class="list">
      <div class="status-tabs">
        <button
          v-for="s in statusTabs"
          :key="s.value"
          class="status-tab"
          :class="{ active: status === s.value }"
          @click="changeStatus(s.value)"
        >{{ s.label }}</button>
      </div>

      <el-skeleton v-if="loading" :rows="4" animated />
      <el-empty v-else-if="myCoupons.length === 0" description="这里还没有优惠券" />
      <div v-for="uc in myCoupons" v-else :key="uc.id" class="coupon-card" :class="{ disabled: uc.status !== 0 }">
        <div class="face">
          <span class="face-value">{{ couponFaceText(templateOf(uc)) }}</span>
          <span class="face-threshold">{{ couponThresholdText(templateOf(uc)) }}</span>
        </div>
        <div class="meta">
          <h3>{{ templateOf(uc)?.name || '优惠券' }}</h3>
          <p class="dim">有效期至 {{ formatDate(uc.expire_at) }}</p>
          <p class="dim">{{ statusText(uc.status) }}</p>
        </div>
        <div class="action">
          <el-button v-if="uc.status === 0" @click="goUse">去使用</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { ensureLogin } from '@/utils/auth'
import {
  getCouponList,
  getUserCoupons,
  receiveCoupon,
  couponFaceText,
  couponThresholdText,
  type Coupon,
  type UserCoupon,
} from '@/api/promotion'

const router = useRouter()
const userStore = useUserStore()

const tab = ref<'center' | 'mine'>('center')
const loading = ref(false)
const receiving = ref(0)

const coupons = ref<Coupon[]>([])
const myCoupons = ref<UserCoupon[]>([])
const status = ref(-1)

const statusTabs = [
  { label: '全部', value: -1 },
  { label: '未使用', value: 0 },
  { label: '已使用', value: 1 },
  { label: '已过期', value: 2 },
]

/** 可领取：启用中、在有效期内、还有剩余名额 */
const claimable = computed(() => {
  const now = Date.now()
  return coupons.value.filter((c) => {
    if (c.status !== 1) return false
    if (new Date(c.valid_start_time).getTime() > now) return false
    if (new Date(c.valid_end_time).getTime() < now) return false
    // 注意：proto 的 Coupon 里没有 totalCount/usedCount，网关不返回这两个字段（值为 0），
    // 所以只有明确配置了限量（> 0）时才判断是否领完，否则会把所有券都过滤掉。
    const total = Number(c.total_count || 0)
    if (total > 0 && Number(c.used_count || 0) >= total) return false
    return true
  })
})

const couponById = computed(() => {
  const map = new Map<number, Coupon>()
  coupons.value.forEach((c) => map.set(Number(c.id), c))
  return map
})

const templateOf = (uc: UserCoupon): Coupon | undefined => couponById.value.get(Number(uc.coupon_id))

const formatDate = (value?: string) => (value ? new Date(value).toLocaleDateString('zh-CN') : '-')

const stockText = (c: Coupon) => {
  const total = Number(c.total_count || 0)
  if (total <= 0) return ''
  const left = Math.max(total - Number(c.used_count), 0)
  return `剩余 ${left} 张`
}

const statusText = (s: number) => (s === 0 ? '未使用' : s === 1 ? '已使用' : '已过期')

const loadCoupons = async () => {
  const res: any = await getCouponList({ status: 1, page: 1, page_size: 100 })
  coupons.value = res?.data || []
}

const loadMine = async () => {
  if (!userStore.token) {
    myCoupons.value = []
    return
  }
  const res: any = await getUserCoupons(Number(userStore.userId || 0), status.value)
  myCoupons.value = res?.data || []
}

const loadAll = async () => {
  loading.value = true
  try {
    await Promise.all([loadCoupons(), tab.value === 'mine' ? loadMine() : Promise.resolve()])
  } catch {
    // 错误提示由 request 拦截器统一处理
  } finally {
    loading.value = false
  }
}

const switchTab = async (next: 'center' | 'mine') => {
  tab.value = next
  if (next === 'mine') {
    if (!ensureLogin(router, '登录后查看我的优惠券')) return
    loading.value = true
    try {
      await loadMine()
    } finally {
      loading.value = false
    }
  }
}

const changeStatus = async (value: number) => {
  status.value = value
  loading.value = true
  try {
    await loadMine()
  } finally {
    loading.value = false
  }
}

const handleReceive = async (c: Coupon) => {
  if (!ensureLogin(router, '登录后即可领取优惠券')) return
  receiving.value = Number(c.id)
  try {
    const res: any = await receiveCoupon(Number(c.id))
    if (res?.code === 0) {
      ElMessage.success('领取成功，可在「我的优惠券」查看')
      await loadCoupons()
    } else {
      ElMessage.warning(res?.message || '领取失败')
    }
  } catch (e: any) {
    ElMessage.warning(e?.message || '领取失败')
  } finally {
    receiving.value = 0
  }
}

const goUse = () => router.push('/products')

onMounted(loadAll)
</script>

<style scoped>
.coupon-page { max-width: 1000px; margin: 0 auto; padding: 28px 20px 60px; }
.page-head h1 { margin: 0 0 6px; font-size: 26px; color: var(--text); }
.page-head .sub { margin: 0 0 20px; color: var(--text-dim); font-size: 13px; }
.tabs { display: flex; gap: 8px; margin-bottom: 18px; }
.tab {
  padding: 8px 18px; border-radius: 999px; cursor: pointer;
  border: 1px solid var(--gf-border, rgba(255,255,255,.12));
  background: rgba(255,255,255,.04); color: var(--text-dim); font-size: 14px;
}
.tab.active { color: #0b1220; background: var(--gf-accent, #4FD8FF); border-color: transparent; font-weight: 600; }
.status-tabs { display: flex; gap: 6px; margin-bottom: 14px; flex-wrap: wrap; }
.status-tab {
  padding: 5px 14px; border-radius: 999px; cursor: pointer; font-size: 13px;
  border: 1px solid var(--gf-border, rgba(255,255,255,.12));
  background: transparent; color: var(--text-dim);
}
.status-tab.active { color: var(--gf-accent, #4FD8FF); border-color: var(--gf-accent, #4FD8FF); }
.list { display: flex; flex-direction: column; gap: 14px; }
.coupon-card {
  display: flex; align-items: center; gap: 20px; padding: 18px 20px;
  border-radius: 16px; border: 1px solid var(--gf-border, rgba(255,255,255,.1));
  background: var(--gf-glass, rgba(16,22,34,.72));
}
.coupon-card.disabled { opacity: .55; }
.face { min-width: 132px; display: flex; flex-direction: column; align-items: center; color: var(--gf-accent, #4FD8FF); }
.face-value { font-size: 24px; font-weight: 700; }
.face-threshold { font-size: 12px; color: var(--text-dim); margin-top: 4px; }
.meta { flex: 1; }
.meta h3 { margin: 0 0 6px; font-size: 16px; color: var(--text); }
.dim { margin: 2px 0; font-size: 12px; color: var(--text-dim); }
.action { min-width: 104px; text-align: right; }
</style>
