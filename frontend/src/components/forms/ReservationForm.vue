<script setup>
import { computed, ref, watch } from 'vue'

import DateTimePicker from './DateTimePicker.vue'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getBusinessDateKey,
  getReservationDisplayStatus,
  parseReservationDateTime
} from '@/utils/reservationTime'
import { getReservationScheduleError } from '@/utils/reservationRules'
import { reservationsService } from '@/services/reservations.service'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },

  slot: {
    type: Object,
    default: null
  },

  resources: {
    type: Array,
    default: () => []
  },

  activities: {
    type: Array,
    default: () => []
  },

  policy: {
    type: Object,
    default: () => ({
      openingMinute: 480,
      closingMinute: 1320,
      slotIntervalMinutes: 15,
      allowedDurations: [60],
      resourceIds: []
    })
  },

  errorMessage: {
    type: String,
    default: ''
  },

  submitting: {
    type: Boolean,
    default: false
  },

  createdReservation: {
    type: Object,
    default: null
  },

  mode: {
    type: String,
    default: 'create',
    validator: value => ['create', 'detail'].includes(value)
  },

  reservation: {
    type: Object,
    default: null
  },

  canCancel: {
    type: Boolean,
    default: false
  },

  cancelDisabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'close',
  'submit'
,
  'cancel'
])

/* FORM */
const form = ref({
  resource: null,

  date: '',

  hour: '',

  durationMinutes: 60,

  activityId: null
})

const fieldErrors = ref({})

const detailJoinCode = ref('')
const detailJoinCodeCopied = ref(false)
const detailJoinCodeLoading = ref(false)
const detailJoinCodeError = ref('')

const groupConditionLabel = computed(() => {
  const condition =
    props.reservation?.groupCondition

  switch (condition) {
    case 'PENDING_MINIMUM':
      return 'Pendiente de participantes'

    case 'AT_RISK':
      return 'Bajo el mínimo de participantes'

    case 'MINIMUM_REACHED':
    case 'READY':
    case 'CONFIRMED':
      return 'Mínimo de participantes alcanzado'

    case 'CANCELLED':
      return 'Cancelada'

    case 'EXPIRED':
      return 'Expirada'

    default:
      return condition
        ? 'Estado del grupo actualizado'
        : 'Pendiente de participantes'
  }
})

const canRegenerateJoinCode = computed(() => {
  return (
    isDetailMode.value &&
    props.reservation?.isGroupReservation === true &&
    props.canCancel === true
  )
})

const regenerateDetailJoinCode = async () => {
  if (
    !canRegenerateJoinCode.value ||
    detailJoinCodeLoading.value
  ) {
    return
  }

  detailJoinCodeLoading.value = true
  detailJoinCode.value = ''
  detailJoinCodeCopied.value = false
  detailJoinCodeError.value = ''

  try {
    const result =
      await reservationsService.rotateJoinCode(
        props.reservation.id
      )

    detailJoinCode.value =
      result?.joinCode || ''

    if (!detailJoinCode.value) {
      detailJoinCodeError.value =
        'No se recibió el nuevo código de invitación.'
    }
  } catch (error) {
    detailJoinCodeError.value =
      error?.response?.data?.error ||
      'No fue posible generar un nuevo código.'
  } finally {
    detailJoinCodeLoading.value = false
  }
}

const copyDetailJoinCode = async () => {
  if (!detailJoinCode.value) {
    return
  }

  try {
    await navigator.clipboard.writeText(
      detailJoinCode.value
    )

    detailJoinCodeCopied.value = true
  } catch {
    detailJoinCodeError.value =
      'No fue posible copiar el código al portapapeles.'
  }
}


const isDetailMode = computed(() => {
  return props.mode === 'detail'
})

const detailStatus = computed(() => {
  return getReservationDisplayStatus(
    props.reservation
  )
})

const detailRows = computed(() => {
  const reservation = props.reservation

  if (!reservation) {
    return []
  }

  const rows = [
    {
      label: 'Recurso',
      value: reservation.resourceName || 'Recurso'
    },
    {
      label: 'Fecha',
      value: formatReservationDate(
        reservation.startTime
      )
    },
    {
      label: 'Horario',
      value: formatReservationTimeRange(
        reservation.startTime,
        reservation.durationMinutes
      )
    },
    {
      label: 'Duración',
      value: `${reservation.durationMinutes || 0} minutos`
    }
  ]

  if (
    reservation.title &&
    reservation.title !== 'Reserva'
  ) {
    rows.push({
      label: 'Actividad',
      value: reservation.title
    })
  }

  if (reservation.isGroupReservation === true) {
    rows.push({
      label: 'Participantes',
      value: `${reservation.participantCount || 0} / ${reservation.minimumParticipants || 0}`
    })
  }

  return rows
})

const handleDetailCancel = () => {
  if (
    !props.reservation ||
    props.cancelDisabled
  ) {
    return
  }

  emit('cancel', props.reservation)
}

const copiedJoinCode = ref(false)

const selectedActivityName = computed(() => {
  if (isOpenUseResource.value || !form.value.activityId) return ''

  return props.activities.find(
    activity => Number(activity.id) === Number(form.value.activityId)
  )?.name || ''
})

const reservationDateLabel = computed(() => {
  const [year, month, day] = String(form.value.date || '')
    .split('-')
    .map(Number)

  if (!year || !month || !day) return form.value.date

  return new Date(year, month - 1, day).toLocaleDateString('es-CL', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
})

const reservationEndHour = computed(() => {
  const [hour, minute] = String(form.value.hour || '')
    .split(':')
    .map(Number)

  if (!Number.isFinite(hour) || !Number.isFinite(minute)) return ''

  const total =
    hour * 60 +
    minute +
    Number(form.value.durationMinutes || 0)

  const end = total % (24 * 60)

  return `${String(Math.floor(end / 60)).padStart(2, '0')}:${String(end % 60).padStart(2, '0')}`
})

const reservationStatusLabel = computed(() => {
  switch (String(props.createdReservation?.status || '').toUpperCase()) {
    case 'CONFIRMED':
      return 'Confirmada'
    case 'PENDING':
      return 'Pendiente'
    case 'CANCELLED':
      return 'Cancelada'
    default:
      return props.createdReservation?.status || 'Creada'
  }
})

const copyJoinCode = async () => {
  const code = props.createdReservation?.joinCode
  if (!code) return

  try {
    await navigator.clipboard.writeText(code)
    copiedJoinCode.value = true
  } catch {
    copiedJoinCode.value = false
  }
}

const canSubmit = computed(() => {
  return (
    !props.submitting &&
    props.resources.length > 0
  )
})

const getDefaultActivityId = () => {
  return props.activities[0]?.id || null
}

const isOpenUseResource = computed(() => {
  return form.value.resource?.reservationMode === 'OPEN_USE'
})

const handleActivityUpdate = () => {
  fieldErrors.value.activityId = ''
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      fieldErrors.value = {}
    }
  }
)

watch(
  () => props.createdReservation,
  () => {
    copiedJoinCode.value = false
  }
)


watch(
  () => props.reservation?.id,
  () => {
    detailJoinCode.value = ''
    detailJoinCodeCopied.value = false
    detailJoinCodeLoading.value = false
    detailJoinCodeError.value = ''
  }
)

/* AUTOFILL FROM SLOT */
watch(
  () => props.slot,
  (slot) => {
    if (!slot) return

    form.value.resource =
      slot.resource || null

    form.value.hour =
      slot.hour || ''

    form.value.date =
      slot.date ||
      getBusinessDateKey()

    form.value.durationMinutes = 60

    form.value.activityId =
      slot.resource?.reservationMode === 'OPEN_USE'
        ? null
        : form.value.activityId || getDefaultActivityId()
  },
  {
    immediate: true
  }
)

watch(
  () => props.activities,
  () => {
    if (
      !props.visible ||
      form.value.activityId ||
      isOpenUseResource.value
    ) return

    form.value.activityId =
      getDefaultActivityId()
  },
  {
    immediate: true
  }
)

/* RESOURCE */
const handleResourceSelect = (resource) => {
  form.value.resource = resource

  if (resource?.reservationMode === 'OPEN_USE') {
    form.value.activityId = null
    fieldErrors.value.activityId = ''
  } else if (!form.value.activityId) {
    form.value.activityId = getDefaultActivityId()
  }

  fieldErrors.value.resource = ''
}

/* DATETIME */
const handleDateTimeUpdate = (data) => {
  form.value.date =
    data.date

  form.value.hour =
    data.hour

  form.value.durationMinutes =
    data.durationMinutes

  fieldErrors.value.date = ''
  fieldErrors.value.hour = ''
  fieldErrors.value.durationMinutes = ''
}

const isPastReservationStart = () => {
  if (!form.value.date || !form.value.hour) {
    return false
  }

  const start = parseReservationDateTime(
    `${form.value.date}T${form.value.hour}:00`
  )

  return start ? start.getTime() <= Date.now() : false
}

const validateForm = () => {
  const errors = {}

  if (!form.value.resource?.id) {
    errors.resource = 'Selecciona un recurso.'
  }

  if (!form.value.date) {
    errors.date = 'Selecciona una fecha.'
  }

  if (!form.value.hour) {
    errors.hour = 'Selecciona una hora de inicio.'
  }

  if (form.value.hour) {
    const scheduleError = getReservationScheduleError({
      hour: form.value.hour,
      durationMinutes: form.value.durationMinutes,
      policy: props.policy
    })

    if (scheduleError) {
      errors[scheduleError.field] = scheduleError.message
    }
  }

  if (
    form.value.date &&
    form.value.hour &&
    isPastReservationStart()
  ) {
    errors.hour = 'No puedes crear reservas en fechas u horarios pasados.'
  }

  fieldErrors.value = errors

  return Object.keys(errors).length === 0
}

/* SUBMIT */
const handleSubmit = () => {
  if (props.submitting || !validateForm()) {
    return
  }

  emit('submit', {
    ...form.value,
    activityId: isOpenUseResource.value
      ? null
      : form.value.activityId
  })
}

/* CLOSE */
const handleClose = () => {
  if (props.submitting) {
    return
  }

  emit('close')
}
</script>

<template>
  <Teleport to="body">

    <div
      v-if="visible"
      class="overlay"
      @click.self="handleClose"
    >

      <div class="modal">

        <!-- HEADER -->
        <div class="modal-header">

          <div>
            <h2>
              {{
                isDetailMode
                  ? 'Detalle de reserva'
                  : createdReservation
                    ? 'Reserva creada correctamente'
                    : 'Crear Reserva'
              }}
            </h2>

            <p>
              {{
                isDetailMode
                  ? 'Información de la reserva seleccionada.'
                  : createdReservation
                    ? 'Revisa la información antes de cerrar.'
                    : 'Completa la información de la reserva.'
              }}
            </p>
          </div>

          <button
            class="close-btn"
            type="button"
            aria-label="Cerrar modal de reserva"
            :disabled="submitting || cancelDisabled"
            @click="handleClose"
          >
            ✕
          </button>

        </div>

        <!-- DETAIL MODE -->
        <template v-if="isDetailMode">

          <section
            v-if="reservation"
            class="reservation-detail-panel"
          >
            <div class="detail-status-row">
              <span
                class="detail-status app-badge"
                :class="detailStatus.className"
              >
                {{ detailStatus.label }}
              </span>
            </div>

            <div class="detail-grid">
              <div
                v-for="row in detailRows"
                :key="row.label"
                class="detail-field"
              >
                <span>{{ row.label }}</span>
                <strong>{{ row.value }}</strong>
              </div>
            </div>

            <div
              v-if="reservation.isGroupReservation"
              class="group-detail"
            >
              <strong>Reserva grupal</strong>

              <p>
                Estado del grupo:
                {{ groupConditionLabel }}
              </p>
            </div>

            <div
              v-if="canRegenerateJoinCode"
              class="join-code-panel"
            >

              <template v-if="detailJoinCode">

                <span>
                  Nuevo código de invitación
                </span>

                <strong class="join-code">
                  {{ detailJoinCode }}
                </strong>

                <button
                  type="button"
                  class="app-button secondary"
                  @click="copyDetailJoinCode"
                >
                  {{
                    detailJoinCodeCopied
                      ? 'Copiado'
                      : 'Copiar código'
                  }}
                </button>

                <small>
                  Este código se muestra una sola vez.
                  Compártelo con los participantes.
                </small>

              </template>

              <template v-else>

                <span>
                  Código de invitación actual
                </span>

                <button
                  type="button"
                  class="app-button secondary"
                  :disabled="detailJoinCodeLoading"
                  @click="regenerateDetailJoinCode"
                >
                  {{
                    detailJoinCodeLoading
                      ? 'Generando...'
                      : 'Generar nuevo código'
                  }}
                </button>

                <small>
                  El código actual no puede recuperarse.
                  Al generar uno nuevo, el anterior dejará de funcionar.
                </small>

              </template>

              <p
                v-if="detailJoinCodeError"
                class="field-error"
              >
                {{ detailJoinCodeError }}
              </p>

            </div>

            <div
              v-if="errorMessage"
              class="form-error"
            >
              {{ errorMessage }}
            </div>

            <div class="actions detail-actions">
              <button
                v-if="canCancel"
                class="app-button danger"
                type="button"
                :disabled="cancelDisabled"
                @click="handleDetailCancel"
              >
                {{ cancelDisabled ? 'Cancelando...' : 'Cancelar reserva' }}
              </button>

              <button
                class="app-button primary"
                type="button"
                :disabled="cancelDisabled"
                @click="handleClose"
              >
                Cerrar
              </button>
            </div>
          </section>

        </template>

        <!-- CREATE MODE -->
        <template v-else>

        <template v-if="createdReservation">

          <section
            class="success-panel"
            role="status"
            aria-live="polite"
          >
            <div class="success-mark">
              ✓
            </div>

            <div class="success-grid">
              <div>
                <span>Recurso</span>
                <strong>{{ form.resource?.name }}</strong>
              </div>

              <div>
                <span>Fecha</span>
                <strong>{{ reservationDateLabel }}</strong>
              </div>

              <div>
                <span>Horario</span>
                <strong>
                  {{ form.hour }} - {{ reservationEndHour }}
                </strong>
              </div>

              <div>
                <span>Duración</span>
                <strong>{{ form.durationMinutes }} minutos</strong>
              </div>

              <div v-if="selectedActivityName">
                <span>Actividad</span>
                <strong>{{ selectedActivityName }}</strong>
              </div>

              <div>
                <span>Estado</span>
                <strong>{{ reservationStatusLabel }}</strong>
              </div>
            </div>

            <div
              v-if="createdReservation.joinCode"
              class="join-code-panel"
            >
              <span>Código de invitación</span>

              <strong class="join-code">
                {{ createdReservation.joinCode }}
              </strong>

              <button
                type="button"
                class="app-button secondary"
                @click="copyJoinCode"
              >
                {{ copiedJoinCode ? 'Copiado' : 'Copiar código' }}
              </button>

              <small>
                Compártelo con los participantes.
                Por seguridad, este código no puede recuperarse después.
              </small>
            </div>

            <div class="actions">
              <button
                class="submit-btn app-button primary"
                type="button"
                @click="handleClose"
              >
                Cerrar
              </button>
            </div>
          </section>

        </template>

        <template v-else>

        <!-- SUMMARY -->
        <div
          v-if="form.resource"
          class="summary"
        >

          <div>

            <span>
              Recurso
            </span>

            <strong>
              {{ form.resource.name }}
            </strong>

          </div>

          <div>

            <span>
              Inicio
            </span>

            <strong>
              {{ form.date }} · {{ form.hour }}
            </strong>

          </div>

        </div>

        <!-- RESOURCE -->
        <div class="field">
          <label for="reservation-resource">
            Recurso
          </label>

          <select
            id="reservation-resource"
            :value="form.resource?.id || ''"
            :disabled="submitting"
            :class="{ invalid: fieldErrors.resource }"
            @change="handleResourceSelect(
              resources.find(
                resource =>
                  String(resource.id) ===
                  String($event.target.value)
              ) || null
            )"
          >
            <option value="" disabled>
              Selecciona un recurso
            </option>

            <option
              v-for="resource in resources"
              :key="resource.id"
              :value="resource.id"
            >
              {{ resource.name }}
            </option>
          </select>
        </div>

        <p
          v-if="fieldErrors.resource"
          class="field-error"
        >
          {{ fieldErrors.resource }}
        </p>

        <!-- DATE TIME -->
        <DateTimePicker
          :initial-date="form.date"
          :initial-hour="form.hour"
        :initial-duration="form.durationMinutes"
          :policy="policy"
          fixed-date
          @update="handleDateTimeUpdate"
        />

        <div
          v-if="fieldErrors.date || fieldErrors.hour || fieldErrors.durationMinutes"
          class="field-errors"
        >
          <p v-if="fieldErrors.date">
            {{ fieldErrors.date }}
          </p>

          <p v-if="fieldErrors.hour">
            {{ fieldErrors.hour }}
          </p>

          <p v-if="fieldErrors.durationMinutes">
            {{ fieldErrors.durationMinutes }}
          </p>
        </div>

        <!-- ACTIVITY -->
        <div
          v-if="!isOpenUseResource"
          class="field"
        >

          <label>
            Actividad
          </label>

          <select
            v-model.number="form.activityId"
            :disabled="submitting"
            :class="{ invalid: fieldErrors.activityId }"
            @change="handleActivityUpdate"
          >

            <option
              :value="null"
            >
              {{ activities.length ? 'Sin actividad específica' : 'No hay actividades disponibles' }}
            </option>

            <option
              v-for="activity in activities"
              :key="activity.id"
              :value="activity.id"
              :title="activity.description"
            >
              {{ activity.name }}
            </option>

          </select>

        </div>

        <!-- FORM ERROR -->
        <div
          v-if="errorMessage"
          class="form-error"
        >
          {{ errorMessage }}
        </div>

        <!-- FOOTER -->
        <div class="actions">

          <button
            class="cancel-btn app-button secondary"
            type="button"
            :disabled="submitting"
            @click="handleClose"
          >
            Cancelar
          </button>

          <button
            class="submit-btn app-button primary"
            type="button"
            :disabled="!canSubmit"
            @click="handleSubmit"
          >
            {{ submitting ? 'Creando reserva...' : 'Confirmar Reserva' }}
          </button>

        </div>

        </template>

        </template>

      </div>

    </div>

  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15, 23, 42, 0.55);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 20px;

  z-index: 9999;

  backdrop-filter: blur(4px);
}

/* MODAL */
.modal {
  width: 100%;
  max-width: 720px;

  max-height: 90vh;

  overflow-y: auto;

  background: var(--color-surface);

  border-radius: var(--radius-xl);

  padding: 28px;

  display: flex;
  flex-direction: column;

  gap: 24px;

  box-shadow: var(--shadow-modal);
}

/* HEADER */
.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;
}

.modal-header h2 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: var(--color-text);
}

.modal-header p {
  margin-top: 6px;

  color: var(--color-text-muted);

  font-size: 14px;
}

/* CLOSE */
.close-btn {
  width: 44px;
  height: 44px;

  border: none;

  border-radius: var(--radius-md);

  background: var(--color-surface-soft);

  color: #334155;

  cursor: pointer;

  font-size: 18px;

  transition: 0.2s;
}

.close-btn:hover {
  background: #e2e8f0;
}

.close-btn:disabled {
  cursor: not-allowed;

  opacity: 0.55;
}

/* SUMMARY */
.summary {
  display: grid;

  grid-template-columns:
    repeat(2, 1fr);

  gap: 14px;

  background: var(--color-primary-soft);

  border: 1px solid #bfdbfe;

  border-radius: var(--radius-lg);

  padding: 16px;
}

.summary div {
  display: flex;
  flex-direction: column;

  gap: 4px;
}

.summary span {
  font-size: 12px;
  font-weight: 700;

  color: var(--color-text-muted);

  text-transform: uppercase;
  letter-spacing: 0;
}

.summary strong {
  font-size: 15px;

  color: var(--color-primary-strong);
}

/* FIELD */
.field {
  display: flex;
  flex-direction: column;

  gap: 8px;
}

.field label {
  font-size: 14px;
  font-weight: 700;

  color: #334155;
}

.field input,
.field select {
  width: 100%;

  height: 50px;

  border: 1px solid var(--color-border);

  border-radius: var(--radius-md);

  padding: 0 16px;

  font-size: 14px;

  outline: none;

  box-sizing: border-box;

  transition: 0.2s;
}

.field input.invalid,
.field select.invalid {
  border-color: var(--color-error-strong);

  box-shadow: var(--shadow-danger-focus);
}

.field select:disabled {
  color: var(--color-text-soft);

  background: var(--color-surface-muted);
}

/* FIELD ERRORS */
.field-error,
.field-errors p {
  margin: 0;

  color: var(--color-error);

  font-size: 13px;
  font-weight: 700;
}

.field-errors {
  display: flex;
  flex-direction: column;

  gap: 6px;

  padding: 12px 14px;

  border-radius: 14px;

  background: var(--color-error-soft);

  border: 1px solid var(--color-error-border);
}

/* ERROR */
/* ACTIONS */
.actions {
  display: flex;
  justify-content: flex-end;

  gap: 14px;

  padding-top: 8px;
}

.actions button {
  height: 50px;

  padding: 0 22px;

  font-size: 14px;
  font-weight: 700;

  cursor: pointer;

  transition: 0.2s;
}

.actions button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.submit-btn:hover {
  transform: translateY(-1px);
}

.submit-btn:disabled:hover {
  transform: none;

  box-shadow: none;
}

.success-panel {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.success-mark {
  width: 64px;
  height: 64px;
  display: grid;
  place-items: center;
  align-self: center;
  border-radius: 999px;
  background: var(--color-success-soft);
  color: var(--color-success);
  font-size: 34px;
  font-weight: 800;
}

.success-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.success-grid div,
.join-code-panel {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface-soft);
}

.success-grid span,
.join-code-panel > span {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text-muted);
  text-transform: uppercase;
}

.join-code-panel {
  text-align: center;
  align-items: center;
  background: var(--color-primary-soft);
  border-color: #bfdbfe;
}

.join-code {
  font-size: 28px;
  letter-spacing: 0.14em;
  color: var(--color-primary-strong);
}

.join-code-panel small {
  color: var(--color-text-muted);
  max-width: 460px;
}

.reservation-detail-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-status-row {
  display: flex;
  justify-content: flex-start;
}

.detail-status {
  font-size: 13px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.detail-field {
  display: flex;
  flex-direction: column;
  gap: 6px;

  padding: 16px;

  background: var(--color-surface-soft);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.detail-field span {
  color: var(--color-text-muted);

  font-size: 12px;
  font-weight: 700;

  text-transform: uppercase;
}

.detail-field strong {
  color: var(--color-text);

  font-size: 15px;
}

.group-detail {
  padding: 16px;

  background: var(--color-primary-soft);

  border: 1px solid #bfdbfe;
  border-radius: var(--radius-lg);
}

.group-detail p {
  margin: 6px 0 0;

  color: var(--color-text-muted);
}

.detail-actions {
  justify-content: flex-end;
}

/* MOBILE */
@media (max-width: 768px) {
  .modal {
    padding: 22px;

    border-radius: 24px;
  }

  .summary,
  .success-grid {
    grid-template-columns: 1fr;
  }

  .actions {
    flex-direction: column;
  }

  .actions button {
    width: 100%;
  }
}
</style>
