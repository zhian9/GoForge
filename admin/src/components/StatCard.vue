<template>
  <div class="stat-card" :class="color">
    <!-- 右上角同色辉光：让四张卡在扫视时有色彩区分，但不抢数字 -->
    <span class="card-glow" aria-hidden="true"></span>

    <div class="card-icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" v-html="icon"></svg>
    </div>

    <div class="card-info">
      <span class="card-label">{{ label }}</span>
      <span class="card-value">{{ value }}</span>
      <span v-if="hint" class="card-hint" :class="hintTone">{{ hint }}</span>
    </div>

    <div class="card-trend" v-if="trend" :class="trend > 0 ? 'up' : 'down'">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline v-if="trend > 0" points="23 6 13.5 15.5 8.5 10.5 1 18" /><polyline v-else points="23 18 13.5 8.5 8.5 13.5 1 6" /></svg>
      {{ Math.abs(trend) }}%
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  icon: string
  value: string
  label: string
  trend?: number
  color?: string
  /** 副指标：一行小字补充说明（今日新增、客单价等） */
  hint?: string
  hintTone?: 'default' | 'ok' | 'warn' | 'info'
}>(), {
  trend: 0,
  color: 'cyan',
  hint: '',
  hintTone: 'default',
})
</script>

<style scoped>
.stat-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 18px 20px;
  border-radius: var(--radius);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
  overflow: hidden;
  transition: transform .3s cubic-bezier(.22, 1, .36, 1), border-color .3s ease, box-shadow .3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  border-color: var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
}

.card-glow {
  position: absolute;
  top: -40%;
  right: -18%;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  filter: blur(46px);
  opacity: .55;
  pointer-events: none;
}

.card-icon {
  position: relative;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  box-shadow: var(--gf-inner-shadow-soft);
}

.card-icon :deep(svg) { width: 21px; height: 21px; }

/* 色板：图标承载颜色，数值保持中性白，避免四个彩色数字互相打架 */
.cyan   .card-icon { background: rgba(79, 216, 255, .13);  color: var(--gf-accent); }
.cyan   .card-glow { background: radial-gradient(circle, rgba(79, 216, 255, .34), transparent 70%); }
.blue   .card-icon { background: rgba(124, 192, 255, .13); color: var(--gf-info); }
.blue   .card-glow { background: radial-gradient(circle, rgba(124, 192, 255, .32), transparent 70%); }
.green  .card-icon { background: rgba(74, 222, 155, .13);  color: var(--gf-success); }
.green  .card-glow { background: radial-gradient(circle, rgba(74, 222, 155, .3), transparent 70%); }
.red    .card-icon { background: rgba(232, 200, 138, .14); color: var(--gf-gold); }
.red    .card-glow { background: radial-gradient(circle, rgba(232, 200, 138, .3), transparent 70%); }
.purple .card-icon { background: rgba(139, 124, 255, .14); color: var(--gf-accent-2); }
.purple .card-glow { background: radial-gradient(circle, rgba(139, 124, 255, .32), transparent 70%); }

.card-info { position: relative; flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px; }
.card-label { font-size: 12px; color: var(--gf-text-dim); font-weight: 500; }
.card-value { font-size: 24px; font-weight: 800; color: var(--gf-text); letter-spacing: -.02em; line-height: 1.15; font-variant-numeric: tabular-nums; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-hint { font-size: 11px; color: var(--gf-text-mute); }
.card-hint.warn { color: var(--gf-warning); }
.card-hint.ok { color: var(--gf-success); }
.card-hint.info { color: var(--gf-info); }

.card-trend {
  position: relative;
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  font-weight: 600;
}

.card-trend.up { color: var(--gf-success); }
.card-trend.down { color: var(--gf-danger); }
.card-trend svg { width: 40px; height: 20px; flex-shrink: 0; }
</style>

