<script setup>
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import PrimaryButton from './PrimaryButton.vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  title: { type: String, default: 'Confirmar acción' },
  message: { type: String, default: 'Revisa esta acción antes de continuar.' },
  confirmText: { type: String, default: 'Confirmar' },
  cancelText: { type: String, default: 'Volver' },
  variant: { type: String, default: 'danger' },
  loading: { type: Boolean, default: false },
  destructive: { type: Boolean, default: false }
})
const emit = defineEmits(['confirm', 'cancel'])
const dialog = ref(null)
const cancelButton = ref(null)
let previousFocus = null
let previousOverflow = ''
const titleId = `confirm-title-${Math.random().toString(36).slice(2)}`
const descriptionId = `confirm-description-${Math.random().toString(36).slice(2)}`

const close = () => { if (!props.loading) emit('cancel') }
const focusables = () => [...(dialog.value?.querySelectorAll('button:not(:disabled),[href],[tabindex]:not([tabindex="-1"])') || [])]
const onKeydown = event => {
  if (event.key === 'Escape') { event.preventDefault(); close(); return }
  if (event.key !== 'Tab') return
  const items = focusables()
  if (!items.length) { event.preventDefault(); dialog.value?.focus(); return }
  const first = items[0], last = items[items.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}

watch(() => props.show, async show => {
  if (show) {
    previousFocus = document.activeElement
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    await nextTick()
    ;(props.destructive ? cancelButton.value?.$el : dialog.value)?.focus()
  } else {
    document.body.style.overflow = previousOverflow
    previousFocus?.focus?.()
  }
}, { immediate: true })
onBeforeUnmount(() => {
  document.body.style.overflow = previousOverflow
  if (props.show) previousFocus?.focus?.()
})
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click.self="close">
      <div ref="dialog" class="modal-card" role="dialog" aria-modal="true" :aria-labelledby="titleId" :aria-describedby="descriptionId" tabindex="-1" @keydown="onKeydown">
        <h2 :id="titleId" class="modal-title">{{ title }}</h2>
        <p :id="descriptionId" class="modal-message">{{ message }}</p>
        <div class="modal-actions">
          <PrimaryButton ref="cancelButton" variant="secondary" :disabled="loading" @click="close">{{ cancelText }}</PrimaryButton>
          <PrimaryButton :variant="variant" :loading="loading" @click="$emit('confirm')">{{ confirmText }}</PrimaryButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop{position:fixed;inset:0;z-index:9999;display:flex;align-items:center;justify-content:center;padding:16px;background:rgba(15,23,42,.45);backdrop-filter:blur(4px)}
.modal-card{width:100%;max-width:440px;padding:24px;border:1px solid var(--color-border-soft,#e7edf5);border-radius:var(--radius-xl,12px);background:var(--color-surface,#fff);box-shadow:var(--shadow-modal);animation:modalPop .2s cubic-bezier(.16,1,.3,1)}
.modal-title{margin:0 0 8px;font-size:18px;color:var(--color-text,#162033)}
.modal-message{margin:0 0 24px;color:var(--color-text-muted,#64748b);font-size:14px;line-height:1.5}
.modal-actions{display:flex;justify-content:flex-end;gap:12px}
@keyframes modalPop{from{opacity:0;transform:scale(.95)}to{opacity:1;transform:scale(1)}}
@media(max-width:520px){.modal-backdrop{align-items:flex-end;padding:0}.modal-card{max-width:none;padding:20px 16px calc(20px + env(safe-area-inset-bottom));border-radius:16px 16px 0 0}.modal-actions{flex-direction:column-reverse}.modal-actions :deep(button){width:100%;min-height:44px}}
@media(prefers-reduced-motion:reduce){.modal-card{animation:none}}
</style>
