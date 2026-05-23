<template>
  <div class="login-page">
    <!-- 背景光晕 -->
    <div class="bg-glow bg-glow-1"></div>
    <div class="bg-glow bg-glow-2"></div>

    <div class="login-card">
      <!-- Logo -->
      <div class="login-header">
        <img src="/logo-forge.png" alt="GoForge" class="login-logo" />
        <h2>GoForge 管理后台</h2>
        <p>仅限管理员登录</p>
      </div>

      <!-- 表单 -->
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
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 100vh; background: #0A0F1C;
  position: relative; overflow: hidden;
}

/* 背景光晕 */
.bg-glow { position: absolute; border-radius: 50%; filter: blur(140px); opacity: .22; pointer-events: none; }
.bg-glow-1 { width: 520px; height: 520px; background: radial-gradient(circle, #00F5FF, transparent); top: -20%; right: -15%; }
.bg-glow-2 { width: 420px; height: 420px; background: radial-gradient(circle, rgba(139,92,246,.5), transparent); bottom: -10%; left: -10%; }
.bg-glow-2::after {
  content: ''; position: absolute; inset: 0; border-radius: 50%;
  background: radial-gradient(circle, rgba(0,245,255,.3), transparent); opacity: .5;
}

/* 卡片 */
.login-card {
  width: 420px; padding: 48px 44px 40px;
  background: rgba(255,255,255,0.025);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 20px; position: relative; z-index: 1;
  backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 32px 64px rgba(0,0,0,.4);
}

/* Header */
.login-header { text-align: center; margin-bottom: 36px; }
.login-logo { height: 42px; margin-bottom: 20px; }
.login-header h2 { font-size: 24px; font-weight: 700; color: #EDF0F5; margin: 0 0 8px; letter-spacing: -.01em; }
.login-header p { font-size: 13px; color: #8890A5; margin: 0; }

/* 按钮 */
.btn-login { width: 100%; height: 46px; border-radius: 12px; font-size: 16px; font-weight: 700; letter-spacing: .06em; margin-top: 8px; }
</style>
