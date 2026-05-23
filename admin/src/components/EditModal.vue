<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="modelValue" class="modal-overlay" @click.self="onCancel">
        <div class="modal-card" :class="[sizeClass]">
          <!-- Close -->
          <button class="modal-close" @click="onCancel">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
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
            <button class="btn-cancel" @click="onCancel" :disabled="loading">{{ cancelText }}</button>
            <button class="btn-confirm" @click="onConfirm" :disabled="loading || confirmDisabled">
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
/* Overlay */
.modal-overlay {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0, 0, 0, 0.55); backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  padding: 20px;
}

/* Card */
.modal-card {
  position: relative;
  background: #111827;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.04);
  max-height: 85vh; display: flex; flex-direction: column;
  width: 100%;
}
/* Sizes */
.size-sm { max-width: 420px; }
.size-md { max-width: 560px; }
.size-lg { max-width: 720px; }
.size-xl { max-width: 900px; }

/* Close */
.modal-close {
  position: absolute; top: 14px; right: 14px; z-index: 1;
  width: 32px; height: 32px; border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.08); background: rgba(255,255,255,0.03);
  color: #8890A5; cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all .2s;
}
.modal-close svg { width: 16px; height: 16px; }
.modal-close:hover { color: #F87171; border-color: rgba(248,113,113,.3); background: rgba(248,113,113,.08); }

/* Header */
.modal-header { padding: 28px 28px 0; }
.modal-title { font-size: 18px; font-weight: 700; color: #EDF0F5; margin: 0; letter-spacing: -.01em; }
.modal-desc { font-size: 13px; color: #8890A5; margin: 6px 0 0; }

/* Body */
.modal-body { flex: 1; overflow-y: auto; padding: 24px 28px; }
.modal-body::-webkit-scrollbar { width: 4px; }
.modal-body::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.08); border-radius: 2px; }

/* Footer */
.modal-footer {
  display: flex; justify-content: flex-end; gap: 10px;
  padding: 18px 28px; border-top: 1px solid rgba(255,255,255,0.06);
  flex-shrink: 0;
}

/* Buttons */
.btn-cancel {
  padding: 10px 22px; border-radius: 10px;
  border: 1px solid rgba(255,255,255,0.1); background: transparent;
  color: #8890A5; font-size: 14px; font-weight: 500; cursor: pointer;
  transition: all .15s;
}
.btn-cancel:hover:not(:disabled) { border-color: rgba(255,255,255,0.2); color: #EDF0F5; }
.btn-cancel:disabled { opacity: .4; cursor: not-allowed; }

.btn-confirm {
  padding: 10px 26px; border-radius: 10px; border: none;
  background: #00F5FF; color: #0A0F1C;
  font-size: 14px; font-weight: 600; cursor: pointer;
  transition: all .2s; display: inline-flex; align-items: center; gap: 6px;
  min-width: 72px; justify-content: center;
}
.btn-confirm:hover:not(:disabled) {
  box-shadow: 0 0 24px rgba(0, 245, 255, 0.35);
  transform: translateY(-1px);
}
.btn-confirm:disabled { opacity: .4; cursor: not-allowed; }

/* Spinner */
.spinner {
  width: 16px; height: 16px; border: 2px solid rgba(10,15,28,.3);
  border-top-color: #0A0F1C; border-radius: 50%;
  animation: spin .6s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Transition */
.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity .2s ease; }
.modal-fade-enter-active .modal-card, .modal-fade-leave-active .modal-card { transition: transform .2s ease; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-from .modal-card { transform: scale(.96) translateY(8px); }
.modal-fade-leave-to .modal-card { transform: scale(.98); }
</style>
