<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', 'Microsoft YaHei', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #333;
  background-color: #f5f5f5;
}

#app {
  min-height: 100vh;
}

/* Element Plus 主题色覆盖 */
:root {
  --el-color-primary: #ff6900;
}

.el-button--primary {
  background-color: #ff6900;
  border-color: #ff6900;
}

.el-button--primary:hover {
  background-color: #ff8533;
  border-color: #ff8533;
}
</style>

