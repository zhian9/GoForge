<template>
  <div class="dashboard">
    <PageHeader title="仪表盘" desc="GoForge 平台核心数据概览">
      <template #actions>
        <el-button @click="loadAll" :loading="loading">
          <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-3-6.7"/><polyline points="21 4 21 10 15 10"/></svg>
          刷新数据
        </el-button>
        <el-button type="primary" @click="router.push('/products')">
          <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"/></svg>
          新增商品
        </el-button>
      </template>
    </PageHeader>

    <!-- ==================== KPI ==================== -->
    <section class="stat-grid">
      <StatCard
        v-for="card in kpiCards"
        :key="card.label"
        :icon="card.icon"
        :value="card.value"
        :label="card.label"
        :hint="card.hint"
        :hint-tone="card.hintTone"
        :color="card.color"
      />
    </section>

    <!-- ==================== 最近订单 + 快捷操作 ==================== -->
    <section class="grid-main">
      <DarkCard title="最近订单" desc="按下单时间倒序展示最新 6 笔">
        <template #header>
          <router-link class="card-link" to="/orders">查看全部 →</router-link>
        </template>

        <el-table v-loading="loading" :data="recentOrders" size="small" empty-text="暂无订单数据">
          <el-table-column label="订单号" min-width="170">
            <template #default="{ row }">
              <span class="mono">{{ row.orderNo || row.order_no || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="用户" width="90">
            <template #default="{ row }">{{ row.userId ?? row.user_id ?? '-' }}</template>
          </el-table-column>
          <el-table-column label="金额" width="110" align="right">
            <template #default="{ row }">
              <span class="amount">¥{{ Number(row.totalAmount ?? row.total_amount ?? 0).toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }">
              <span class="status-tag" :class="`s-${row.status ?? 0}`">{{ statusText(row.status ?? 0) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="下单时间" width="150">
            <template #default="{ row }">
              <span class="muted">{{ formatTime(row.createdAt || row.created_at) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </DarkCard>

      <DarkCard title="快捷操作" desc="常用入口">
        <div class="quick-actions">
          <button v-for="a in quickActions" :key="a.label" class="qa-btn" type="button" @click="router.push(a.path)">
            <span class="qa-icon" :class="a.tint">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="a.icon"></svg>
            </span>
            <span class="qa-label">{{ a.label }}</span>
            <span class="qa-desc">{{ a.desc }}</span>
          </button>
        </div>
      </DarkCard>
    </section>

    <!-- ==================== 状态分布 / 库存预警 / 最新用户 ==================== -->
    <section class="grid-thirds">
      <DarkCard title="订单状态分布" desc="按全部订单统计">
        <div v-if="hasStatusData" class="status-list">
          <div v-for="s in statusRows" :key="s.label" class="status-row">
            <span class="status-name">{{ s.label }}</span>
            <div class="status-track">
              <div class="status-fill" :class="s.tone" :style="{ width: s.percent + '%' }"></div>
            </div>
            <span class="status-count">{{ s.count }}</span>
          </div>
        </div>
        <div v-else class="mini-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="9"/><path d="M12 8v5M12 16h.01"/></svg>
          订单列表接口暂未返回数据
        </div>
      </DarkCard>

      <DarkCard title="库存预警" :desc="lowStock.length ? '库存低于 20 件，建议补货' : '库存水位正常'">
        <ul v-if="lowStock.length" class="mini-list">
          <li v-for="p in lowStock" :key="p.id" class="mini-row">
            <span class="mini-name" :title="p.name">{{ p.name }}</span>
            <span class="stock-chip" :class="Number(p.stock) <= 5 ? 'danger' : 'warn'">剩 {{ p.stock }}</span>
          </li>
        </ul>
        <div v-else class="mini-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="m9 12 2 2 4-4"/><circle cx="12" cy="12" r="9"/></svg>
          暂无低库存商品
        </div>
      </DarkCard>

      <DarkCard title="最新用户" desc="最近注册的 5 位用户">
        <template #header>
          <router-link class="card-link" to="/users">全部用户 →</router-link>
        </template>

        <ul v-if="latestUsers.length" class="mini-list">
          <li v-for="u in latestUsers" :key="u.id" class="mini-row">
            <span class="mini-avatar">{{ (u.nickname || u.username || 'U')[0] }}</span>
            <span class="mini-body">
              <span class="mini-name">{{ u.nickname || u.username }}</span>
              <span class="mini-sub">{{ formatDate(u.createdAt || u.created_at) }}</span>
            </span>
            <span v-if="Number(u.isAdmin ?? u.is_admin ?? 0) === 1" class="admin-chip">管理员</span>
          </li>
        </ul>
        <div v-else class="mini-empty">暂无用户数据</div>
      </DarkCard>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'
import StatCard from '@/components/StatCard.vue'
import { getUserList } from '@/api/user'
import { getProductList } from '@/api/product'
import { getOrderList, getOrderStats } from '@/api/order'

const router = useRouter()
const loading = ref(true)

const users = ref<any[]>([])
const products = ref<any[]>([])
const recentOrders = ref<any[]>([])

const userTotal = ref(0)
const productTotal = ref(0)
const orderTotal = ref(0)
const totalSales = ref(0)
const todayOrders = ref(0)
const statusCounts = ref<Record<number, number>>({})

const DANGER_STOCK = 20

const fmt = (v: any) => {
  const n = Number(v || 0)
  return Number.isNaN(n) ? '0' : n.toLocaleString()
}

const formatTime = (raw?: string) => (raw ? raw.replace('T', ' ').slice(0, 16) : '-')
const formatDate = (raw?: string) => (raw ? raw.replace('T', ' ').slice(0, 10) : '-')

const adminCount = computed(() => users.value.filter((u) => Number(u.isAdmin ?? u.is_admin ?? 0) === 1).length)
const lowStock = computed(() =>
  products.value
    .filter((p) => Number(p.stock) < DANGER_STOCK)
    .sort((a, b) => Number(a.stock) - Number(b.stock))
    .slice(0, 5)
)
const latestUsers = computed(() =>
  [...users.value]
    .sort((a, b) => String(b.createdAt || b.created_at || '').localeCompare(String(a.createdAt || a.created_at || '')))
    .slice(0, 5)
)
const avgOrderValue = computed(() => (orderTotal.value > 0 ? totalSales.value / orderTotal.value : 0))

const kpiCards = computed(() => [
  {
    label: '总用户数',
    value: fmt(userTotal.value),
    hint: `其中管理员 ${adminCount.value} 位`,
    hintTone: 'default' as const,
    color: 'cyan',
    icon: '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 4-7 8-7s8 3 8 7"/>',
  },
  {
    label: '商品总数',
    value: fmt(productTotal.value),
    hint: lowStock.value.length ? `低库存 ${lowStock.value.length} 件待补货` : '库存水位正常',
    hintTone: lowStock.value.length ? ('warn' as const) : ('ok' as const),
    color: 'blue',
    icon: '<path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4zM3 6h18"/><path d="M16 10a4 4 0 01-8 0"/>',
  },
  {
    label: '订单总数',
    value: fmt(orderTotal.value),
    hint: todayOrders.value > 0 ? `今日新增 ${todayOrders.value} 单` : '今日暂无新订单',
    hintTone: todayOrders.value > 0 ? ('ok' as const) : ('default' as const),
    color: 'green',
    icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/>',
  },
  {
    label: '总销售额',
    value: '¥' + totalSales.value.toFixed(2),
    hint: `客单价 ¥${avgOrderValue.value.toFixed(2)}`,
    hintTone: 'info' as const,
    color: 'red',
    icon: '<rect x="2" y="4" width="20" height="16" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/>',
  },
])

const statusText = (s: number) => ({ 0: '已取消', 1: '待支付', 2: '待发货', 3: '待收货', 4: '已完成', 5: '已退款' } as Record<number, string>)[s] || '未知'

const STATUS_LIST = [
  { key: 1, label: '待支付', tone: 'pending' },
  { key: 2, label: '待发货', tone: 'active' },
  { key: 3, label: '待收货', tone: 'progress' },
  { key: 4, label: '已完成', tone: 'done' },
  { key: 0, label: '已取消', tone: 'cancel' },
]

const statusRows = computed(() => {
  const max = Math.max(1, ...STATUS_LIST.map((s) => statusCounts.value[s.key] || 0))
  return STATUS_LIST.map((s) => {
    const count = statusCounts.value[s.key] || 0
    return { ...s, count, percent: Math.round((count / max) * 100) }
  })
})
const hasStatusData = computed(() => STATUS_LIST.some((s) => (statusCounts.value[s.key] || 0) > 0))

const quickActions = [
  { label: '添加商品', desc: '上架新 SPU', path: '/products', tint: 'tint-cyan', icon: '<path d="M12 5v14M5 12h14"/>' },
  { label: '新建秒杀', desc: '配置活动场次', path: '/seckill-activities', tint: 'tint-rose', icon: '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>' },
  { label: '订单处理', desc: '发货 / 取消', path: '/orders', tint: 'tint-violet', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/>' },
  { label: '用户管理', desc: '账号与权限', path: '/users', tint: 'tint-mint', icon: '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 4-7 8-7s8 3 8 7"/>' },
]

/**
 * 仪表盘数据。
 *
 * 说明：user / product 两个列表接口的返回是 camelCase（网关行为），而 api 层的类型声明
 * 写的是 snake_case，所以这里取值统一写成 `camel ?? snake`，两种命名都兼容。
 */
const loadAll = async () => {
  loading.value = true
  try {
    await Promise.all([
      (async () => {
        try {
          const r: any = await getUserList({ page: 1, page_size: 100 })
          if (r.code === 0) {
            users.value = r.data?.list || r.data?.users || []
            userTotal.value = Number(r.data?.total || users.value.length || 0)
          }
        } catch { /* 单项失败不影响其余卡片 */ }
      })(),
      (async () => {
        try {
          const r: any = await getProductList({ page: 1, page_size: 100 })
          if (r.code === 0) {
            products.value = r.data?.list || []
            productTotal.value = Number(r.data?.total || products.value.length || 0)
          }
        } catch { /* ignore */ }
      })(),
      (async () => {
        try {
          const r: any = await getOrderStats()
          if (r.code === 0) {
            orderTotal.value = Number(r.totalOrders ?? 0)
            totalSales.value = Number(r.totalSales ?? 0)
            todayOrders.value = Number(r.todayOrders ?? 0)
          }
        } catch { /* ignore */ }
      })(),
      (async () => {
        try {
          const r: any = await getOrderList({ page: 1, page_size: 6, status: -1 })
          if (r.code === 0) recentOrders.value = r.data?.orders || r.data?.list || []
        } catch { /* ignore */ }
      })(),
      // 状态分布：按 status 逐个取 total，比拉全量再前端统计更准（订单量大时也不会失真）
      (async () => {
        const entries = await Promise.all(
          STATUS_LIST.map(async (s) => {
            try {
              const r: any = await getOrderList({ page: 1, page_size: 1, status: s.key })
              return [s.key, r.code === 0 ? Number(r.data?.total || 0) : 0] as const
            } catch {
              return [s.key, 0] as const
            }
          })
        )
        statusCounts.value = Object.fromEntries(entries)
      })(),
    ])
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadAll()
})
</script>

<style scoped>
.dashboard { display: flex; flex-direction: column; gap: 18px; }

.btn-icon { width: 14px; height: 14px; }

/* ==================== KPI ==================== */
.stat-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; }

/* ==================== 主区（2:1） ==================== */
.grid-main { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 16px; align-items: start; }

/* ==================== 三等分 ==================== */
.grid-thirds { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; align-items: start; }

.card-link {
  font-size: 12px;
  color: var(--gf-accent);
  text-decoration: none;
  white-space: nowrap;
  transition: opacity .2s ease;
}
.card-link:hover { opacity: .75; }

.mono { font-family: var(--gf-font-num); font-size: 12px; color: var(--gf-text); }
.amount { color: var(--gf-gold); font-weight: 600; }
.muted { color: var(--gf-text-mute); font-size: 12px; }

/* ==================== 状态标签（订单状态全站统一） ==================== */
.status-tag { display: inline-block; padding: 2px 10px; border-radius: var(--radius-pill); font-size: 11px; font-weight: 600; white-space: nowrap; }
.s-0, .s-5 { background: rgba(255, 107, 129, .13); color: var(--gf-danger); }
.s-1 { background: rgba(251, 191, 107, .13); color: var(--gf-warning); }
.s-2 { background: rgba(124, 192, 255, .13); color: var(--gf-info); }
.s-3 { background: var(--accent-dim); color: var(--gf-accent); }
.s-4 { background: rgba(74, 222, 155, .13); color: var(--gf-success); }

/* ==================== 快捷操作 ==================== */
.quick-actions { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }

.qa-btn {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--gf-text-dim);
  font-family: var(--gf-font);
  text-align: left;
  cursor: pointer;
  transition: transform .25s cubic-bezier(.22, 1, .36, 1), border-color .2s ease, background .2s ease, box-shadow .25s ease;
}

.qa-btn:hover {
  transform: translateY(-3px);
  border-color: var(--gf-stroke-strong);
  background: var(--gf-glass-2);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow-soft);
}

.qa-icon { width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: var(--radius-xs); box-shadow: var(--gf-inner-shadow-soft); }
.qa-icon svg { width: 16px; height: 16px; }
.qa-label { font-size: 13px; font-weight: 600; color: var(--gf-text); }
.qa-desc { font-size: 11px; color: var(--gf-text-mute); }

.tint-cyan { background: rgba(79, 216, 255, .13); color: var(--gf-accent); }
.tint-rose { background: rgba(255, 107, 129, .14); color: var(--gf-danger); }
.tint-violet { background: rgba(139, 124, 255, .14); color: var(--gf-accent-2); }
.tint-mint { background: rgba(110, 231, 200, .13); color: var(--gf-accent-3); }

/* ==================== 状态分布 ==================== */
.status-list { display: flex; flex-direction: column; gap: 13px; }
.status-row { display: flex; align-items: center; gap: 10px; }
.status-name { width: 52px; flex-shrink: 0; font-size: 12px; color: var(--gf-text-dim); }

.status-track {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, .06);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, .45);
  overflow: hidden;
}

.status-fill { height: 100%; border-radius: 3px; min-width: 2px; transition: width .7s cubic-bezier(.22, 1, .36, 1); }
.status-fill.pending { background: linear-gradient(90deg, rgba(251, 191, 107, .6), var(--gf-warning)); }
.status-fill.active { background: var(--gf-gradient); }
.status-fill.progress { background: linear-gradient(90deg, rgba(124, 192, 255, .6), var(--gf-info)); }
.status-fill.done { background: linear-gradient(90deg, rgba(74, 222, 155, .6), var(--gf-success)); }
.status-fill.cancel { background: linear-gradient(90deg, rgba(255, 107, 129, .5), var(--gf-danger)); }

.status-count { width: 46px; text-align: right; flex-shrink: 0; font-size: 13px; font-weight: 700; color: var(--gf-text); font-variant-numeric: tabular-nums; }

/* ==================== 小列表（库存 / 用户） ==================== */
.mini-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }

.mini-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--gf-stroke);
}

.mini-row:last-child { border-bottom: none; }
.mini-row:first-child { padding-top: 0; }

.mini-name { flex: 1; min-width: 0; font-size: 13px; color: var(--gf-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mini-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.mini-body .mini-name { flex: none; }
.mini-sub { font-size: 11px; color: var(--gf-text-mute); }

.mini-avatar {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--accent-dim);
  color: var(--gf-accent);
  font-size: 13px;
  font-weight: 700;
}

.stock-chip { flex-shrink: 0; padding: 2px 9px; border-radius: var(--radius-pill); font-size: 11px; font-weight: 600; }
.stock-chip.warn { background: rgba(251, 191, 107, .13); color: var(--gf-warning); }
.stock-chip.danger { background: rgba(255, 107, 129, .14); color: var(--gf-danger); }

.admin-chip { flex-shrink: 0; padding: 2px 9px; border-radius: var(--radius-pill); background: var(--accent-dim); color: var(--gf-accent); font-size: 11px; font-weight: 600; }

.mini-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 26px 0;
  font-size: 13px;
  color: var(--gf-text-mute);
}

/* 空状态图标用中性色：这里是「暂无数据」的提示语境，不是成功反馈 */
.mini-empty svg { width: 18px; height: 18px; color: var(--gf-text-mute); }

/* ==================== 响应式 ==================== */
@media (max-width: 1280px) {
  .grid-thirds { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 1080px) {
  .stat-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .grid-main { grid-template-columns: minmax(0, 1fr); }
}

@media (max-width: 720px) {
  .stat-grid,
  .grid-thirds { grid-template-columns: minmax(0, 1fr); }
}
</style>
