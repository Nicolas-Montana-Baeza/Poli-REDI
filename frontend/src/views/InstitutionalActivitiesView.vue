<script setup>
import {
  computed,
  onMounted,
  reactive,
  ref,
  watch
} from 'vue'

import { resourcesService } from '../services/resources.service'
import {
  institutionalUnitsService
} from '../services/institutionalUnits.service'
import {
  useInstitutionalActivitiesStore
} from '../stores/institutionalActivities'


const activitiesStore =
  useInstitutionalActivitiesStore()

const units = ref([])
const resources = ref([])
const selectedUnitId = ref('')

const loadingBootstrap = ref(false)
const successMessage = ref('')

const activityTypes = [
  {
    value: 'ACADEMIC_CLASS',
    label: 'Clase académica'
  },
  {
    value: 'WORKSHOP',
    label: 'Taller'
  },
  {
    value: 'TRAINING',
    label: 'Entrenamiento'
  },
  {
    value: 'EVENT',
    label: 'Evento'
  },
  {
    value: 'CHAMPIONSHIP',
    label: 'Campeonato'
  },
  {
    value: 'OTHER',
    label: 'Otro'
  }
]

const weekdays = [
  { value: 1, label: 'Lunes' },
  { value: 2, label: 'Martes' },
  { value: 3, label: 'Miércoles' },
  { value: 4, label: 'Jueves' },
  { value: 5, label: 'Viernes' },
  { value: 6, label: 'Sábado' },
  { value: 7, label: 'Domingo' }
]

const form = reactive({
  resourceId: '',
  activityType: 'ACADEMIC_CLASS',
  title: '',
  description: '',
  requiresEnrollment: false,
  capacity: '',
  scheduleType: 'SINGLE',

  specificDate: '',
  dayOfWeek: 1,
  startTime: '10:00',
  endTime: '11:00',
  validFrom: '',
  validTo: ''
})


const selectedUnit = computed(() => {
  return units.value.find(
    (unit) =>
      unit.id === Number(selectedUnitId.value)
  ) || null
})


const activities = computed(() => {
  if (!selectedUnitId.value) {
    return []
  }

  return (
    activitiesStore.activitiesByUnit[
      Number(selectedUnitId.value)
    ] || []
  )
})


const activeResources = computed(() => {
  return resources.value
    .filter(
      (resource) =>
        resource.isActive !== false
    )
    .sort(
      (a, b) =>
        String(a.name)
          .localeCompare(
            String(b.name),
            'es'
          )
    )
})


const typeLabel = (value) => {
  return (
    activityTypes.find(
      (type) => type.value === value
    )?.label || value
  )
}


const weekdayLabel = (value) => {
  return (
    weekdays.find(
      (day) => day.value === value
    )?.label || value
  )
}


const scheduleDescription = (schedule) => {
  if (schedule.scheduleType === 'SINGLE') {
    return (
      `${schedule.specificDate} · ` +
      `${schedule.startTime}–${schedule.endTime}`
    )
  }

  return (
    `${weekdayLabel(schedule.dayOfWeek)} · ` +
    `${schedule.startTime}–${schedule.endTime} · ` +
    `${schedule.validFrom} → ${schedule.validTo}`
  )
}


const resetForm = () => {
  form.resourceId =
    activeResources.value[0]?.id || ''

  form.activityType = 'ACADEMIC_CLASS'
  form.title = ''
  form.description = ''
  form.requiresEnrollment = false
  form.capacity = ''
  form.scheduleType = 'SINGLE'

  form.specificDate = ''
  form.dayOfWeek = 1
  form.startTime = '10:00'
  form.endTime = '11:00'
  form.validFrom = ''
  form.validTo = ''
}


const loadSelectedUnit = async () => {
  successMessage.value = ''
  activitiesStore.clearError()

  if (!selectedUnitId.value) {
    return
  }

  try {
    await activitiesStore.loadByUnit(
      Number(selectedUnitId.value)
    )
  } catch {
    // El store conserva el mensaje de error.
  }
}


watch(
  selectedUnitId,
  loadSelectedUnit
)


watch(
  () => form.activityType,
  (activityType) => {
    // WORKSHOP exige inscripción y capacidad en el backend.
    if (activityType === 'WORKSHOP') {
      form.requiresEnrollment = true

      if (!form.capacity) {
        form.capacity = 20
      }
    }
  }
)


watch(
  () => form.requiresEnrollment,
  (enabled) => {
    if (!enabled) {
      form.capacity = ''
    }
  }
)


const submit = async () => {
  activitiesStore.clearError()
  successMessage.value = ''

  if (!selectedUnitId.value) {
    activitiesStore.error =
      'Debes seleccionar una unidad institucional.'

    return
  }

  if (!form.resourceId) {
    activitiesStore.error =
      'Debes seleccionar un recurso.'

    return
  }

  const schedule =
    form.scheduleType === 'SINGLE'
      ? {
          scheduleType: 'SINGLE',
          specificDate: form.specificDate,
          startTime: form.startTime,
          endTime: form.endTime
        }
      : {
          scheduleType: 'WEEKLY',
          dayOfWeek: Number(form.dayOfWeek),
          startTime: form.startTime,
          endTime: form.endTime,
          validFrom: form.validFrom,
          validTo: form.validTo
        }

  const payload = {
    unitId: Number(selectedUnitId.value),
    resourceId: Number(form.resourceId),
    activityType: form.activityType,
    title: form.title.trim(),
    description: form.description.trim(),
    requiresEnrollment:
      form.requiresEnrollment,
    schedules: [schedule]
  }

  if (form.requiresEnrollment) {
    payload.capacity =
      Number(form.capacity)
  }

  try {
    const activity =
      await activitiesStore.create(payload)

    successMessage.value =
      `Actividad "${activity.title}" creada correctamente.`

    resetForm()
  } catch {
    // El store mantiene el error del backend.
  }
}


onMounted(async () => {
  loadingBootstrap.value = true

  try {
    const [
      unitResponse,
      resourceResponse
    ] = await Promise.all([
      institutionalUnitsService.getAll(),
      resourcesService.getAll()
    ])

    units.value =
      Array.isArray(unitResponse)
        ? unitResponse
        : []

    resources.value =
      Array.isArray(resourceResponse)
        ? resourceResponse
        : []

    if (units.value.length > 0) {
      selectedUnitId.value =
        units.value[0].id
    }

    resetForm()
  } catch (error) {
    activitiesStore.error =
      error?.response?.data?.error ||
      'No se pudo cargar la programación institucional.'
  } finally {
    loadingBootstrap.value = false
  }
})
</script>


<template>
  <main class="institutional-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">
          Programación institucional
        </p>

        <h1>
          Actividades
        </h1>

        <p class="page-description">
          Programa clases, entrenamientos, talleres,
          eventos y otras actividades institucionales.
        </p>
      </div>
    </header>

    <div
      v-if="activitiesStore.error"
      class="message error"
      role="alert"
    >
      {{ activitiesStore.error }}
    </div>

    <div
      v-if="successMessage"
      class="message success"
    >
      {{ successMessage }}
    </div>

    <p
      v-if="loadingBootstrap"
      class="state-card"
    >
      Cargando programación institucional...
    </p>

    <section
      v-else-if="units.length === 0"
      class="state-card"
    >
      <h2>
        Sin unidades disponibles
      </h2>

      <p>
        Tu cuenta no tiene unidades institucionales
        habilitadas para programar.
      </p>
    </section>

    <template v-else>
      <section class="panel">
        <label class="field">
          <span>
            Unidad institucional
          </span>

          <select
            v-model="selectedUnitId"
          >
            <option
              v-for="unit in units"
              :key="unit.id"
              :value="unit.id"
            >
              {{ unit.name }} · {{ unit.code }}
            </option>
          </select>
        </label>
      </section>

      <div class="workspace">
        <section class="panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">
                Nueva programación
              </p>

              <h2>
                Crear actividad
              </h2>
            </div>
          </div>

          <form
            class="activity-form"
            @submit.prevent="submit"
          >
            <label class="field">
              <span>
                Título
              </span>

              <input
                v-model="form.title"
                type="text"
                required
                maxlength="160"
                placeholder="Ej. Clase de Programación I"
              >
            </label>

            <label class="field">
              <span>
                Tipo
              </span>

              <select
                v-model="form.activityType"
              >
                <option
                  v-for="type in activityTypes"
                  :key="type.value"
                  :value="type.value"
                >
                  {{ type.label }}
                </option>
              </select>
            </label>

            <label class="field">
              <span>
                Recurso
              </span>

              <select
                v-model="form.resourceId"
                required
              >
                <option
                  value=""
                  disabled
                >
                  Selecciona un recurso
                </option>

                <option
                  v-for="resource in activeResources"
                  :key="resource.id"
                  :value="resource.id"
                >
                  {{ resource.name }}
                </option>
              </select>
            </label>

            <label class="field full">
              <span>
                Descripción
              </span>

              <textarea
                v-model="form.description"
                rows="3"
                placeholder="Descripción opcional"
              />
            </label>

            <label class="checkbox-field full">
              <input
                v-model="form.requiresEnrollment"
                type="checkbox"
                :disabled="
                  form.activityType === 'WORKSHOP'
                "
              >

              <span>
                Requiere inscripción
              </span>
            </label>

            <label
              v-if="form.requiresEnrollment"
              class="field"
            >
              <span>
                Capacidad
              </span>

              <input
                v-model.number="form.capacity"
                type="number"
                min="1"
                required
              >
            </label>

            <fieldset class="schedule-box full">
              <legend>
                Programación
              </legend>

              <div class="schedule-type">
                <label>
                  <input
                    v-model="form.scheduleType"
                    type="radio"
                    value="SINGLE"
                  >
                  Única
                </label>

                <label>
                  <input
                    v-model="form.scheduleType"
                    type="radio"
                    value="WEEKLY"
                  >
                  Semanal
                </label>
              </div>

              <div class="schedule-grid">
                <label
                  v-if="form.scheduleType === 'SINGLE'"
                  class="field"
                >
                  <span>
                    Fecha
                  </span>

                  <input
                    v-model="form.specificDate"
                    type="date"
                    required
                  >
                </label>

                <template v-else>
                  <label class="field">
                    <span>
                      Día
                    </span>

                    <select
                      v-model.number="form.dayOfWeek"
                    >
                      <option
                        v-for="day in weekdays"
                        :key="day.value"
                        :value="day.value"
                      >
                        {{ day.label }}
                      </option>
                    </select>
                  </label>

                  <label class="field">
                    <span>
                      Desde
                    </span>

                    <input
                      v-model="form.validFrom"
                      type="date"
                      required
                    >
                  </label>

                  <label class="field">
                    <span>
                      Hasta
                    </span>

                    <input
                      v-model="form.validTo"
                      type="date"
                      required
                    >
                  </label>
                </template>

                <label class="field">
                  <span>
                    Hora inicio
                  </span>

                  <input
                    v-model="form.startTime"
                    type="time"
                    required
                  >
                </label>

                <label class="field">
                  <span>
                    Hora término
                  </span>

                  <input
                    v-model="form.endTime"
                    type="time"
                    required
                  >
                </label>
              </div>
            </fieldset>

            <div class="form-actions full">
              <button
                class="primary-button"
                type="submit"
                :disabled="activitiesStore.creating"
              >
                {{
                  activitiesStore.creating
                    ? 'Creando...'
                    : 'Crear actividad'
                }}
              </button>
            </div>
          </form>
        </section>

        <section class="panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">
                {{
                  selectedUnit?.code ||
                  'Unidad'
                }}
              </p>

              <h2>
                Actividades programadas
              </h2>
            </div>
          </div>

          <p
            v-if="activitiesStore.loading"
            class="state-card"
          >
            Cargando actividades...
          </p>

          <p
            v-else-if="activities.length === 0"
            class="state-card"
          >
            Esta unidad todavía no tiene
            actividades programadas.
          </p>

          <div
            v-else
            class="activities-list"
          >
            <article
              v-for="activity in activities"
              :key="activity.id"
              class="activity-card"
            >
              <header>
                <div>
                  <span class="activity-type">
                    {{ typeLabel(activity.activityType) }}
                  </span>

                  <h3>
                    {{ activity.title }}
                  </h3>
                </div>

                <span class="status-badge">
                  {{ activity.status }}
                </span>
              </header>

              <p class="resource-name">
                {{ activity.resourceName }}
              </p>

              <p
                v-if="activity.description"
                class="activity-description"
              >
                {{ activity.description }}
              </p>

              <div class="schedule-list">
                <div
                  v-for="schedule in activity.schedules"
                  :key="schedule.id"
                  class="schedule-row"
                >
                  <strong>
                    {{ schedule.scheduleType }}
                  </strong>

                  <span>
                    {{ scheduleDescription(schedule) }}
                  </span>
                </div>
              </div>

              <small>
                Creada por
                {{ activity.createdBy || 'usuario institucional' }}
              </small>
            </article>
          </div>
        </section>
      </div>
    </template>
  </main>
</template>


<style scoped>
.institutional-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.page-header,
.section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.page-header h1,
.section-heading h2 {
  margin: 0.2rem 0;
}

.eyebrow {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  opacity: 0.6;
}

.page-description {
  margin: 0.4rem 0 0;
  opacity: 0.7;
}

.workspace {
  display: grid;
  grid-template-columns:
    minmax(0, 0.9fr)
    minmax(0, 1.1fr);
  gap: 1.25rem;
  align-items: start;
}

.panel,
.state-card {
  border: 1px solid rgba(127, 127, 127, 0.2);
  border-radius: 1rem;
  padding: 1.2rem;
  background: var(--card-bg, transparent);
}

.activity-form {
  display: grid;
  grid-template-columns:
    repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field > span {
  font-size: 0.85rem;
  font-weight: 600;
}

.field input,
.field select,
.field textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(127, 127, 127, 0.3);
  border-radius: 0.65rem;
  padding: 0.7rem;
  font: inherit;
  background: inherit;
  color: inherit;
}

.full {
  grid-column: 1 / -1;
}

.checkbox-field {
  display: flex;
  gap: 0.55rem;
  align-items: center;
}

.schedule-box {
  border: 1px solid rgba(127, 127, 127, 0.2);
  border-radius: 0.8rem;
  padding: 1rem;
}

.schedule-box legend {
  padding: 0 0.4rem;
  font-weight: 700;
}

.schedule-type {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.schedule-type label {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}

.schedule-grid {
  display: grid;
  grid-template-columns:
    repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.primary-button {
  border: 0;
  border-radius: 0.7rem;
  padding: 0.75rem 1rem;
  cursor: pointer;
  font: inherit;
  font-weight: 700;
}

.primary-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.message {
  border-radius: 0.75rem;
  padding: 0.9rem 1rem;
}

.message.error {
  background: rgba(220, 38, 38, 0.12);
}

.message.success {
  background: rgba(22, 163, 74, 0.12);
}

.activities-list {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  margin-top: 1rem;
}

.activity-card {
  border: 1px solid rgba(127, 127, 127, 0.2);
  border-radius: 0.85rem;
  padding: 1rem;
}

.activity-card header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.activity-card h3 {
  margin: 0.25rem 0;
}

.activity-type,
.status-badge {
  font-size: 0.72rem;
  font-weight: 700;
}

.status-badge {
  flex-shrink: 0;
  border-radius: 999px;
  padding: 0.3rem 0.55rem;
  background: rgba(22, 163, 74, 0.13);
}

.resource-name {
  margin: 0.5rem 0;
  font-weight: 600;
}

.activity-description {
  opacity: 0.75;
}

.schedule-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  margin: 0.8rem 0;
}

.schedule-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  border-radius: 0.6rem;
  padding: 0.6rem;
  background: rgba(127, 127, 127, 0.08);
}

.activity-card small {
  opacity: 0.55;
}

@media (max-width: 900px) {
  .workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 650px) {
  .activity-form,
  .schedule-grid {
    grid-template-columns: 1fr;
  }

  .full {
    grid-column: auto;
  }

  .schedule-row {
    flex-direction: column;
  }
}
</style>
