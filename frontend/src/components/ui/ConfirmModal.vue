<script setup>
import PrimaryButton from './PrimaryButton.vue'

defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Confirmar Acción'
  },
  message: {
    type: String,
    default: '¿Estás seguro de que deseas realizar esta acción?'
  },
  confirmText: {
    type: String,
    default: 'Confirmar'
  },
  cancelText: {
    type: String,
    default: 'Cancelar'
  },
  variant: {
    type: String,
    default: 'danger' // 'danger', 'primary'
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['confirm', 'cancel'])

function onConfirm() {
  emit('confirm')
}

function onCancel() {
  emit('cancel')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click.self="onCancel">
      <div class="modal-card">
        <h3 class="modal-title">{{ title }}</h3>
        <p class="modal-message">{{ message }}</p>

        <div class="modal-actions">
          <PrimaryButton variant="secondary" :disabled="loading" @click="onCancel">
            {{ cancelText }}
          </PrimaryButton>
          <PrimaryButton :variant="variant" :loading="loading" @click="onConfirm">
            {{ confirmText }}
          </PrimaryButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background-color: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal-card {
  background-color: var(--color-surface, #ffffff);
  border-radius: var(--radius-xl, 12px);
  box-shadow: var(--shadow-modal);
  padding: 24px;
  width: 100%;
  max-width: 440px;
  border: 1px solid var(--color-border-soft, #e7edf5);

  animation: modalPop 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text, #162033);
  margin-bottom: 8px;
}

.modal-message {
  font-size: 14px;
  color: var(--color-text-muted, #64748b);
  line-height: 1.5;
  margin-bottom: 24px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@keyframes modalPop {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
