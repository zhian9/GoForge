<template>
  <div class="seckill-detail" v-loading="loading">
    <div v-if="!activity && !loading" class="empty">活动不存在或已下线</div>

    <div v-else-if="activity" class="wrap">
      <div class="left">
        <div class="img">
          <img :src="imgUrl" :alt="activity.sku_name || activity.name" @error="handleImgError" />
        </div>
      </div>

      <div class="right">
        <h2 class="name">{{ activity.name }}</h2>
        <div class="sku">{{ activity.sku_name }}</div>

        <div class="price-box">
          <div class="price">
            <span class="label">秒杀价</span>
            <span class="val">¥{{ formatMoney(activity.seckill_price) }}</span>
          </div>
          <div class="price origin">
            <span class="label">原价</span>
            <span class="val">¥{{ formatMoney(activity.original_price) }}</span>
          </div>
        </div>

        <div class="stats">
          <div class="stat">
            <div class="k">状态</div>
            <div class="v" :class="statusClass">{{ statusText }}</div>
          </div>
          <div class="stat">
            <div class="k">剩余库存</div>
            <div class="v">{{ remainStock }}</div>
          </div>
          <div class="stat">
            <div class="k">倒计时</div>
            <div class="v">{{ countdownText }}</div>
          </div>
        </div>

        <div class="action">
          <el-input-number v-model="quantity" :min="1" :max="1" :step="1" disabled />
          <el-button
            type="danger"
            :loading="submitLoading"
            :disabled="buttonDisabled"
            @click="handleSeckill"
          >
            {{ buttonText }}
          </el-button>
        </div>

        <div class="hint">
          秒杀成功后订单将异步创建，可能需要几秒钟后在「我的订单」中看到。
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getSeckillActivity, seckill, type SeckillActivity } from '@/api/seckill'
import { getPublicUrl } from '@/utils/image'
import { placeholderImage } from '@/utils/placeholder'
import { useUserStore } from '@/stores/user'
import { ensureLogin } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const submitLoading = ref(false)
const activity = ref<SeckillActivity | null>(null)
const quantity = ref(1)

const tick = ref(0)
let timer: number | undefined

const nowSec = () => Math.floor(Date.now() / 1000)

const startSec = computed(() => Number(activity.value?.start_time || 0) || 0)
const endSec = computed(() => Number(activity.value?.end_time || 0) || 0)

const statusText = computed(() => {
  const st = Number(activity.value?.status ?? -1)
  if (st === 0) return '未开始'
  if (st === 1) return '进行中'
  if (st === 2) return '已结束'
  const now = nowSec()
  if (startSec.value && now < startSec.value) return '未开始'
  if (endSec.value && now > endSec.value) return '已结束'
  return '进行中'
})

const statusClass = computed(() => {
  if (statusText.value === '进行中') return 'on'
  if (statusText.value === '未开始') return 'pending'
  return 'off'
})

const remainStock = computed(() => {
  const stock = Number(activity.value?.stock || 0)
  const sold = Number(activity.value?.sold || 0)
  return Math.max(stock - sold, 0)
})

const imgUrl = computed(() => getPublicUrl(activity.value?.sku_image || ''))

const formatMoney = (v?: string) => {
  const n = Number(v || 0)
  return n.toFixed(2)
}

const countdownText = computed(() => {
  // 触发更新
  void tick.value
  const now = nowSec()
  if (!activity.value) return '-'
  if (statusText.value === '未开始' && startSec.value) {
    return formatRemain(Math.max(startSec.value - now, 0)) + ' 后开始'
  }
  if (statusText.value === '进行中' && endSec.value) {
    return formatRemain(Math.max(endSec.value - now, 0)) + ' 后结束'
  }
  return '-'
})

const formatRemain = (sec: number) => {
  const s = Math.max(sec, 0)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = s % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(ss)}`
}

const buttonDisabled = computed(() => {
  if (submitLoading.value) return true
  if (!activity.value) return true
  if (remainStock.value <= 0) return true
  return statusText.value !== '进行中'
})

const buttonText = computed(() => {
  if (!activity.value) return '不可用'
  if (remainStock.value <= 0) return '已抢光'
  if (statusText.value === '未开始') return '未开始'
  if (statusText.value === '已结束') return '已结束'
  return '立即秒杀'
})

const handleImgError = (e: Event) => {
  const img = e.target as HTMLImageElement
  // 原先指向 /placeholder.png，但 public/ 下并没有这个文件，失败会二次 404
  img.src = placeholderImage(400, '暂无图片')
}

const fetchDetail = async () => {
  const id = Number(route.params.id)
  if (!Number.isFinite(id) || id <= 0) {
    activity.value = null
    return
  }
  loading.value = true
  try {
    const resp = await getSeckillActivity(id)
    activity.value = resp.data || null
  } finally {
    loading.value = false
  }
}

const handleSeckill = async () => {
  if (!activity.value) return
  if (!ensureLogin(router, '请先登录再参与秒杀')) return
  if (!userStore.userId) { ElMessage.warning('账号信息异常，请重新登录'); return }
  submitLoading.value = true
  try {
    const resp = await seckill({
      user_id: userStore.userId,
      sku_id: Number(activity.value.sku_id),
      quantity: 1,
    })
    const msg = resp.data?.message || resp.message || '请求已受理'
    if (resp.data?.success) {
      ElMessage.success(resp.data?.order_no ? `抢购成功，订单号：${resp.data.order_no}` : msg)
      router.push('/orders')
    } else {
      ElMessage.warning(msg)
    }
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  await fetchDetail()
  timer = window.setInterval(() => {
    tick.value++
  }, 1000)
})

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<style scoped>
.seckill-detail {
  padding: 20px;
}

.empty {
  padding: 60px 0;
  text-align: center;
  color: var(--text-dim);
}

.wrap {
  display: grid;
  grid-template-columns: 420px 1fr;
  gap: 24px;
}

.img {
  width: 100%;
  aspect-ratio: 1 / 1;
  background: linear-gradient(150deg, rgba(255, 255, 255, 0.055), rgba(255, 255, 255, 0.012));
  border: 1px solid var(--gf-stroke);
  border-radius: var(--radius);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
  overflow: hidden;
}

.img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.name {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
}

.sku {
  margin-top: 6px;
  color: var(--text-dim);
  font-size: 14px;
}

.price-box {
  margin-top: 18px;
  display: flex;
  gap: 20px;
  align-items: baseline;
}

.price .label {
  font-size: 12px;
  color: var(--text-dim);
  margin-right: 8px;
}

.price .val {
  font-size: 26px;
  font-weight: 800;
  color: var(--price);
  letter-spacing: -0.01em;
}

.price.origin .val {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-dim);
  text-decoration: line-through;
}

.stats {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.stat {
  border: 1px solid var(--gf-stroke);
  border-radius: var(--radius-sm);
  padding: 12px;
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
}

.stat .k {
  font-size: 12px;
  color: var(--text-dim);
}

.stat .v {
  margin-top: 6px;
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
}

.stat .v.on {
  color: var(--danger);
}

.stat .v.pending {
  color: var(--warning);
}

.stat .v.off {
  color: var(--text-dim);
}

.action {
  margin-top: 18px;
  display: flex;
  gap: 12px;
  align-items: center;
}

.hint {
  margin-top: 12px;
  font-size: 12px;
  color: var(--text-dim);
}

@media (max-width: 1000px) {
  .wrap {
    grid-template-columns: 1fr;
  }
}
</style>


