<template>
  <div class="register-page">
    <div class="bg-glow bg-glow-1"></div>
    <div class="bg-glow bg-glow-2"></div>

    <div class="register-card">
      <div class="register-header">
        <img src="/logo-forge.png" alt="GoForge" class="register-logo" />
        <h2>用户注册</h2>
      </div>

      <el-form ref="formRef" :model="formData" :rules="rules" size="large">
        <el-form-item prop="username">
          <el-input v-model="formData.username" placeholder="用户名" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="formData.password" type="password" placeholder="密码" show-password />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input v-model="formData.confirmPassword" type="password" placeholder="确认密码" show-password />
        </el-form-item>
        <el-form-item prop="phone">
          <el-input v-model="formData.phone" placeholder="手机号（选填）" />
        </el-form-item>
        <el-form-item prop="email">
          <el-input v-model="formData.email" placeholder="邮箱（选填）" />
        </el-form-item>
        <el-form-item prop="verifyCode">
          <Captcha v-model="formData.verifyCode" ref="captchaRef" @enter="handleRegister" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="btn-submit" :loading="loading" @click="handleRegister">注 册</el-button>
        </el-form-item>
        <div class="form-footer">
          <router-link to="/login">已有账号？立即登录</router-link>
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
import Captcha from '@/components/Captcha.vue'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const captchaRef = ref<InstanceType<typeof Captcha>>()
const loading = ref(false)

const formData = reactive({ username: '', password: '', confirmPassword: '', phone: '', email: '', verifyCode: '' })

const validateConfirmPassword = (_r: any, v: string, cb: any) => { if (v !== formData.password) cb(new Error('密码不一致')); else cb() }
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
  confirmPassword: [{ required: true, message: '请确认密码', trigger: 'blur' }, { validator: validateConfirmPassword, trigger: 'blur' }],
  verifyCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

const handleRegister = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return; loading.value = true
    try {
      await userStore.register({ username: formData.username, password: formData.password, phone: formData.phone, email: formData.email, verify_code: formData.verifyCode })
      ElMessage.success('注册成功，请登录')
      router.push('/login')
    } catch (e: any) { ElMessage.error(e.message || '注册失败') }
    finally { loading.value = false }
  })
}
</script>

<style scoped>
.register-page { display: flex; align-items: center; justify-content: center; width: 100%; min-height: 100vh; background: #0A0F1C; position: relative; overflow: hidden; }
.bg-glow { position: absolute; border-radius: 50%; filter: blur(140px); opacity: .2; pointer-events: none; }
.bg-glow-1 { width: 500px; height: 500px; background: radial-gradient(circle, #00F5FF, transparent); top: -15%; right: -20%; }
.bg-glow-2 { width: 400px; height: 400px; background: radial-gradient(circle, rgba(139,92,246,.5), transparent); bottom: -10%; left: -10%; }

.register-card { width: 420px; padding: 40px 40px 32px; background: rgba(255,255,255,0.025); border: 1px solid rgba(255,255,255,0.06); border-radius: 20px; position: relative; z-index: 1; backdrop-filter: blur(12px); box-shadow: 0 32px 64px rgba(0,0,0,.4); }
.register-header { text-align: center; margin-bottom: 28px; }
.register-logo { height: 38px; margin-bottom: 18px; }
.register-header h2 { font-size: 22px; font-weight: 700; color: #EDF0F5; margin: 0; }

.btn-submit { width: 100%; height: 44px; border-radius: 12px; font-size: 15px; font-weight: 700; letter-spacing: .06em; margin-top: 4px; }
.form-footer { text-align: center; margin-top: 8px; }
.form-footer a { color: #8890A5; font-size: 13px; text-decoration: none; transition: color .2s; }
.form-footer a:hover { color: #00F5FF; }
</style>
