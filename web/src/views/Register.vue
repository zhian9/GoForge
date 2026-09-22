<template>
  <div class="register-page">
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

    <div class="register-card gf-rise">
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
          <span class="footer-sep">·</span>
          <router-link to="/">先去逛逛</router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import Captcha from '@/components/Captcha.vue'
import { loginLocation } from '@/utils/auth'

const route = useRoute()
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
      // 注册前若被登录拦截过（如加购），把回跳地址透传给登录页
      router.push(loginLocation(route.query.redirect as string))
    } catch (e: any) { ElMessage.error(e.message || '注册失败') }
    finally { loading.value = false }
  })
}
</script>

<style scoped>
.register-page { display: flex; align-items: center; justify-content: center; width: 100%; min-height: 100vh; position: relative; overflow: hidden; }
.bg-glow { position: absolute; border-radius: 50%; filter: blur(140px); opacity: .22; pointer-events: none; }

.bg-glow-1 { width: 500px; height: 500px; background: radial-gradient(circle, var(--accent), transparent); top: -15%; right: -20%; }
.bg-glow-2 { width: 400px; height: 400px; background: radial-gradient(circle, rgba(139,124,255,.55), transparent); bottom: -10%; left: -10%; }

/* 与登录页保持一致的「返回首页」浮层胶囊 */
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

.register-card {
  width: 420px; padding: 40px 40px 32px; position: relative; z-index: 1;
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong); border-radius: var(--radius-lg);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}
.register-header { text-align: center; margin-bottom: 28px; }
.register-logo { height: 38px; margin-bottom: 18px; }
.register-header h2 { font-size: 22px; font-weight: 700; color: var(--text); margin: 0; }

.btn-submit { width: 100%; height: 46px; border-radius: var(--radius-sm); font-size: 15px; font-weight: 700; letter-spacing: .06em; margin-top: 4px; }
.form-footer { display: flex; align-items: center; justify-content: center; gap: 8px; text-align: center; margin-top: 8px; }
.form-footer .footer-sep { color: var(--text-dim); opacity: .4; }
.form-footer a { color: var(--text-dim); font-size: 13px; text-decoration: none; transition: color .2s; }
.form-footer a:hover { color: var(--accent); }

@media (max-width: 520px) {
  .btn-home { top: 16px; left: 16px; padding: 8px 14px; font-size: 12px; }
}
</style>
