<template>
  <div class="captcha-container">
    <div class="captcha-input-wrapper">
      <el-input
        v-model="inputValue"
        placeholder="请输入验证码"
        :maxlength="4"
        @input="handleInput"
        @keyup.enter="$emit('enter')"
      />
      <div class="captcha-image" @click="refreshCaptcha" :title="'点击刷新验证码'">
        <canvas ref="canvasRef" :width="width" :height="height"></canvas>
        <div class="captcha-refresh">
          <el-icon><Refresh /></el-icon>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'enter': []
}>()

const canvasRef = ref<HTMLCanvasElement>()
const inputValue = ref(props.modelValue || '')
const width = 140
const height = 45

// 生成随机验证码
const generateCode = (): string => {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789' // 排除容易混淆的字符
  let code = ''
  for (let i = 0; i < 4; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return code
}

let currentCode = ''

// 绘制验证码
const drawCaptcha = () => {
  if (!canvasRef.value) return
  
  const canvas = canvasRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  // 清空画布
  ctx.clearRect(0, 0, width, height)

  // 生成验证码
  currentCode = generateCode()

  // 设置背景
  ctx.fillStyle = '#f5f5f5'
  ctx.fillRect(0, 0, width, height)

  // 绘制干扰线
  for (let i = 0; i < 3; i++) {
    ctx.strokeStyle = `rgb(${Math.floor(Math.random() * 200)}, ${Math.floor(Math.random() * 200)}, ${Math.floor(Math.random() * 200)})`
    ctx.beginPath()
    ctx.moveTo(Math.random() * width, Math.random() * height)
    ctx.lineTo(Math.random() * width, Math.random() * height)
    ctx.stroke()
  }

  // 绘制验证码文字
  const fontSize = 20
  const fontFamily = 'Arial'
  ctx.font = `${fontSize}px ${fontFamily}`
  ctx.textBaseline = 'middle'

  for (let i = 0; i < currentCode.length; i++) {
    const char = currentCode[i]
    const x = (width / (currentCode.length + 1)) * (i + 1)
    const y = height / 2

    // 随机颜色
    ctx.fillStyle = `rgb(${Math.floor(Math.random() * 100)}, ${Math.floor(Math.random() * 100)}, ${Math.floor(Math.random() * 100)})`
    
    // 随机旋转
    ctx.save()
    ctx.translate(x, y)
    ctx.rotate((Math.random() - 0.5) * 0.5)
    ctx.fillText(char, 0, 0)
    ctx.restore()
  }

  // 绘制干扰点
  for (let i = 0; i < 30; i++) {
    ctx.fillStyle = `rgb(${Math.floor(Math.random() * 255)}, ${Math.floor(Math.random() * 255)}, ${Math.floor(Math.random() * 255)})`
    ctx.beginPath()
    ctx.arc(Math.random() * width, Math.random() * height, 1, 0, 2 * Math.PI)
    ctx.fill()
  }
}

// 刷新验证码
const refreshCaptcha = () => {
  drawCaptcha()
  inputValue.value = ''
  emit('update:modelValue', '')
}

// 处理输入
const handleInput = (value: string) => {
  inputValue.value = value.toUpperCase()
  emit('update:modelValue', value.toUpperCase())
}

// 验证验证码
const validate = (): boolean => {
  return inputValue.value.toUpperCase() === currentCode.toUpperCase()
}

// 暴露验证方法
defineExpose({
  validate,
  refreshCaptcha,
  getCode: () => currentCode,
})

// 监听 modelValue 变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== inputValue.value) {
    inputValue.value = newVal || ''
  }
})

onMounted(() => {
  drawCaptcha()
})
</script>

<style scoped>
.captcha-container {
  width: 100%;
}

.captcha-input-wrapper {
  display: flex;
  gap: 12px;
  align-items: center;
  width: 100%;
}

.captcha-input-wrapper :deep(.el-input) {
  flex: 1;
}

.captcha-image {
  position: relative;
  cursor: pointer;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  background: #f5f5f5;
  transition: all 0.3s;
  flex-shrink: 0;
}

.captcha-image:hover {
  border-color: #409eff;
}

.captcha-image canvas {
  display: block;
}

.captcha-refresh {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.05);
  opacity: 0;
  transition: opacity 0.3s;
}

.captcha-image:hover .captcha-refresh {
  opacity: 1;
}

.captcha-refresh .el-icon {
  font-size: 18px;
  color: #409eff;
}
</style>

