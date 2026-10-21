<script setup>
import {
  computed,
  nextTick,
  onMounted,
  reactive,
  ref,
  watch
} from 'vue'

import { resourcesService } from '../services/resources.service'
import { reservationPolicyService } from '../services/reservationPolicy.service'
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
const institutionalHours = ref(null)
const clockMinutes = (value) => {
  const match = /^(\d{2}):(\d{2})$/.exec(String(value || ''))
  return match ? Number(match[1])*60 + Number(match[2]) : NaN
}
const scheduleHoursLabel = computed(() => {
  const hours = institutionalHours.value
  if (!hours) return 'Consultando jornada vigente...'
  return `Jornada permitida: ${hours.start}–${hours.end}`
})
const editingId = ref(null)
const pendingCancellation = ref(null)
const activitySearch = ref('')
const statusFilter = ref('SCHEDULED')
const editorSection = ref(null)

const statusFilters = [
  { value: 'SCHEDULED', label: 'Programadas' },
  { value: 'CANCELLED', label: 'Canceladas' },
  { value: 'ALL', label: 'Todas' }
]

const statusLabel = (status) => ({
  SCHEDULED: 'Programada',
  CANCELLED: 'Cancelada',
  COMPLETED: 'Completada',
  DRAFT: 'Borrador'
})[status] || status

const focusEditor = () => {
  nextTick(() => {
    editorSection.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
    editorSection.value?.querySelector('input[type="text"]')?.focus()
  })
}

const startNew = () => {
  resetForm()
  activitiesStore.clearError()
  successMessage.value = ''
  focusEditor()
}

const askToCancel = (activity) => {
  pendingCancellation.value = activity.id
  activitiesStore.clearError()
  successMessage.value = ''
  nextTick(() => {
    document.querySelector('.cancel-confirm .danger-button')?.focus()
  })
}


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


const scheduledCount = computed(
  () => activities.value.filter(activity => activity.status === 'SCHEDULED').length
)

const cancelledCount = computed(
  () => activities.value.filter(activity => activity.status === 'CANCELLED').length
)

const filteredActivities = computed(() => {
  const search = activitySearch.value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim()
    .toLowerCase()

  return activities.value.filter(activity => {
    if (statusFilter.value !== 'ALL' && activity.status !== statusFilter.value) {
      return false
    }
    if (!search) return true

    return [
      activity.title,
      activity.description,
      activity.resourceName,
      typeLabel(activity.activityType)
    ].some(value => String(value || '')
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .toLowerCase()
      .includes(search))
  })
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
  editingId.value = null
  pendingCancellation.value = null
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


const beginEdit = (activity) => {
  if (activity.status !== 'SCHEDULED') return
  if ((activity.schedules || []).length !== 1) {
    activitiesStore.error = 'Esta actividad tiene múltiples reglas horarias; usa la edición avanzada más adelante.'
    return
  }
  if (activity.enrollmentCount > 0) {
    activitiesStore.error = 'Esta actividad tiene inscripciones activas y no puede editarse por ahora.'
    return
  }
  const schedule = activity.schedules[0]
  pendingCancellation.value = null
  editingId.value = activity.id
  activitiesStore.clearError()
  successMessage.value = ''
  form.resourceId = activity.resourceId
  form.activityType = activity.activityType
  form.title = activity.title
  form.description = activity.description || ''
  form.requiresEnrollment = activity.requiresEnrollment
  form.capacity = activity.capacity ?? ''
  form.scheduleType = schedule.scheduleType
  form.specificDate = schedule.specificDate || ''
  form.dayOfWeek = schedule.dayOfWeek || 1
  form.startTime = schedule.startTime
  form.endTime = schedule.endTime
  form.validFrom = schedule.validFrom || ''
  form.validTo = schedule.validTo || ''
  focusEditor()
}

const cancelActivity = async (activity) => {
  activitiesStore.clearError()
  successMessage.value = ''
  try {
    await activitiesStore.cancel(activity.id, Number(selectedUnitId.value))
    pendingCancellation.value = null
    if (editingId.value === activity.id) resetForm()
    successMessage.value = `Actividad "${activity.title}" cancelada.`
  } catch {
    // El store conserva el error.
  }
}

const loadSelectedUnit = async () => {
  activitySearch.value = ''
  statusFilter.value = 'SCHEDULED'
  pendingCancellation.value = null
  resetForm()
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

  if (!institutionalHours.value) {
    activitiesStore.error = 'No se pudo consultar la jornada permitida. Actualiza la página.'
    return
  }
  const starts = clockMinutes(form.startTime)
  const ends = clockMinutes(form.endTime)
  const opening = clockMinutes(institutionalHours.value.start)
  const closing = clockMinutes(institutionalHours.value.end)
  if (!Number.isFinite(starts) || !Number.isFinite(ends) ||
      starts < opening || ends > closing || ends <= starts) {
    activitiesStore.error =
      `La actividad debe comenzar y terminar dentro de ${institutionalHours.value.start}–${institutionalHours.value.end}.`
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
    const activity = editingId.value !== null
      ? await activitiesStore.update(editingId.value, payload)
      : await activitiesStore.create(payload)

    successMessage.value = editingId.value !== null
      ? `Actividad "${activity.title}" actualizada correctamente.`
      : `Actividad "${activity.title}" creada correctamente.`

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
      resourceResponse,
      policyResponse
    ] = await Promise.all([
      institutionalUnitsService.getAll(),
      resourcesService.getAll(),
      reservationPolicyService.getCurrent()
    ])
    const toClock = (minutes) =>
      `${String(Math.floor(minutes / 60)).padStart(2, '0')}:${String(minutes % 60).padStart(2, '0')}`
    institutionalHours.value = {
      start: toClock(policyResponse.openingMinute),
      end: toClock(policyResponse.closingMinute)
    }

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
        <p class="eyebrow">GESTIÓN INSTITUCIONAL · MVP3</p>
        <h1>Programación de actividades</h1>
        <p class="page-description">
          Organiza clases, talleres y eventos. Consulta la programación,
          modifica actividades y cancélalas cuando sea necesario.
        </p>
        <p class="page-context">Entorno de pruebas · Los cambios se reflejan en disponibilidad.</p>
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

      <div class="activity-overview" aria-label="Resumen de actividades de la unidad">
        <div class="overview-item">
          <span class="overview-label">Programadas</span>
          <strong class="overview-value">{{ scheduledCount }}</strong>
        </div>
        <div class="overview-item">
          <span class="overview-label">Canceladas</span>
          <strong class="overview-value">{{ cancelledCount }}</strong>
        </div>
        <div class="overview-item overview-caption">
          <strong>{{ selectedUnit?.name || 'Unidad institucional' }}</strong>
          <span>Selecciona una actividad para editarla o gestionar su estado.</span>
        </div>
      </div>

      <div class="workspace">
        <section ref="editorSection" class="panel editor-panel" :class="{ 'editing-mode': editingId !== null }">
          <div class="section-heading">
            <div>
              <p class="eyebrow">
                {{ editingId === null ? 'Nueva programación' : 'Editar programación' }}
              </p>

              <h2>
                {{ editingId === null ? 'Crear actividad' : 'Actualizar actividad' }}
              </h2>
              <p class="section-description">
                {{ editingId === null
                  ? 'Completa la información y define cuándo se utilizará la instalación.'
                  : 'Modifica los datos y guarda para actualizar la programación.' }}
              </p>
            </div>
            <span v-if="editingId !== null" class="editing-pill">Editando #{{ editingId }}</span>
          </div>

          <p v-if="editingId !== null" class="editor-notice" role="status">
            Los cambios de horario se aplican a la actividad completa.
            El recurso y la unidad no se modifican en esta edición.
          </p>

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
                :disabled="editingId !== null"
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
                    :min="institutionalHours?.start"
                    :max="institutionalHours?.end"
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
                    :min="institutionalHours?.start"
                    :max="institutionalHours?.end"
                    required
                  >
                </label>
              </div>
              <p class="field-hint">{{ scheduleHoursLabel }}. Se aplica a actividades únicas y semanales.</p>
            </fieldset>

            <div class="form-actions full">
              <button
                class="primary-button"
                type="submit"
                :disabled="activitiesStore.creating || activitiesStore.saving"
              >
                {{ editingId !== null
                  ? (activitiesStore.saving ? 'Guardando...' : 'Guardar cambios')
                  : (activitiesStore.creating ? 'Creando...' : 'Crear actividad')
                }}
              </button>
              <button
                v-if="editingId !== null"
                type="button"
                class="secondary-button"
                @click="resetForm"
              >
                Descartar edición
              </button>
            </div>
          </form>
        </section>

        <section class="panel catalog-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">{{ selectedUnit?.code || 'Unidad' }} · PROGRAMACIÓN</p>
              <h2>Actividades de la unidad</h2>
              <p class="section-description">Busca y administra actividades existentes.</p>
            </div>
            <button type="button" class="create-shortcut" @click="startNew">
              + Nueva actividad
            </button>
          </div>

          <div class="catalog-tools">
            <label class="search-field">
              <span class="sr-only">Buscar actividad por nombre, recurso o tipo</span>
              <input
                v-model="activitySearch"
                type="search"
                placeholder="Buscar actividad o instalación..."
                autocomplete="off"
              >
            </label>
            <div class="status-filters" role="group" aria-label="Filtrar actividades por estado">
              <button
                v-for="filter in statusFilters"
                :key="filter.value"
                type="button"
                :aria-pressed="statusFilter === filter.value"
                :class="['filter-button', { active: statusFilter === filter.value }]"
                @click="statusFilter = filter.value"
              >
                {{ filter.label }}
              </button>
            </div>
          </div>

          <p
            v-if="activitiesStore.loading"
            class="state-card"
          >
            Cargando actividades...
          </p>

          <div v-else-if="activities.length === 0" class="empty-catalog" role="status">
            <strong>Aún no hay actividades registradas</strong>
            <span>Crea la primera actividad para esta unidad institucional.</span>
            <button type="button" class="secondary-button" @click="startNew">Crear actividad</button>
          </div>

          <div v-else-if="filteredActivities.length === 0" class="empty-catalog" role="status">
            <strong>No encontramos coincidencias</strong>
            <span>Prueba otro nombre o cambia el filtro de estado.</span>
            <button type="button" class="secondary-button" @click="activitySearch = ''; statusFilter = 'ALL'">Ver todas</button>
          </div>

          <div v-else class="activities-list">
            <article
              v-for="activity in filteredActivities"
              :key="activity.id"
              class="activity-card"
              :class="{
                'is-editing': editingId === activity.id,
                'is-cancelled': activity.status === 'CANCELLED'
              }"
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

                <span class="status-badge" :class="'status-' + activity.status.toLowerCase()">
                  {{ statusLabel(activity.status) }}
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
                    {{ schedule.scheduleType === 'WEEKLY' ? 'Semanal' : 'Única' }}
                  </strong>

                  <span>
                    {{ scheduleDescription(schedule) }}
                  </span>
                </div>
              </div>

              <small class="creator-note">
                Creada por {{ activity.createdBy || 'usuario institucional' }}
              </small>
              <p v-if="activity.status === 'SCHEDULED' && Number(activity.enrollmentCount) > 0"
                 class="enrollment-warning">
                Esta actividad tiene {{ activity.enrollmentCount }} inscripción(es) activa(s):
                no se puede editar hasta resolverlas.
              </p>

              <div v-if="activity.status === 'SCHEDULED' && pendingCancellation !== activity.id"
                   class="manage-actions">
                <button
                  type="button"
                  class="secondary-button"
                  :disabled="Number(activity.enrollmentCount) > 0 || (activity.schedules || []).length !== 1"
                  :title="Number(activity.enrollmentCount) > 0
                    ? 'No se puede editar una actividad con inscripciones activas'
                    : (activity.schedules || []).length !== 1
                      ? 'La edición de varias reglas horarias no está disponible'
                      : 'Editar actividad'"
                  @click="beginEdit(activity)"
                >Editar</button>
                <button type="button" class="danger-button" @click="askToCancel(activity)">
                  Cancelar actividad
                </button>
              </div>
              <div v-if="pendingCancellation === activity.id"
                   class="cancel-confirm" role="group" aria-label="Confirmar cancelación de actividad">
                <strong>¿Cancelar «{{ activity.title }}»?</strong>
                <p>La actividad dejará de ocupar sus horarios y quedará visible en el historial.
                   También se cancelarán sus inscripciones si las tiene.</p>
                <div class="cancel-actions">
                  <button type="button" class="danger-button"
                    :disabled="activitiesStore.cancellingId === activity.id"
                    @click="cancelActivity(activity)">
                    {{ activitiesStore.cancellingId === activity.id ? 'Cancelando...' : 'Sí, cancelar actividad' }}
                  </button>
                  <button type="button" class="secondary-button"
                    :disabled="activitiesStore.cancellingId === activity.id"
                    @click="pendingCancellation = null">Volver</button>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </template>
  </main>
</template>


<style scoped>
/* ADMIN-005: utiliza los tokens visuales globales de Poli-REDI. */
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
    minmax(320px, 0.95fr)
    minmax(0, 1.15fr);
  gap: 1.25rem;
  align-items: start;
}

.panel,
.state-card {
  border: 1px solid rgba(127, 127, 127, 0.2);
  border-radius: 1rem;
  padding: 1.2rem;
  background: var(--color-surface);
  box-shadow: var(--shadow-card);
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
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 0.75rem;
  font: inherit;
  background: var(--color-surface);
  color: var(--color-text);
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
  gap: .6rem;
  flex-wrap: wrap;
}

.primary-button {
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-lg);
  background: var(--color-primary);
  color: var(--color-primary-contrast);
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
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-xl);
  padding: 1rem;
  background: var(--color-surface);
  transition: border-color .18s ease, box-shadow .18s ease;
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
  background: var(--color-success-soft);
  color: var(--color-success);
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

.manage-actions, .cancel-confirm {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
  margin-top: 0.85rem;
}
.secondary-button, .danger-button {
  cursor: pointer;
  border: 1px solid rgba(127,127,127,.3);
  border-radius: .65rem;
  background: transparent;
  color: inherit;
  padding: .65rem .9rem;
  font: inherit;
}
.danger-button {color: #c44848; border-color: #c44848;}
.danger-button:disabled { opacity:.5; cursor:not-allowed; }
.cancel-confirm {background: rgba(196,72,72,.08); border-radius: .8rem; padding: .85rem;}
.cancel-confirm p { width: 100%; margin: 0; }

/* Cabecera, resumen y espacio de trabajo */
.institutional-page { max-width: 1440px; margin: 0 auto; width: 100%; }
.page-header h1 { font-size: clamp(1.45rem, 3vw, 2rem); letter-spacing: -.025em; }
.page-description { color: var(--color-text-muted); line-height: 1.55; max-width: 72ch; }
.page-context { margin-top: .6rem; font-size: .8rem; color: var(--color-text-muted); }
.unit-panel { padding: .9rem 1.15rem; }
.unit-panel .field { max-width: 520px; }
.activity-overview {
  display: grid;
  grid-template-columns: minmax(130px, 170px) minmax(130px, 170px) minmax(0, 1fr);
  gap: .75rem;
}
.overview-item {
  display: flex; flex-direction: column; gap: .3rem;
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-xl); padding: .95rem 1.15rem;
  background: var(--color-surface);
}
.overview-label { color: var(--color-text-muted); font-size: .81rem; font-weight: 600; }
.overview-value { font-size: 1.65rem; line-height: 1.15; font-variant-numeric: tabular-nums; }
.overview-caption { justify-content: center; gap: .35rem; }
.overview-caption strong { font-size: .92rem; }
.overview-caption span { font-size: .81rem; color: var(--color-text-muted); }
.section-description { font-size: .84rem; line-height: 1.45; margin-top: .4rem; color: var(--color-text-muted); }
.editor-panel { scroll-margin-top: 5.5rem; }
.editor-panel.editing-mode { border-color: var(--color-primary); box-shadow: var(--shadow-focus); }
.editing-pill { flex-shrink: 0; font-size: .76rem; font-weight: 700;
  border-radius: var(--radius-pill); padding: .35rem .7rem;
  background: var(--color-primary-soft); color: var(--color-primary-strong); }
.editor-notice { margin-top: .9rem; padding: .75rem .9rem; border-radius: var(--radius-lg);
  background: var(--color-primary-soft); color: var(--color-text); font-size: .84rem; line-height: 1.5; }
.field input:focus-visible, .field textarea:focus-visible, .field select:focus-visible,
.search-field input:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 1px; }
.catalog-tools { display: flex; flex-direction: column; gap: .8rem; margin-top: 1rem; }
.search-field input { width: 100%; min-height: 42px; padding: .7rem .85rem;
  border: 1px solid var(--color-border); border-radius: var(--radius-lg);
  background: var(--color-surface); color: var(--color-text); }
.status-filters { display: flex; flex-wrap: wrap; gap: .4rem; }
.filter-button { border: 1px solid var(--color-border); border-radius: var(--radius-pill);
  background: var(--color-surface); color: var(--color-text-muted);
  padding: .45rem .85rem; font-size: .8rem; font-weight: 700; cursor: pointer; }
.filter-button.active { border-color: var(--color-primary); background: var(--color-primary-soft);
  color: var(--color-primary-strong); }
.create-shortcut { border: 1px solid var(--color-primary);
  border-radius: var(--radius-lg); color: var(--color-primary-strong);
  background: var(--color-primary-soft); cursor: pointer;
  padding: .6rem .8rem; font-size: .83rem; font-weight: 700; white-space: nowrap; }
.empty-catalog { margin-top: 1rem; min-height: 180px; padding: 1.5rem; border-radius: var(--radius-xl);
  background: var(--color-surface-muted); border: 1px dashed var(--color-border);
  display: flex; flex-direction: column; align-items: flex-start; justify-content: center; gap: .65rem; }
.empty-catalog strong { font-size: .97rem; }
.empty-catalog span { font-size: .86rem; color: var(--color-text-muted); }
.sr-only { position:absolute; width:1px; height:1px; padding:0; margin:-1px;
  overflow:hidden; clip:rect(0,0,0,0); white-space:nowrap; border:0; }

/* Actividades y acciones */
.activity-card:hover { border-color: var(--color-border); }
.activity-card.is-editing { border-color: var(--color-primary); box-shadow: var(--shadow-focus); }
.activity-card.is-cancelled { background: var(--color-surface-muted); }
.activity-card h3 { font-size: 1.05rem; line-height: 1.4; }
.status-cancelled { background: var(--color-error-soft); color: var(--color-error); }
.status-completed { background: var(--color-surface-soft); color: var(--color-text-muted); }
.status-draft { background: var(--color-warning-soft); color: var(--color-warning); }
.resource-name { font-size: .89rem; }
.activity-description { color: var(--color-text-muted); line-height: 1.5; font-size: .88rem; }
.schedule-row { font-size: .83rem; line-height: 1.4; gap: .65rem; }
.schedule-row strong { color: var(--color-primary-strong); }
.creator-note { display: block; padding-top: .65rem; border-top: 1px solid var(--color-border-soft); }
.enrollment-warning { font-size: .78rem; color: var(--color-warning);
  background: var(--color-warning-soft); border-radius: var(--radius-md); padding: .65rem; margin-top: .8rem; }
.secondary-button, .danger-button { font-weight: 650; font-size: .83rem; }
.secondary-button { background: var(--color-surface); color: var(--color-text);
  border-color: var(--color-border); }
.secondary-button:disabled { opacity: .5; cursor: not-allowed; }
.danger-button { color: var(--color-error); border-color: var(--color-error-border);
  background: var(--color-surface); }
.cancel-confirm { border: 1px solid var(--color-error-border);
  background: var(--color-error-soft); color: var(--color-text); display: block; }
.cancel-confirm strong { color: var(--color-error); font-size: .9rem; }
.cancel-confirm p { padding: .55rem 0 .8rem; font-size: .83rem; line-height: 1.5; }
.cancel-actions { display: flex; flex-wrap: wrap; gap: .55rem; }
.cancel-actions .danger-button { background: var(--color-error); color: white; }

.field-hint {
  margin: .3rem 0 0;
  color: var(--color-text-muted);
  font-size: .82rem;
  font-weight: 600;
}

@media (max-width: 900px) {
  .workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 650px) {
  .institutional-page { padding: .85rem; }
  .activity-overview { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .overview-caption { grid-column: 1 / -1; }
  .section-heading { flex-direction: column; align-items: stretch; }
  .create-shortcut { align-self: start; }
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
