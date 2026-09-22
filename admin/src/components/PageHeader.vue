<template>
  <div class="page-header">
    <div class="title-wrap">
      <h2 class="title">{{ title }}</h2>
      <p class="desc" v-if="desc">{{ desc }}</p>
    </div>
    <div class="actions" v-if="$slots.actions"><slot name="actions" /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; desc?: string }>()
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.title-wrap { min-width: 0; }

.title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--gf-text);
  letter-spacing: -.01em;
}

/* 标题前的渐变竖条，与卡片标题同一套视觉语言 */
.title::before {
  content: '';
  width: 3px;
  height: 18px;
  border-radius: 2px;
  background: var(--gf-gradient);
  box-shadow: 0 0 12px var(--accent-glow);
}

.desc { margin: 7px 0 0 13px; font-size: 13px; color: var(--gf-text-dim); }
.actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
</style>
