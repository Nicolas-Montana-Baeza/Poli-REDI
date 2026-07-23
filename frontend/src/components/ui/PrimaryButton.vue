<script setup>
defineProps({
  variant: {
    type: String,
    default: 'primary' // 'primary', 'secondary', 'danger', 'ghost'
  },
  size: {
    type: String,
    default: 'md' // 'sm', 'md', 'lg'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'button'
  }
})

const emit = defineEmits(['click'])

function handleClick(event) {
  emit('click', event)
}
</script>

<template>
  <button
    :type="type"
    class="btn-ui"
    :class="[`btn-${variant}`, `btn-${size}`, { 'is-loading': loading }]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <span v-if="loading" class="spinner" />
    <slot />
  </button>
</template>

<style scoped>
.btn-ui {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: var(--radius-md, 8px);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease-in-out;
  border: 1px solid transparent;
  outline: none;
}

.btn-ui:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Tamaños */
.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}
.btn-md {
  padding: 10px 18px;
  font-size: 14px;
}
.btn-lg {
  padding: 14px 24px;
  font-size: 16px;
}

/* Variantes */
.btn-primary {
  background-color: var(--color-primary, #2563eb);
  color: #ffffff;

  &:hover:not(:disabled) {
    background-color: var(--color-primary-strong, #1d4ed8);
    box-shadow: var(--shadow-hover);
  }
}

.btn-secondary {
  background-color: var(--color-surface-soft, #edf3fb);
  color: var(--color-text, #162033);
  border-color: var(--color-border, #d8e0ec);

  &:hover:not(:disabled) {
    background-color: var(--color-border-soft, #e7edf5);
  }
}

.btn-danger {
  background-color: var(--color-error-strong, #dc2626);
  color: #ffffff;

  &:hover:not(:disabled) {
    background-color: var(--color-error, #b91c1c);
    box-shadow: var(--shadow-hover);
  }
}

.btn-ghost {
  background-color: transparent;
  color: var(--color-text-muted, #64748b);

  &:hover:not(:disabled) {
    background-color: var(--color-surface-soft, #edf3fb);
    color: var(--color-text, #162033);
  }
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
