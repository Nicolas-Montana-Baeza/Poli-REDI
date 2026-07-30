<script setup>
import { computed } from 'vue'
import { formatBusinessDateTime } from '@/utils/businessTime'
import StatusBadge from './StatusBadge.vue'
import PrimaryButton from './PrimaryButton.vue'

const props = defineProps({
  progress: { type: Object, required: true },
  busy: { type: Boolean, default: false },
  announce: { type: Boolean, default: true },
  showStatus: { type: Boolean, default: true }
})
defineEmits(['confirm', 'withdraw'])
const count = computed(() => Math.max(0, Number(props.progress.participantCount) || 0))
const target = computed(() => Math.max(0, Number(props.progress.targetParticipants) || 0))
const capacity = computed(() => Math.max(0, Number(props.progress.capacity) || 0))
const capacityPercent = value => capacity.value > 0 ? Math.min(100, Math.max(0, value / capacity.value * 100)) : 0
const fillPercent = computed(() => capacityPercent(count.value))
const capacityProgressPercent = computed(() => Math.round(fillPercent.value))
const minimum = computed(() => Math.max(0, Number(props.progress.minimumParticipants) || 0))
const reachedMinimum = computed(() => count.value >= minimum.value)
const markers = computed(() => {
  const minimumMarker = { key: 'minimum', kind: 'minimum', label: 'Mínimo', value: minimum.value, percent: capacityPercent(minimum.value) }
  const targetMarker = { key: 'target', kind: 'target', label: 'Objetivo', value: target.value, percent: capacityPercent(target.value) }
  if (minimumMarker.percent === targetMarker.percent) {
    return [{ ...minimumMarker, kind: 'combined', label: 'Mínimo y objetivo' }]
  }
  return [minimumMarker, targetMarker]
})
const semanticProgressLabel = computed(() => {
  const participant = count.value === 1 ? 'participante confirmado' : 'participantes confirmados'
  return `${count.value} ${participant}; mínimo ${minimum.value}; objetivo ${target.value}; capacidad ${capacity.value}`
})
const status = computed(() => String(props.progress.status || '').toUpperCase())
const isTerminal = computed(() => ['CANCELLED', 'EXPIRED'].includes(status.value))
const active = computed(() => ['PENDING', 'CONFIRMED'].includes(status.value))
const deadlineOpen = computed(() => {
  const deadline = new Date(props.progress.confirmationDeadline).getTime()
  return Number.isFinite(deadline) && Date.now() <= deadline
})
const canConfirm = computed(() => active.value && deadlineOpen.value && !props.progress.isMember)
const canWithdraw = computed(() => active.value && deadlineOpen.value && props.progress.isMember && !props.progress.isOwner)
const statusText = { PENDING: 'Pendiente', CONFIRMED: 'Confirmada', CANCELLED: 'Cancelada', EXPIRED: 'Vencida', REJECTED: 'Rechazada' }
const summary = computed(() => `${count.value} de ${target.value} participantes confirmados`)
const deadline = computed(() => formatBusinessDateTime(props.progress.confirmationDeadline))
const terminalMessage = computed(() => status.value === 'CANCELLED'
  ? 'Esta reserva fue cancelada.'
  : 'El plazo para reunir participantes terminó y la reserva no fue confirmada.')
const accessibleLabel = computed(() => isTerminal.value
  ? `Estado de la invitación: ${statusText[status.value]}. ${terminalMessage.value}`
  : `Progreso de participantes: ${summary.value}. Estado ${statusText[status.value] || status.value || 'desconocido'}`)
const closedMessage = computed(() => {
  if (status.value === 'EXPIRED') return 'La invitación venció y ya no admite confirmaciones.'
  if (status.value === 'CANCELLED') return 'La solicitud fue cancelada y ya no admite cambios.'
  if (!deadlineOpen.value) return 'El plazo de confirmación ya venció.'
  if (props.progress.isOwner) return 'Creaste esta reserva: no puedes retirar tu propia participación.'
  return ''
})
</script>
<template>
  <section class="participants-progress" :class="{ terminal: isTerminal }" :role="announce ? 'status' : 'region'" :aria-live="announce ? 'polite' : 'off'" :aria-label="accessibleLabel" tabindex="0">
    <template v-if="isTerminal">
      <div class="heading"><h2>Estado de la invitación</h2><StatusBadge v-if="showStatus" :status="status" /></div>
      <p class="terminal-message">{{ terminalMessage }}</p>
    </template>
    <template v-else>
      <div class="heading"><h2>Participantes</h2><StatusBadge v-if="showStatus" :status="status" /></div>
      <p class="count">{{ summary }}</p>
      <progress class="sr-only" :value="count" :max="capacity || 1" :aria-label="semanticProgressLabel">{{ capacityProgressPercent }}%</progress>
      <div class="visual-progress" aria-hidden="true">
        <div class="visual-track">
          <span class="visual-fill" :class="{ reached: reachedMinimum }" :style="{ width: `${fillPercent}%` }" />
          <span v-for="marker in markers" :key="marker.key" class="visual-marker" :class="`marker-${marker.kind}`" :style="{ left: `${marker.percent}%` }" />
        </div>
      </div>
      <div class="progress-legend" aria-hidden="true">
        <span>Mínimo {{ minimum }}</span>
        <span>Capacidad {{ capacity }}</span>
      </div>
      <p>La reserva se confirmará automáticamente al llegar a {{ minimum }} participantes.</p>
      <p class="deadline">Disponible hasta el {{ deadline }}</p>
      <div class="actions">
        <PrimaryButton v-if="canConfirm" :loading="busy" aria-pressed="false" @click="$emit('confirm')">Confirmar participación</PrimaryButton>
        <PrimaryButton v-else-if="canWithdraw" variant="danger" :loading="busy" aria-pressed="true" @click="$emit('withdraw')">Retirar participación</PrimaryButton>
        <p v-if="closedMessage" class="closed-message">{{ closedMessage }}</p>
      </div>
    </template>
  </section>
</template>
<style scoped>
.participants-progress{display:grid;gap:1rem;padding:1.25rem;border:1px solid var(--color-border);border-radius:var(--radius-xl);background:var(--color-surface)}.participants-progress.terminal{gap:var(--space-3)}.heading{display:flex;align-items:center;justify-content:space-between;gap:.75rem;flex-wrap:wrap}.heading h2{margin:0}.count{font-size:1rem;font-weight:800}.sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}.visual-progress{padding:var(--space-2) 0}.visual-track{position:relative;height:12px;overflow:visible;border:1px solid var(--color-border);border-radius:var(--radius-pill);background:var(--color-surface-muted)}.visual-fill{display:block;height:100%;border-radius:var(--radius-pill);background:var(--color-primary);transition:width 240ms ease}.visual-fill.reached{background:var(--color-success)}.visual-marker{position:absolute;top:-4px;width:2px;height:18px;transform:translateX(-1px);border-radius:var(--radius-pill)}.marker-minimum{background:var(--color-warning)}.marker-target{background:var(--color-primary-strong)}.marker-combined{background:var(--color-primary-strong);box-shadow:0 0 0 1px var(--color-warning)}.progress-legend{display:flex;flex-wrap:wrap;gap:var(--space-2) var(--space-4);color:var(--color-text-muted);font-size:var(--text-help);font-weight:650}dl{display:grid;grid-template-columns:repeat(auto-fit,minmax(9rem,1fr));gap:.75rem}dl div{padding:.75rem;border-radius:.75rem;background:var(--color-surface-muted)}dt{color:var(--color-text-muted);font-size:.85rem}dd{margin:.25rem 0 0;font-weight:700}.actions{display:flex;gap:.75rem;flex-wrap:wrap}.closed-message,.terminal-message{margin:0;color:var(--color-text-muted);font-weight:650;line-height:1.5}@media(max-width:520px){.participants-progress{padding:1rem}.actions :deep(button){width:100%}}@media(prefers-reduced-motion:reduce){.visual-fill{transition:none}}
</style>
