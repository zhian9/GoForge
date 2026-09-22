<template>
  <div class="login-page">
    <div class="gf-aurora" aria-hidden="true">
      <span class="gf-aurora__blob gf-aurora__blob--1"></span>
      <span class="gf-aurora__blob gf-aurora__blob--2"></span>
    </div>
    <div class="gf-grain" aria-hidden="true"></div>
    <div class="bg-glow bg-glow-1"></div>
    <div class="bg-glow bg-glow-2"></div>

    <router-link to="/" class="btn-home">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 10.5 12 3l9 7.5"/><path d="M5.5 9.5V21h13V9.5"/></svg>
      返回首页
    </router-link>

    <div class="login-card gf-rise">
      <div class="login-header">
        <img src="/logo-forge.png" alt="GoForge" class="login-logo" />
        <h2>用户登录</h2>
      </div>

      <el-form ref="formRef" :model="formData" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="formData.username" placeholder="用户名" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="formData.password" type="password" placeholder="密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="btn-submit" :loading="loading" @click="handleLogin">登 录</el-button>
        </el-form-item>
        <div class="form-footer">
          <router-link to="/register">没有账号？立即注册</router-link>
          <span class="footer-sep">·</span>
          <router-link to="/">先去逛逛</router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { normalizeRedirect } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
// 登录成功后回到用户原本想去的页面（如加购时被拦下的商品详情、购物车）
const redirect = computed(() => normalizeRedirect(route.query.redirect as string))
const formData = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
}

const handleLogin = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return; loading.value = true
    try { await userStore.login({ username: formData.username, password: formData.password }); ElMessage.success('登录成功'); router.replace(redirect.value) }
    catch (e: any) { ElMessage.error(e.message || '登录失败') }
    finally { loading.value = false }
  })
}
</script>

<style scoped>
.login-page { display: flex; align-items: center; justify-content: center; width: 100%; min-height: 100vh; position: relative; overflow: hidden; }

/* 返回首页：贴在左上角的浮层胶囊，让登录页不再是“死胡同” */
.btn-home {
  position: absolute; top: 28px; left: 28px; z-index: 2;
  display: inline-flex; align-items: center; gap: 8px; padding: 10px 18px;
  background: var(--gf-glass-1); border: 1px solid var(--gf-stroke); border-radius: var(--radius-pill);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--text-dim); font-size: 13px; font-weight: 500; text-decoration: none; cursor: pointer;
  transition: color .2s, border-color .2s, background .2s, transform .25s cubic-bezier(.22,1,.36,1);
}
.btn-home svg { width: 16px; height: 16px; }
.btn-home:hover { color: var(--accent); border-color: rgba(79,216,255,.5); background: var(--gf-glass-2); transform: translateX(-2px); }

.bg-glow { position: absolute; border-radius: 50%; filter: blur(140px); opacity: .22; pointer-events: none; }
.bg-glow-1 { width: 500px; height: 500px; background: radial-gradient(circle, var(--accent), transparent); top: -15%; right: -20%; }
.bg-glow-2 { width: 400px; height: 400px; background: radial-gradient(circle, rgba(139,124,255,.55), transparent); bottom: -10%; left: -10%; }

/* 登录卡是页面上唯一的主容器，用最厚的一档玻璃 + 28px 大圆角 */
.login-card {
  width: 400px; padding: 44px 40px 36px; position: relative; z-index: 1;
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong); border-radius: var(--radius-lg);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}
.login-header { text-align: center; margin-bottom: 32px; }
.login-logo { height: 38px; margin-bottom: 18px; }
.login-header h2 { font-size: 22px; font-weight: 700; color: var(--text); margin: 0; }

.btn-submit { width: 100%; height: 46px; border-radius: var(--radius-sm); font-size: 15px; font-weight: 700; letter-spacing: .06em; margin-top: 4px; }
.form-footer { display: flex; align-items: center; justify-content: center; gap: 8px; text-align: center; margin-top: 12px; }
.form-footer .footer-sep { color: var(--text-dim); opacity: .4; }
.form-footer a { color: var(--text-dim); font-size: 13px; text-decoration: none; transition: color .2s; }
.form-footer a:hover { color: var(--accent); }

@media (max-width: 520px) {
  .btn-home { top: 16px; left: 16px; padding: 8px 14px; font-size: 12px; }
}
</style>
