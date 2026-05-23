<template>
  <div class="dark-card" :class="{ hover: hover, 'no-pad': noPad }">
    <div class="card-header" v-if="title || $slots.header">
      <h3 class="card-title">{{ title }}</h3>
      <div class="card-extra"><slot name="header" /></div>
    </div>
    <div class="card-body"><slot /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title?: string; hover?: boolean; noPad?: boolean }>()
</script>

<style scoped>
.dark-card {
  background: rgba(255,255,255,0.025);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 14px;
  overflow: hidden;
  transition: all .3s ease;
}
.dark-card.hover:hover {
  border-color: rgba(255,255,255,0.12);
  transform: translateY(-2px);
  box-shadow: 0 12px 36px rgba(0,0,0,.3);
}
.card-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18px 22px; border-bottom: 1px solid rgba(255,255,255,0.06);
}
.card-title { font-size: 15px; font-weight: 600; color: #EDF0F5; margin: 0; }
.card-extra { display: flex; align-items: center; gap: 8px; }
.card-body { padding: 22px; }
.dark-card.no-pad .card-body { padding: 0; }
</style>
