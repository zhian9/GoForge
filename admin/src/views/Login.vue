<template>
  <div class="login-page">
    <div class="gf-aurora" aria-hidden="true">
      <span class="gf-aurora__blob gf-aurora__blob--1"></span>
      <span class="gf-aurora__blob gf-aurora__blob--2"></span>
    </div>

    <div class="login-card gf-rise">
      <div class="login-header">
        <img src="/logo-forge.png" alt="GoForge" class="login-logo" />
        <h2>GoForge 管理后台</h2>
        <p>仅限管理员登录</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="UserIcon" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" show-password :prefix-icon="LockIcon" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="btn-login" :loading="loading" @click="handleLogin">
            登 录
          </el-button>
        </el-form-item>
      </el-form>

      <p class="login-foot">© 2026 GoForge · 仅供内部运营使用</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, h } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

// SVG icons as render functions
const UserIcon = h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', style: 'width:18px;height:18px' }, [
  h('circle', { cx: '12', cy: '8', r: '4' }),
  h('path', { d: 'M4 20c0-4 4-7 8-7s8 3 8 7' }),
])
const LockIcon = h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', style: 'width:18px;height:18px' }, [
  h('rect', { x: '3', y: '11', width: '18', height: '11', rx: '2', ry: '2' }),
  h('path', { d: 'M7 11V7a5 5 0 0110 0v4' }),
])

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
}

const handleLogin = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await userStore.login({ username: form.username, password: form.password }) as any
      const isAdmin = (res?.data?.isAdmin ?? res?.data?.is_admin ?? (userStore.userInfo as any)?.isAdmin ?? (userStore.userInfo as any)?.is_admin ?? 0)
      if (Number(isAdmin) !== 1) {
        userStore.logout()
        ElMessage.error('您不是管理员，无法登录管理后台')
        return
      }
      ElMessage.success('登录成功')
      router.push('/')
    } catch (e: any) {
      ElMessage.error(e.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-page {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  font-family: var(--gf-font);
}

.login-card {
  position: relative;
  z-index: 1;
  width: 420px;
  max-width: calc(100vw - 40px);
  padding: 44px 40px 32px;
  border-radius: var(--radius-lg);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}

.login-header { text-align: center; margin-bottom: 32px; }
.login-logo { height: 42px; margin-bottom: 18px; }
.login-header h2 { margin: 0 0 8px; font-size: 23px; font-weight: 700; color: var(--gf-text); letter-spacing: -.01em; }
.login-header p { margin: 0; font-size: 13px; color: var(--gf-text-dim); }

.btn-login {
  width: 100%;
  height: 46px;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 700;
  letter-spacing: .06em;
  margin-top: 6px;
}

.login-foot {
  margin: 22px 0 0;
  text-align: center;
  font-size: 11px;
  color: var(--gf-text-mute);
}
</style>
