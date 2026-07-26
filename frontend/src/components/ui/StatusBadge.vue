<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, default: '' },
  label: { type: String, default: '' }
})

const statuses = {
  PENDING: ['Pendiente', 'pending'],
  CONFIRMED: ['Confirmada', 'confirmed'],
  CANCELLED: ['Cancelada', 'cancelled'],
  EXPIRED: ['Vencida', 'expired'],
  REJECTED: ['Rechazada', 'rejected'],
  ACTIVE: ['Activo', 'active'],
  INACTIVE: ['Inactivo', 'inactive']
}

const config = computed(() => {
  const value = String(props.status || '').trim()
  const known = statuses[value.toUpperCase()]
  return known
    ? { text: props.label || known[0], variant: known[1] }
    : { text: props.label || value || 'Estado desconocido', variant: 'neutral' }
})
</script>

<template>
  <span class="status-badge" :class="`status-${config.variant}`">
    <span class="status-dot" aria-hidden="true" />
    {{ config.text }}
  </span>
</template>

<style scoped>
.status-badge{display:inline-flex;align-items:center;gap:6px;padding:4px 10px;border:1px solid transparent;border-radius:var(--radius-pill,999px);font-size:12px;font-weight:700;white-space:nowrap}
.status-dot{width:6px;height:6px;border-radius:50%;background:currentColor}
.status-confirmed,.status-active{background:var(--color-success-soft,#dcfce7);color:var(--color-success,#166534);border-color:var(--color-success-border,#bbf7d0)}
.status-pending{background:var(--color-warning-soft,#fef3c7);color:var(--color-warning,#92400e);border-color:#fde68a}
.status-cancelled,.status-expired,.status-rejected{background:var(--color-error-soft,#fee2e2);color:var(--color-error,#b91c1c);border-color:var(--color-error-border,#fecaca)}
.status-inactive,.status-neutral{background:var(--color-surface-muted,#f1f5f9);color:var(--color-text-muted,#64748b);border-color:var(--color-border,#d8e0ec)}
</style>
