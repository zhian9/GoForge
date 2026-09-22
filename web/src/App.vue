<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
/**
 * 全局样式已迁到 src/styles/（tokens.css / base.css / glass.css），由 main.ts 统一引入。
 *
 * 这里原本把 --el-color-primary 设成橙色 #ff6900、body 背景设成浅灰 #f5f5f5，
 * 而各页面自身是深底 #0A0F1C + 青色 #00F5FF —— 于是 Element Plus 组件在用户端
 * 呈现为橙色浅色系统，夹在深色页面之间是断裂的。现已统一到设计 token。
 */
</style>

