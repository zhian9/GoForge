import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// Element Plus 官方暗色变量：由 <html class="dark"> 激活，
// 这样组件库自带的表单/按钮/提示走暗色基调，再叠加下面的玻璃材质覆盖。
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

// 设计系统：token → 基座 → 玻璃材质与组件对齐。顺序不可调换，
// glass.css 需要覆盖 element-plus 的默认值，必须排在最后。
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

