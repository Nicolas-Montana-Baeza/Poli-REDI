<script setup>
import { computed } from 'vue'
import { formatBusinessDateTime } from '@/utils/businessTime'
const props = defineProps({ progress: { type: Object, required: true }, busy: { type: Boolean, default: false } })
defineEmits(['confirm', 'withdraw'])
const percent = computed(() => Math.min(100, Math.round((props.progress.participantCount / props.progress.targetParticipants) * 100)))
const active = computed(() => ['PENDING', 'CONFIRMED'].includes(props.progress.status))
const deadlineOpen = computed(() => {
  const deadline = new Date(props.progress.confirmationDeadline).getTime()
  return Number.isFinite(deadline) && Date.now() <= deadline
})
const canConfirm = computed(() => active.value && deadlineOpen.value && !props.progress.isMember)
const canWithdraw = computed(() => active.value && deadlineOpen.value && props.progress.isMember && !props.progress.isOwner)
const closedMessage = computed(() => {
  if (props.progress.status === 'CANCELLED') return 'La solicitud fue cancelada y ya no admite cambios.'
  if (!deadlineOpen.value) return 'El plazo de confirmación ya venció.'
  if (props.progress.isOwner) return 'El solicitante no puede retirarse.'
  return ''
})
</script>
<template>
  <section class="participants-progress" aria-live="polite">
    <h2>Progreso de participantes</h2>
    <p class="count">{{ progress.participantCount }} de {{ progress.targetParticipants }}</p>
    <progress :value="progress.participantCount" :max="progress.targetParticipants">{{ percent }}%</progress>
    <dl>
      <div><dt>Mínimo requerido</dt><dd>{{ progress.minimumParticipants }}</dd></div>
      <div><dt>Objetivo</dt><dd>{{ progress.targetParticipants }}</dd></div>
      <div><dt>Capacidad</dt><dd>{{ progress.capacity }}</dd></div>
      <div><dt>Estado</dt><dd>{{ progress.status }}</dd></div>
      <div><dt>Plazo</dt><dd>{{ formatBusinessDateTime(progress.confirmationDeadline) }}</dd></div>
    </dl>
    <p>La reserva se confirma al alcanzar el mínimo requerido.</p>
    <div class="actions">
      <button v-if="canConfirm" type="button" :disabled="busy" @click="$emit('confirm')">Confirmar participación</button>
      <button v-else-if="canWithdraw" type="button" :disabled="busy" @click="$emit('withdraw')">Retirar participación</button>
      <p v-if="closedMessage">{{ closedMessage }}</p>
    </div>
  </section>
</template>
<style scoped>
.participants-progress{display:grid;gap:1rem;padding:1.25rem;border:1px solid #d8e1f0;border-radius:1rem;background:white}.count{font-size:1.5rem;font-weight:800}progress{width:100%;height:1rem}dl{display:grid;grid-template-columns:repeat(auto-fit,minmax(9rem,1fr));gap:.75rem}dl div{padding:.75rem;border-radius:.75rem;background:#f4f7fc}dt{color:#53657d;font-size:.85rem}dd{margin:.25rem 0 0;font-weight:700}.actions{display:flex;gap:.75rem;flex-wrap:wrap}button{min-height:44px;padding:.65rem 1rem;border:0;border-radius:.65rem;background:#2563eb;color:white;font-weight:700}
</style>
