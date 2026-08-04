<script setup>
import { computed, ref, watch } from 'vue'
import {
  CalendarDays,
  Clock,
  Info,
  LockKeyhole,
  MapPin,
  Minus,
  Plus,
  Timer,
  X
} from 'lucide-vue-next'
import { reservationsService } from '@/services/reservations.service'
import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import {
  getAvailabilityDisplayTitle,
  getAvailabilityType
} from '@/utils/availabilityType'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus,
  isReservationCancelable
} from '@/utils/reservationTime'
import { formatBusinessDateTime } from '@/utils/businessTime'
import { useAccessibleDialog } from '@/composables/useAccessibleDialog'

const props = defineProps({
  visible: { type: Boolean, default: false },
  reservation: { type: Object, default: null },
  errorMessage: { type: String, default: '' },
  canCancel: { type: Boolean, default: true },
  canEditTarget: { type: Boolean, default: false },
  canManageJoinCode: { type: Boolean, default: false },
  participationMode: { type: Boolean, default: false },
  participationBusy: { type: Boolean, default: false },
  participationMessage: { type: String, default: '' },
  readOnly: { type: Boolean, default: false },
  workshopEnrollmentMode: { type: Boolean, default: false },
  canEnrollWorkshop: { type: Boolean, default: false },
  workshopEnrolling: { type: Boolean, default: false },
  canWithdrawWorkshop: { type: Boolean, default: false },
  workshopWithdrawing: { type: Boolean, default: false },
  workshopActionMessage: { type: String, default: '' }
})
const emit = defineEmits(['close', 'cancel', 'update-target', 'confirm-participation', 'withdraw-participation', 'enroll-workshop', 'withdraw-workshop'])
const closeDialog = () => emit('close')
const { dialogRef, onKeydown } = useAccessibleDialog({
  visible: computed(() => props.visible),
  close: closeDialog,
  focusOnOpen: computed(() => !props.participationMode)
})

const targetValue = ref(null)
const joinCode = ref('')
const codeError = ref('')
const codeBusy = ref(false)
const codeOpen = ref(false)
const copyFeedback = ref('')

watch(() => props.reservation, value => {
  targetValue.value = value?.targetParticipants ?? null
  joinCode.value = ''
  codeOpen.value = false
  codeError.value = ''
  copyFeedback.value = ''
}, { immediate: true })

const availabilityType = computed(() =>
  getAvailabilityType(props.reservation)
)
const isInstitutional = computed(() =>
  ['workshop', 'institutional'].includes(availabilityType.value.family)
)
const isWorkshop = computed(() =>
  availabilityType.value.key === 'workshop'
)
const date = computed(() => formatReservationDate(props.reservation?.startTime))
const time = computed(() => formatReservationTimeRange(props.reservation?.startTime, props.reservation?.durationMinutes || 60))
const title = computed(() => getAvailabilityDisplayTitle(
  props.reservation,
  isInstitutional.value ? 'Actividad programada' : 'Reserva'
))
const status = computed(() =>
  getReservationDisplayStatus(props.reservation)
)
const showInstitutionalStatus = computed(() => {
  if (!isInstitutional.value || !props.workshopEnrollmentMode) return false

  return ['CANCELLED', 'REJECTED', 'EXPIRED', 'COMPLETED'].includes(
    String(props.reservation?.status || '').toUpperCase()
  )
})
const institutionalDescription = computed(() => {
  if (props.workshopEnrollmentMode) {
    return 'Información de tu inscripción al taller.'
  }

  return isWorkshop.value
    ? 'Información del taller seleccionado.'
    : 'Información de la actividad seleccionada.'
})
const institutionalCallout = computed(() =>
  isWorkshop.value
    ? 'Este horario está reservado para el taller y no admite reservas particulares.'
    : 'Este horario está reservado para la actividad y no admite reservas particulares.'
)
const participantCount = computed(() => Number(props.reservation?.participantCount) || 0)
const target = computed(() => Number(props.reservation?.targetParticipants) || 0)
const minimum = computed(() => Number(props.reservation?.minimumParticipants) || 0)
const capacity = computed(() => Number(props.reservation?.capacity) || target.value || 1)
const progress = computed(() => Math.min(100, participantCount.value / capacity.value * 100))
const minTarget = computed(() => Math.max(minimum.value, participantCount.value))
const targetEditingAllowed = computed(() =>
  props.canEditTarget && !props.readOnly && !props.participationMode
)
const canDecrease = computed(() => targetEditingAllowed.value && Number(targetValue.value) > minTarget.value)
const canIncrease = computed(() => targetEditingAllowed.value && Number(targetValue.value) < capacity.value)
const showCancelAction = computed(() => !props.readOnly && props.canCancel && !isInstitutional.value && isReservationCancelable(props.reservation))
const showWorkshopEnrollment = computed(() => (
  !props.readOnly &&
  isWorkshop.value &&
  props.canEnrollWorkshop
))
const showWorkshopWithdrawal = computed(() => (
  !props.readOnly &&
  isWorkshop.value &&
  props.canWithdrawWorkshop
))
const showPermissionNote = computed(() => !props.readOnly && !props.participationMode && !isInstitutional.value && !showCancelAction.value)
const enrollmentDate = computed(() => {
  const value = new Date(props.reservation?.enrolledAt)
  if (Number.isNaN(value.getTime())) return 'Fecha no disponible'
  return new Intl.DateTimeFormat('es-CL', { day: '2-digit', month: 'long', year: 'numeric' }).format(value)
})
const participationStatus = computed(() => String(props.reservation?.status || '').toUpperCase())
const participationActive = computed(() => ['PENDING', 'CONFIRMED'].includes(participationStatus.value))
const participationDeadlineOpen = computed(() => {
  const deadline = new Date(props.reservation?.confirmationDeadline).getTime()
  return Number.isFinite(deadline) && Date.now() <= deadline
})
const canConfirmParticipation = computed(() => (
  props.participationMode &&
  participationActive.value &&
  participationDeadlineOpen.value &&
  !props.reservation?.isMember
))
const canWithdrawParticipation = computed(() => (
  props.participationMode &&
  participationActive.value &&
  participationDeadlineOpen.value &&
  props.reservation?.isMember &&
  !props.reservation?.isOwner
))
const participationClosedMessage = computed(() => {
  if (!props.participationMode) return ''
  if (participationStatus.value === 'EXPIRED') return 'La invitación venció y ya no admite confirmaciones.'
  if (participationStatus.value === 'CANCELLED') return 'La solicitud fue cancelada y ya no admite cambios.'
  if (!participationDeadlineOpen.value) return 'El plazo de confirmación ya venció.'
  if (props.reservation?.isOwner) return 'Creaste esta reserva: no puedes retirar tu propia participación.'
  return ''
})

const updateTarget = delta => {
  if (!targetEditingAllowed.value) return
  const current = Number(targetValue.value || target.value)
  targetValue.value = Math.min(capacity.value, Math.max(minTarget.value, current + delta))
}
const saveTarget = () => {
  if (!targetEditingAllowed.value) return
  emit('update-target', targetValue.value)
}
const loadJoinCode = async () => {
  if (!props.canManageJoinCode) return
  codeOpen.value = true
  if (joinCode.value) return
  codeBusy.value = true
  codeError.value = ''
  try {
    joinCode.value = (await reservationsService.getJoinCode(props.reservation.id)).joinCode
  } catch {
    codeError.value = 'No se pudo recuperar el código.'
  } finally {
    codeBusy.value = false
  }
}
const rotateJoinCode = async () => {
  codeBusy.value = true
  codeError.value = ''
  try {
    joinCode.value = (await reservationsService.rotateJoinCode(props.reservation.id)).joinCode
  } catch {
    codeError.value = 'No se pudo generar un código nuevo.'
  } finally {
    codeBusy.value = false
  }
}
const copyJoinCode = async () => {
  try {
    await navigator.clipboard.writeText(joinCode.value)
    copyFeedback.value = 'Código copiado al portapapeles.'
  } catch {
    copyFeedback.value = 'No se pudo copiar. Selecciona el código manualmente.'
  }
}
const shareJoinCode = async () => {
  const url = `${window.location.origin}/join/${encodeURIComponent(joinCode.value)}`
  if (navigator.share) {
    try {
      await navigator.share({ title: 'Invitación a reserva grupal', url })
      copyFeedback.value = 'Invitación compartida.'
      return
    } catch (error) {
      if (error?.name === 'AbortError') return
    }
  }
  try {
    await navigator.clipboard.writeText(url)
    copyFeedback.value = 'Enlace de invitación copiado.'
  } catch {
    copyFeedback.value = 'No se pudo compartir la invitación.'
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="detail-overlay" @click.self="closeDialog">
      <article ref="dialogRef" class="detail-modal" role="dialog" aria-modal="true" aria-labelledby="reservation-detail-title" tabindex="-1" @keydown="onKeydown">
        <header class="detail-header">
          <div class="header-copy">
            <div v-if="isInstitutional" class="detail-title-row">
              <AvailabilityTypeChip
                :meta="availabilityType"
                aria-hidden
              />
              <h2 id="reservation-detail-title">{{ title }}</h2>
            </div>
            <h2 v-else id="reservation-detail-title">{{ participationMode ? 'Detalle de la invitación' : 'Detalle de reserva' }}</h2>
            <p>{{ participationMode ? 'Revisa el estado y confirma tu participación.' : isInstitutional ? institutionalDescription : 'Información de la reserva seleccionada.' }}</p>
          </div>
          <button class="close-button" type="button" aria-label="Cerrar" @click="closeDialog">
            <X class="icon close-icon" :size="22" aria-hidden="true" />
          </button>
        </header>

        <div class="detail-body">
          <template v-if="reservation">
            <section v-if="!participationMode" class="summary-card">
              <div v-if="isInstitutional" class="institutional-resource">
                <MapPin class="icon note-icon" :size="20" aria-hidden="true" />
                <span>
                  <small>Instalación</small>
                  <strong>{{ reservation.resourceName || 'Instalación' }}</strong>
                </span>
                <span v-if="showInstitutionalStatus" class="status-pill" :class="status.className">{{ status.label }}</span>
              </div>
              <div v-else class="summary-heading">
                <div class="summary-copy">
                  <span class="field-label">Actividad</span>
                  <h3>{{ title }}</h3>
                  <p>{{ reservation.resourceName || 'Instalación' }}</p>
                </div>
                <span class="status-pill" :class="status.className">{{ status.label }}</span>
              </div>

              <div v-if="workshopEnrollmentMode" class="facts">
                <div class="fact">
                  <CalendarDays class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Horario del taller</small><strong>{{ reservation.dayText }} · {{ reservation.scheduleText }}</strong></span>
                </div>
                <div class="fact">
                  <Clock class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Fecha de inscripción</small><strong>{{ enrollmentDate }}</strong></span>
                </div>
                <div v-if="reservation.instructorName" class="fact">
                  <Info class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Instructor</small><strong>{{ reservation.instructorName }}</strong></span>
                </div>
              </div>
              <div v-else class="facts">
                <div class="fact">
                  <CalendarDays class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Fecha</small><strong>{{ date }}</strong></span>
                </div>
                <div class="fact">
                  <Clock class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Horario</small><strong>{{ time }}</strong></span>
                </div>
                <div class="fact">
                  <Timer class="icon fact-icon" :size="20" aria-hidden="true" />
                  <span><small>Duración</small><strong>{{ reservation.durationMinutes }} minutos</strong></span>
                </div>
              </div>

            </section>

            <div v-if="isInstitutional && !workshopEnrollmentMode" class="info-callout">
              <Info class="icon callout-icon" :size="20" aria-hidden="true" />
              <span>{{ institutionalCallout }}</span>
            </div>

            <p
              v-if="isWorkshop && workshopActionMessage"
              class="workshop-action-feedback"
              role="status"
              aria-live="polite"
            >
              {{ workshopActionMessage }}
            </p>

            <section v-if="target && !isInstitutional" class="participants-card">
              <div>
                <span class="field-label">Participantes</span>
                <h3>{{ participantCount }} de {{ target }} confirmados</h3>
              </div>
              <div class="progress-track" aria-hidden="true"><span :style="{ width: `${progress}%` }" /></div>
              <progress class="sr-only" :value="participantCount" :max="capacity">{{ participantCount }} de {{ capacity }}</progress>
              <div class="metrics">
                <span>Mínimo <strong>{{ minimum }}</strong></span>
                <span>Capacidad <strong>{{ capacity }}</strong></span>
              </div>
              <p>La reserva se confirmará automáticamente al llegar a {{ minimum }} participantes.</p>
              <p v-if="participationMode && reservation.confirmationDeadline">
                Disponible hasta el {{ formatBusinessDateTime(reservation.confirmationDeadline) }}.
              </p>

              <div v-if="targetEditingAllowed" class="target-editor">
                <div class="target-copy">
                  <strong>Objetivo de participantes</strong>
                  <small v-if="reservation.confirmationDeadline">Puedes editarlo hasta {{ formatBusinessDateTime(reservation.confirmationDeadline) }}.</small>
                </div>
                <div class="target-controls">
                  <div class="stepper" aria-label="Objetivo de participantes">
                    <button type="button" aria-label="Disminuir objetivo" :disabled="!canDecrease" @click="updateTarget(-1)">
                      <Minus class="icon stepper-icon" :size="18" aria-hidden="true" />
                    </button>
                    <output>{{ targetValue }}</output>
                    <button type="button" aria-label="Aumentar objetivo" :disabled="!canIncrease" @click="updateTarget(1)">
                      <Plus class="icon stepper-icon" :size="18" aria-hidden="true" />
                    </button>
                  </div>
                  <button class="app-button primary save-target" type="button" @click="saveTarget">Guardar cambios</button>
                </div>
              </div>
              <div v-if="participationMode" class="participation-actions">
                <button
                  v-if="canConfirmParticipation"
                  class="app-button primary"
                  type="button"
                  :disabled="participationBusy"
                  @click="emit('confirm-participation')"
                >
                  {{ participationBusy ? 'Confirmando…' : 'Confirmar participación' }}
                </button>
                <button
                  v-else-if="canWithdrawParticipation"
                  class="app-button danger"
                  type="button"
                  :disabled="participationBusy"
                  @click="emit('withdraw-participation')"
                >
                  {{ participationBusy ? 'Retirando…' : 'Retirar participación' }}
                </button>
                <p v-if="participationClosedMessage" class="participation-closed">{{ participationClosedMessage }}</p>
                <p v-if="participationMessage" class="participation-feedback" role="status">{{ participationMessage }}</p>
              </div>
            </section>

            <section v-if="!readOnly && canManageJoinCode" class="join-code">
              <div class="join-code-heading">
                <div>
                  <strong>Invita a tu grupo</strong>
                  <p>Consulta el código solo cuando necesites compartir esta reserva.</p>
                </div>
                <button
                  v-if="!codeOpen"
                  type="button"
                  class="app-button secondary"
                  :disabled="codeBusy"
                  @click="loadJoinCode"
                >
                  Consultar código
                </button>
              </div>
              <div v-if="codeOpen" class="join-code-content">
                <p v-if="codeBusy">Consultando código…</p>
                <p v-if="codeError" role="alert">{{ codeError }}</p>
                <template v-if="joinCode">
                  <output aria-label="Código de invitación">{{ joinCode }}</output>
                  <div class="code-actions">
                    <button type="button" class="app-button secondary" @click="copyJoinCode">Copiar código</button>
                    <button type="button" class="app-button secondary" @click="shareJoinCode">Compartir invitación</button>
                    <button type="button" class="app-button secondary" :disabled="codeBusy" @click="rotateJoinCode">Generar código nuevo</button>
                  </div>
                  <p v-if="copyFeedback" class="copy-feedback" role="status">{{ copyFeedback }}</p>
                </template>
              </div>
            </section>
          </template>

          <div v-if="showPermissionNote" class="permission-note">
            <LockKeyhole class="icon callout-icon" :size="20" aria-hidden="true" />
            <span>Solo quien creó la reserva o un administrador puede cancelarla.</span>
          </div>
          <div v-if="errorMessage" class="error" role="alert">{{ errorMessage }}</div>
        </div>

        <footer class="detail-actions">
          <button class="app-button secondary" type="button" @click="closeDialog">{{ isInstitutional ? 'Entendido' : 'Cerrar' }}</button>
          <button
            v-if="showWorkshopEnrollment"
            class="app-button primary"
            type="button"
            :disabled="workshopEnrolling"
            @click="emit('enroll-workshop')"
          >
            {{ workshopEnrolling ? 'Inscribiendo…' : 'Inscribirme' }}
          </button>
          <button
            v-if="showWorkshopWithdrawal"
            class="app-button danger"
            type="button"
            :disabled="workshopWithdrawing"
            @click="emit('withdraw-workshop')"
          >
            {{ workshopWithdrawing ? 'Desinscribiendo…' : 'Desinscribirme' }}
          </button>
          <button v-if="showCancelAction" class="app-button danger" type="button" @click="emit('cancel')">Cancelar reserva</button>
        </footer>
      </article>
    </div>
  </Teleport>
</template>

<style scoped>
.detail-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(15, 23, 42, .58);
  backdrop-filter: blur(4px);
}
.detail-modal {
  width: min(100%, 600px);
  max-height: min(90vh, 760px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  border: 1px solid var(--color-border, #d8e0ec);
  border-radius: var(--radius-xl, 16px);
  background: var(--color-surface, #fff);
  box-shadow: 0 24px 64px rgba(15, 23, 42, .22);
}
.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex: none;
  padding: 22px 22px 16px;
  border-bottom: 1px solid var(--color-border, #d8e0ec);
}
.header-copy { min-width: 0; }
.detail-header h2 {
  margin: 3px 0 0;
  overflow-wrap: anywhere;
  color: var(--color-text, #14213d);
  font-size: clamp(21px, 4vw, 25px);
  line-height: 1.2;
}
.detail-header p,
.summary-card p,
.participants-card p {
  margin: 5px 0 0;
  color: var(--color-text-muted, #60708a);
  font-size: 14px;
  line-height: 1.5;
}
.field-label {
  color: var(--color-primary, #2563eb);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: .04em;
  text-transform: uppercase;
}
.detail-title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 9px;
}
.detail-title-row h2 { min-width: 0; }
.close-button {
  width: 44px;
  height: 44px;
  flex: none;
  display: grid;
  place-items: center;
  padding: 0;
  border: 0;
  border-radius: 12px;
  background: var(--color-surface-soft, #edf3fb);
  color: var(--color-text, #14213d);
  cursor: pointer;
}
.icon {
  width: 20px !important;
  height: 20px !important;
  min-width: 20px;
  min-height: 20px;
  max-width: 20px;
  max-height: 20px;
  flex: 0 0 20px;
  display: block;
  flex-shrink: 0;
}
.close-button .close-icon {
  width: 22px !important;
  height: 22px !important;
  min-width: 22px;
  min-height: 22px;
  max-width: 22px;
  max-height: 22px;
  flex-basis: 22px;
}
.detail-body {
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  display: grid;
  gap: 16px;
  padding: 18px 22px;
}
.summary-card,
.participants-card,
.join-code {
  min-width: 0;
  display: grid;
  gap: 15px;
  padding: 17px;
  border: 1px solid var(--color-border, #d8e0ec);
  border-radius: var(--radius-lg, 12px);
  background: var(--color-surface, #fff);
}
.summary-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}
.summary-copy { min-width: 0; }
.institutional-resource {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 2px;
}
.institutional-resource > span:not(.status-pill) {
  min-width: 0;
  display: grid;
  flex: 1 1 auto;
  gap: 3px;
}
.institutional-resource small {
  color: var(--color-text-muted, #60708a);
  font-size: 12px;
}
.institutional-resource strong {
  overflow-wrap: anywhere;
  color: var(--color-text, #14213d);
  font-size: 14px;
}
.institutional-resource .note-icon {
  color: var(--color-primary, #2563eb);
}
.summary-card h3,
.participants-card h3 {
  margin: 3px 0 0;
  overflow-wrap: anywhere;
  color: var(--color-text, #14213d);
  font-size: 19px;
  line-height: 1.3;
}
.status-pill {
  flex: none;
  padding: 6px 10px;
  border-radius: 999px;
  background: var(--color-warning-soft, #fff4d6);
  color: var(--color-warning, #9a5500);
  font-size: 12px;
  font-weight: 800;
  white-space: nowrap;
}
.status-pill.confirmed,
.status-pill.completed { background: var(--color-success-soft, #dcfce7); color: var(--color-success, #16803a); }
.status-pill.cancelled,
.status-pill.rejected { background: var(--color-error-soft, #fee2e2); color: var(--color-error, #c62828); }
.facts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.fact {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: 9px;
  padding: 12px;
  border-radius: var(--radius-md, 10px);
  background: var(--color-surface-muted, #f7f9fc);
}
.fact .icon,
.info-callout .icon,
.permission-note .icon {
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.8;
}
.fact > span {
  min-width: 0;
  display: grid;
  gap: 4px;
}
.fact small { color: var(--color-text-muted, #60708a); font-size: 12px; }
.fact strong { overflow-wrap: anywhere; color: var(--color-text, #14213d); font-size: 13px; line-height: 1.35; }
.info-callout,
.permission-note {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 14px 15px;
  border: 1px solid var(--color-border, #d8e0ec);
  border-radius: var(--radius-md, 10px);
  background: var(--color-surface-muted, #f7f9fc);
  color: var(--color-text, #14213d);
  font-size: 14px;
  font-weight: 650;
  line-height: 1.5;
}
.fact > svg.fact-icon {
  width: 20px !important;
  height: 20px !important;
  min-width: 20px;
  min-height: 20px;
  max-width: 20px;
  max-height: 20px;
  flex: 0 0 20px;
  box-sizing: border-box;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: var(--color-primary, #2563eb);
  overflow: visible;
}
.info-callout .icon,
.permission-note .icon { color: var(--color-primary, #2563eb); margin-top: 1px; }
.progress-track {
  height: 12px;
  overflow: hidden;
  border: 1px solid var(--color-border, #d8e0ec);
  border-radius: 999px;
  background: var(--color-surface-muted, #f7f9fc);
}
.progress-track span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--color-primary, #2563eb), #60a5fa);
}
.metrics { display: flex; flex-wrap: wrap; gap: 8px; }
.metrics span {
  padding: 7px 11px;
  border-radius: 999px;
  background: var(--color-surface-muted, #f7f9fc);
  color: var(--color-text-muted, #60708a);
  font-size: 12px;
}
.metrics strong { color: var(--color-text, #14213d); }
.target-editor {
  display: grid;
  gap: 13px;
  padding-top: 15px;
  border-top: 1px solid var(--color-border, #d8e0ec);
}
.target-copy { display: grid; gap: 4px; }
.target-copy small { color: var(--color-text-muted, #60708a); line-height: 1.4; }
.target-controls { display: flex; align-items: center; gap: 12px; }
.stepper {
  width: 150px;
  min-height: 46px;
  display: grid;
  grid-template-columns: 46px 58px 46px;
  overflow: hidden;
  border: 1px solid var(--color-border, #d8e0ec);
  border-radius: var(--radius-md, 10px);
  background: var(--color-surface, #fff);
}
.stepper button {
  min-width: 44px;
  border: 0;
  background: var(--color-surface-muted, #f7f9fc);
  color: var(--color-text, #14213d);
  font-size: 22px;
  cursor: pointer;
  display: grid;
  place-items: center;
}
.stepper .stepper-icon {
  width: 18px !important;
  height: 18px !important;
  min-width: 18px;
  min-height: 18px;
  max-width: 18px;
  max-height: 18px;
  flex-basis: 18px;
}
.stepper button:disabled { opacity: .4; cursor: not-allowed; }
.stepper output {
  display: grid;
  place-items: center;
  border-inline: 1px solid var(--color-border, #d8e0ec);
  font-size: 16px;
  font-weight: 800;
}
.save-target { min-height: 46px; white-space: nowrap; }
.join-code-content { min-width: 0; display: grid; gap: 10px; }
.join-code-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}
.join-code-heading > div { min-width: 0; }
.join-code-heading p { margin: 4px 0 0; color: var(--color-text-muted, #60708a); font-size: 13px; }
.copy-feedback { color: var(--color-success, #16803a); font-weight: 700; }
.participation-actions { display: grid; gap: 10px; padding-top: 15px; border-top: 1px solid var(--color-border, #d8e0ec); }
.participation-actions .app-button { width: 100%; min-height: 46px; }
.participation-closed,
.participation-feedback { margin: 0 !important; font-weight: 700; line-height: 1.5; }
.participation-closed { color: var(--color-text-muted, #60708a) !important; }
.participation-feedback { color: var(--color-success, #16803a) !important; }
.workshop-action-feedback {
  margin: 0;
  padding: 12px 14px;
  border: 1px solid var(--color-success-border, #bbf7d0);
  border-radius: var(--radius-md, 10px);
  background: var(--color-success-soft, #dcfce7);
  color: var(--color-success, #16803a);
  font-size: 14px;
  font-weight: 700;
  line-height: 1.5;
}
.join-code output {
  min-width: 0;
  padding: 11px;
  overflow-wrap: anywhere;
  border-radius: var(--radius-md, 10px);
  background: var(--color-surface-muted, #f7f9fc);
}
.code-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.error {
  padding: 12px;
  border: 1px solid var(--color-error-border, #fecaca);
  border-radius: var(--radius-md, 10px);
  background: var(--color-error-soft, #fee2e2);
  color: var(--color-error, #c62828);
  font-weight: 750;
}
.detail-actions {
  flex: none;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 15px 22px;
  border-top: 1px solid var(--color-border, #d8e0ec);
  background: var(--color-surface, #fff);
}
.detail-actions .app-button { min-height: 44px; }
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
@media (max-width: 520px) {
  .detail-overlay { align-items: flex-end; padding: 8px; }
  .detail-modal { max-height: calc(100dvh - 16px); border-radius: 18px; }
  .detail-header { padding: 17px 16px 13px; }
  .detail-body { gap: 13px; padding: 14px 16px; }
  .summary-card, .participants-card, .join-code { gap: 13px; padding: 14px; }
  .facts { grid-template-columns: 1fr; }
  .fact { align-items: center; }
  .summary-heading { align-items: flex-start; }
  .institutional-resource { align-items: flex-start; flex-wrap: wrap; }
  .institutional-resource .status-pill { margin-left: 30px; }
  .status-pill { white-space: normal; text-align: center; }
  .target-controls { align-items: stretch; flex-direction: column; }
  .join-code-heading { align-items: stretch; flex-direction: column; }
  .stepper { width: 100%; grid-template-columns: 48px 1fr 48px; }
  .save-target { width: 100%; }
  .detail-actions { padding: 12px 16px; }
  .detail-actions .app-button { flex: 1; }
}

@media (max-width: 380px) {
  .detail-overlay { padding: 0; }
  .detail-modal { max-height: 100dvh; border-radius: 0; border-inline: 0; }
  .summary-heading { display: grid; }
  .status-pill { justify-self: start; }
  .detail-actions { flex-direction: column-reverse; }
  .detail-actions .app-button { width: 100%; }
}
</style>
