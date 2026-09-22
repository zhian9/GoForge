<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="modelValue" class="modal-overlay" @click.self="onCancel">
        <div class="modal-card" :class="[sizeClass]">
          <!-- Close -->
          <button class="modal-close" type="button" aria-label="关闭" @click="onCancel">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>

          <!-- Title -->
          <div class="modal-header">
            <h3 class="modal-title">{{ title }}</h3>
            <p class="modal-desc" v-if="desc">{{ desc }}</p>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot />
          </div>

          <!-- Footer -->
          <div class="modal-footer" v-if="!hideFooter">
            <button class="btn-cancel" type="button" @click="onCancel" :disabled="loading">{{ cancelText }}</button>
            <button class="btn-confirm" type="button" @click="onConfirm" :disabled="loading || confirmDisabled">
              <span v-if="loading" class="spinner"></span>
              {{ loading ? loadingText : confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title: string
  desc?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  confirmText?: string
  cancelText?: string
  loadingText?: string
  loading?: boolean
  confirmDisabled?: boolean
  hideFooter?: boolean
}>(), {
  size: 'md',
  confirmText: '确定',
  cancelText: '取消',
  loadingText: '提交中...',
  loading: false,
  confirmDisabled: false,
  hideFooter: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
  cancel: []
}>()

const sizeClass = computed(() => `size-${props.size}`)

const onConfirm = () => emit('confirm')
const onCancel = () => {
  if (!props.loading) {
    emit('update:modelValue', false)
    emit('cancel')
  }
}
</script>

<style scoped>
/* ==================== 遮罩 ==================== */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(4, 6, 12, 0.62);
  -webkit-backdrop-filter: blur(6px);
  backdrop-filter: blur(6px);
}

/* ==================== 卡片 ==================== */
.modal-card {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: 85vh;
  border-radius: var(--radius-lg);
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
  overflow: hidden;
}

.size-sm { max-width: 420px; }
.size-md { max-width: 560px; }
.size-lg { max-width: 720px; }
.size-xl { max-width: 900px; }

/* ==================== 关闭按钮 ==================== */
.modal-close {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 1;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  color: var(--gf-text-mute);
  cursor: pointer;
  transition: color .2s ease, border-color .2s ease, background .2s ease;
}

.modal-close svg { width: 15px; height: 15px; }
.modal-close:hover { color: var(--gf-danger); border-color: rgba(255, 107, 129, 0.35); background: rgba(255, 107, 129, 0.1); }

/* ==================== 头 / 体 / 尾 ==================== */
.modal-header { padding: 22px 24px 0; }

.modal-title {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: var(--gf-text);
  letter-spacing: -.01em;
}

.modal-title::before {
  content: '';
  width: 3px;
  height: 16px;
  border-radius: 2px;
  background: var(--gf-gradient);
  box-shadow: 0 0 10px var(--accent-glow);
}

.modal-desc { margin: 7px 0 0 12px; font-size: 12px; color: var(--gf-text-mute); }

.modal-body { flex: 1; overflow-y: auto; padding: 20px 24px; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-shrink: 0;
  padding: 16px 24px;
  border-top: 1px solid var(--gf-stroke);
}

/* ==================== 按钮 ==================== */
.btn-cancel,
.btn-confirm {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 22px;
  border-radius: var(--radius-sm);
  font-family: var(--gf-font);
  font-size: 14px;
  cursor: pointer;
  transition: transform .2s cubic-bezier(.22, 1, .36, 1), box-shadow .25s ease, border-color .2s ease, background .2s ease;
}

.btn-cancel {
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--gf-text-dim);
  font-weight: 500;
}

.btn-cancel:hover:not(:disabled) { border-color: var(--gf-stroke-strong); color: var(--gf-text); background: var(--gf-glass-2); }
.btn-cancel:disabled { opacity: .45; cursor: not-allowed; }

.btn-confirm {
  min-width: 76px;
  border: none;
  background: var(--gf-gradient);
  color: #04121a;
  font-weight: 700;
  box-shadow: var(--gf-inner-shadow-soft), 0 8px 22px -12px var(--accent-glow);
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  filter: brightness(1.06);
  box-shadow: var(--gf-inner-shadow-soft), 0 14px 30px -14px var(--accent-glow);
}

.btn-confirm:disabled { opacity: .5; cursor: not-allowed; filter: none; }

/* ==================== 加载指示 ==================== */
.spinner {
  width: 15px;
  height: 15px;
  border: 2px solid rgba(4, 18, 26, 0.3);
  border-top-color: #04121a;
  border-radius: 50%;
  animation: spin .6s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* ==================== 过渡 ==================== */
.modal-fade-enter-active,
.modal-fade-leave-active { transition: opacity .2s ease; }
.modal-fade-enter-active .modal-card,
.modal-fade-leave-active .modal-card { transition: transform .24s cubic-bezier(.22, 1, .36, 1); }
.modal-fade-enter-from,
.modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-from .modal-card { transform: scale(.96) translateY(8px); }
.modal-fade-leave-to .modal-card { transform: scale(.98); }
</style>
