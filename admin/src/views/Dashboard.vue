<template>
  <div class="dashboard">
    <PageHeader title="仪表盘" desc="GoForge 平台核心数据概览" />

    <!-- KPI 卡片 -->
    <div class="stat-grid">
      <StatCard
        v-for="card in kpiCards"
        :key="card.label"
        :icon="card.icon"
        :value="card.value"
        :label="card.label"
        :trend="card.trend"
        :color="card.color"
      />
    </div>

    <!-- 下方区域示例 -->
    <div class="bottom-grid">
      <DarkCard title="最近订单" hover style="grid-column:span 2">
        <el-table :data="recentOrders" size="small" stripe>
          <el-table-column prop="orderNo" label="订单号" width="180" />
          <el-table-column prop="user" label="用户" width="120" />
          <el-table-column prop="amount" label="金额" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span class="status-tag" :class="row.status==='已完成'?'done':'pending'">{{ row.status }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="time" label="时间" />
        </el-table>
      </DarkCard>

      <DarkCard title="快速操作">
        <div class="quick-actions">
          <button v-for="a in quickActions" :key="a.label" class="qa-btn" @click="$router.push(a.path)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" v-html="a.icon"></svg>
            <span>{{ a.label }}</span>
          </button>
        </div>
      </DarkCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'
import StatCard from '@/components/StatCard.vue'

const kpiCards = [
  { label: '总用户数', value: '1,234', trend: 12, color: 'cyan', icon: '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 4-7 8-7s8 3 8 7"/>' },
  { label: '商品总数', value: '5,678', trend: 8, color: 'blue', icon: '<path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4zM3 6h18"/><path d="M16 10a4 4 0 01-8 0"/>' },
  { label: '订单总数', value: '9,012', trend: 24, color: 'green', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/>' },
  { label: '总销售额', value: '¥123K', trend: -3, color: 'red', icon: '<rect x="2" y="4" width="20" height="16" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/>' },
]

const recentOrders = [
  { orderNo: 'ORD20260512000001', user: 'testuser99', amount: '¥4,999', status: '已完成', time: '2026-05-12 10:30' },
  { orderNo: 'ORD20260512000002', user: 'testuser99', amount: '¥6,999', status: '待支付', time: '2026-05-12 09:15' },
  { orderNo: 'ORD20260511000003', user: 'testuser99', amount: '¥3,299', status: '已完成', time: '2026-05-11 18:00' },
]

const quickActions = [
  { label: '添加商品', path: '/products', icon: '<path d="M12 5v14M5 12h14"/>' },
  { label: '新建秒杀', path: '/seckill-activities', icon: '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>' },
  { label: '查看订单', path: '/orders', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/>' },
  { label: '用户管理', path: '/users', icon: '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 4-7 8-7s8 3 8 7"/>' },
]
</script>

<style scoped>
.dashboard { padding: 4px 0; }
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.bottom-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 18px; }

.status-tag { display: inline-block; padding: 2px 10px; border-radius: 100px; font-size: 11px; font-weight: 600; }
.status-tag.done { background: rgba(16,185,129,.12); color: #10B981; }
.status-tag.pending { background: rgba(245,158,11,.12); color: #F59E0B; }

.quick-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.qa-btn { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 16px 10px; border-radius: 12px; border: 1px solid rgba(255,255,255,0.06); background: rgba(255,255,255,0.02); color: #8890A5; font-size: 12px; cursor: pointer; transition: all .2s; }
.qa-btn:hover { border-color: #00F5FF; color: #00F5FF; background: rgba(0,245,255,0.06); }
.qa-btn :deep(svg) { width: 22px; height: 22px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }

@media(max-width:900px){ .stat-grid{grid-template-columns:repeat(2,1fr)} .bottom-grid{grid-template-columns:1fr} }
@media(max-width:500px){ .stat-grid{grid-template-columns:1fr} }
</style>
