<template>
  <div class="login-page">
    <div class="bg-glow bg-glow-1"></div>
    <div class="bg-glow bg-glow-2"></div>

    <div class="login-card">
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
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const formData = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
}

const handleLogin = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return; loading.value = true
    try { await userStore.login({ username: formData.username, password: formData.password }); ElMessage.success('登录成功'); router.push('/') }
    catch (e: any) { ElMessage.error(e.message || '登录失败') }
    finally { loading.value = false }
  })
}
</script>

<style scoped>
.login-page { display: flex; align-items: center; justify-content: center; width: 100%; min-height: 100vh; background: #0A0F1C; position: relative; overflow: hidden; }
.bg-glow { position: absolute; border-radius: 50%; filter: blur(140px); opacity: .2; pointer-events: none; }
.bg-glow-1 { width: 500px; height: 500px; background: radial-gradient(circle, #00F5FF, transparent); top: -15%; right: -20%; }
.bg-glow-2 { width: 400px; height: 400px; background: radial-gradient(circle, rgba(139,92,246,.5), transparent); bottom: -10%; left: -10%; }

.login-card { width: 400px; padding: 44px 40px 36px; background: rgba(255,255,255,0.025); border: 1px solid rgba(255,255,255,0.06); border-radius: 20px; position: relative; z-index: 1; backdrop-filter: blur(12px); box-shadow: 0 32px 64px rgba(0,0,0,.4); }
.login-header { text-align: center; margin-bottom: 32px; }
.login-logo { height: 38px; margin-bottom: 18px; }
.login-header h2 { font-size: 22px; font-weight: 700; color: #EDF0F5; margin: 0; }

.btn-submit { width: 100%; height: 44px; border-radius: 12px; font-size: 15px; font-weight: 700; letter-spacing: .06em; margin-top: 4px; }
.form-footer { text-align: center; margin-top: 12px; }
.form-footer a { color: #8890A5; font-size: 13px; text-decoration: none; transition: color .2s; }
.form-footer a:hover { color: #00F5FF; }
</style>
