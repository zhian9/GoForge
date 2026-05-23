<template>
  <div class="page-header">
    <h2 class="title">{{ title }}</h2>
    <p class="desc" v-if="desc">{{ desc }}</p>
    <div class="actions" v-if="$slots.actions"><slot name="actions" /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; desc?: string }>()
</script>

<style scoped>
.page-header { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
.title { font-size: 22px; font-weight: 700; color: #EDF0F5; margin: 0; letter-spacing: -.01em; }
.desc { font-size: 13px; color: #8890A5; margin: 0; }
.actions { margin-left: auto; display: flex; align-items: center; gap: 8px; }
</style>
