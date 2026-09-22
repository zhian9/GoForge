<template>
  <router-view />
</template>

<script setup lang="ts">
// App root component
</script>

<style>
/**
 * 全局样式已迁到 src/styles/（tokens.css / base.css / glass.css），由 main.ts 统一引入。
 *
 * 这里原本是 150 行写死色值的 !important 覆盖（#00F5FF / #111827 等），
 * 与用户端新色板脱节且无法按组件维护，现已全部收敛到 glass.css。
 */
</style>

