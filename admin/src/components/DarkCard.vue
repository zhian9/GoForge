<template>
  <div class="dark-card" :class="{ hover: hover, 'no-pad': noPad }">
    <div class="card-header" v-if="title || $slots.header">
      <div class="card-title-wrap">
        <h3 class="card-title">{{ title }}</h3>
        <p v-if="desc" class="card-desc">{{ desc }}</p>
      </div>
      <div class="card-extra"><slot name="header" /></div>
    </div>
    <div class="card-body"><slot /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title?: string; desc?: string; hover?: boolean; noPad?: boolean }>()
</script>

<style scoped>
.dark-card {
  position: relative;
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  border-radius: var(--radius);
  box-shadow: var(--gf-inner-shadow-soft);
  overflow: hidden;
  transition: border-color .3s ease, transform .3s cubic-bezier(.22, 1, .36, 1), box-shadow .3s ease;
}

.dark-card.hover:hover {
  border-color: var(--gf-stroke-strong);
  transform: translateY(-3px);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow-soft);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--gf-stroke);
}

.card-title-wrap { min-width: 0; }

.card-title {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--gf-text);
  letter-spacing: -.01em;
}

/* 标题前的渐变短竖条：全站标题统一标识，与用户端一致 */
.card-title::before {
  content: '';
  width: 3px;
  height: 15px;
  border-radius: 2px;
  background: var(--gf-gradient);
  box-shadow: 0 0 10px var(--accent-glow);
}

.card-desc { margin: 5px 0 0 12px; font-size: 12px; color: var(--gf-text-mute); }
.card-extra { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.card-body { padding: 20px; }
.dark-card.no-pad .card-body { padding: 0; }
</style>

