<template>
  <div class="stat-card" :class="{ [color]: true }">
    <div class="card-icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" v-html="icon"></svg>
    </div>
    <div class="card-info">
      <span class="card-value">{{ value }}</span>
      <span class="card-label">{{ label }}</span>
    </div>
    <div class="card-trend" v-if="trend" :class="trend > 0 ? 'up' : 'down'">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline v-if="trend>0" points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline v-else points="23 18 13.5 8.5 8.5 13.5 1 6"/></svg>
      {{ Math.abs(trend) }}%
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ icon: string; value: string; label: string; trend?: number; color?: string }>()
</script>

<style scoped>
.stat-card {
  background: rgba(255,255,255,0.025);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 16px;
  padding: 22px 24px;
  display: flex; align-items: center; gap: 14px;
  transition: all .3s;
  position: relative; overflow: hidden;
}
.stat-card:hover {
  transform: translateY(-3px);
  border-color: rgba(255,255,255,0.12);
  box-shadow: 0 12px 40px rgba(0,0,0,.3);
}
.card-icon {
  width: 44px; height: 44px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.card-icon :deep(svg) { width: 22px; height: 22px; }
.cyan  .card-icon { background: rgba(0,245,255,.1); color: #00F5FF; }
.blue  .card-icon { background: rgba(59,130,246,.1); color: #3B82F6; }
.green .card-icon { background: rgba(16,185,129,.1); color: #10B981; }
.red   .card-icon { background: rgba(248,113,113,.1); color: #F87171; }
.purple.card-icon { background: rgba(139,92,246,.1); color: #8B5CF6; }
.card-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.card-value { font-size: 24px; font-weight: 800; color: #00F5FF; letter-spacing: -.02em; }
.card-label { font-size: 12px; color: #8890A5; font-weight: 500; }
.card-trend { font-size: 12px; font-weight: 600; display: flex; align-items: center; gap: 3px; }
.card-trend.up { color: #10B981; } .card-trend.down { color: #F87171; }
.card-trend svg { width: 40px; height: 20px; flex-shrink: 0; }
</style>
