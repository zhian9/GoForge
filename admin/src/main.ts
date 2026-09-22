import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

// 设计系统：token → 基座 → 玻璃材质与 Element Plus 对齐。
// 顺序不可调换，glass.css 需要覆盖组件库默认值，必须排在最后。
import '@/styles/tokens.css'
import '@/styles/base.css'
import '@/styles/glass.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, {
  locale: zhCn,
})

app.mount('#app')

