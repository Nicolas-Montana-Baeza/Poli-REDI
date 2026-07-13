<script setup>
import { computed, ref, watch } from 'vue'

import ResourcePicker from './ResourcePicker.vue'
import DateTimePicker from './DateTimePicker.vue'

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

  errorMessage: {
    type: String,
    default: ''
  },

  submitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'close',
  'submit'
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
      new Date().toISOString().slice(0, 10)

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

  const start = new Date(`${form.value.date}T${form.value.hour}:00`)

  return start.getTime() <= Date.now()
}

const validateForm = () => {
  const errors = {}

  if (!form.value.resource?.id) {
    errors.resource = 'Selecciona una instalación.'
  }

  if (!form.value.date) {
    errors.date = 'Selecciona una fecha.'
  }

  if (!form.value.hour) {
    errors.hour = 'Selecciona una hora de inicio.'
  }

  if (Number(form.value.durationMinutes) <= 0) {
    errors.durationMinutes = 'La duración debe ser mayor a 0.'
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
              Crear Reserva
            </h2>

            <p>
              Completa la información de la reserva.
            </p>

          </div>

          <button
            class="close-btn"
            type="button"
            aria-label="Cerrar formulario de reserva"
            :disabled="submitting"
            @click="handleClose"
          >
            ✕
          </button>

        </div>

        <!-- SUMMARY -->
        <div
          v-if="form.resource"
          class="summary"
        >

          <div>

            <span>
              Instalación
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
        <ResourcePicker
          :resources="resources"
          :selected-id="form.resource?.id"
          @select="handleResourceSelect"
        />

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

/* MOBILE */
@media (max-width: 768px) {
  .modal {
    padding: 22px;

    border-radius: 24px;
  }

  .summary {
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
