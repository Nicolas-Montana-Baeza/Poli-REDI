<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: true
  },
  label: {
    type: String,
    default: ''
  }
})

const statusConfig = computed(() => {
  const norm = (props.status || '').toUpperCase()
  switch (norm) {
    case 'CONFIRMED':
    case 'CONFIRMADA':
    case 'APPROVED':
      return { text: props.label || 'Confirmada', variant: 'status-confirmed' }
    case 'PENDING':
    case 'PENDIENTE':
      return { text: props.label || 'Pendiente Quorum', variant: 'status-pending' }
    case 'PRIORITY':
    case 'INSTITUTIONAL':
    case 'EFI':
      return { text: props.label || 'Institucional', variant: 'status-priority' }
    case 'WORKSHOP':
    case 'TALLER':
      return { text: props.label || 'Taller', variant: 'status-workshop' }
    case 'CANCELLED':
    case 'CANCELADA':
    case 'REJECTED':
    default:
      return { text: props.label || 'Cancelada', variant: 'status-cancelled' }
  }
})
</script>

<template>
  <span class="status-badge" :class="statusConfig.variant">
    <span class="status-dot" />
    {{ statusConfig.text }}
  </span>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--radius-pill, 999px);
  font-size: 12px;
  font-weight: 700;
  border: 1px solid transparent;
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
}

.status-confirmed {
  background-color: var(--color-status-confirmed-bg);
  color: var(--color-status-confirmed-text);
  border-color: var(--color-status-confirmed-border);
}

.status-pending {
  background-color: var(--color-status-pending-bg);
  color: var(--color-status-pending-text);
  border-color: var(--color-status-pending-border);
}

.status-priority {
  background-color: var(--color-status-priority-bg);
  color: var(--color-status-priority-text);
  border-color: var(--color-status-priority-border);
}

.status-workshop {
  background-color: var(--color-status-workshop-bg);
  color: var(--color-status-workshop-text);
  border-color: var(--color-status-workshop-border);
}

.status-cancelled {
  background-color: var(--color-status-cancelled-bg);
  color: var(--color-status-cancelled-text);
  border-color: var(--color-status-cancelled-border);
}
</style>
