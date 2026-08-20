<script setup>
import { ref, computed, nextTick, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import CalendarToolbar from './CalendarToolbar.vue'
import CalendarMini from './CalendarMini.vue'
import ScheduleGrid from './ScheduleGrid.vue'
import GeneralCalendarView from './GeneralCalendarView.vue'
import ReservationDetailModal from './ReservationDetailModal.vue'
import ReservationForm from '../forms/ReservationForm.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import { useAuthStore } from '@/stores/auth'
import { useActivitiesStore } from '@/stores/activities'
import { useReservationPolicyStore } from '@/stores/reservationPolicy'
import {
  getBusinessDateKey,
  parseReservationDateTime
} from '@/utils/reservationTime'
import {
  availabilityRangeForDate,
  getEligibleResources,
  hasReservationConflict,
  isDateWithinPolicyWindow
} from '@/utils/availabilityRules'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()
const authStore = useAuthStore()
const activitiesStore = useActivitiesStore()
const policyStore = useReservationPolicyStore()

const route = useRoute()
const router = useRouter()

const formatDateKey = (date) => {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0')
  ].join('-')
}

const todayKey = () => {
  return getBusinessDateKey()
}

const buildLocalDateTime = (date, hour) => {
  return parseReservationDateTime(`${date}T${hour}:00`)
}

const isPastStart = (date, hour) => {
  const start = buildLocalDateTime(date, hour)

  return start.getTime() <= Date.now()
}

const hasAvailabilityConflict = ({
  resourceId,
  date,
  hour,
  durationMinutes
}) => {
  const resource = eligibleResources.value.find(
    item => item.id === resourceId
  )
  const start = buildLocalDateTime(date, hour)
  const end = new Date(start.getTime() + Number(durationMinutes) * 60000)

  return hasReservationConflict({
    items: availabilityItems.value,
    resource,
    userId: authStore.user?.id,
    start,
    end
  })
}

/* DATE */
const selectedDate = ref(
  todayKey()
)

const viewMode = ref('resources')

const currentDateLabel = computed(() => {
  const date = new Date(`${selectedDate.value}T00:00:00`)

  return date.toLocaleDateString('es-CL', {
    weekday: 'long',
    day: 'numeric',
    month: 'long'
  })
})

const isLoadingAvailability = computed(() => {
  return (
    authStore.loading ||
    resourcesStore.loading ||
    reservationsStore.availabilityLoading ||
    policyStore.loading
  )
})

const loadWarning = computed(() => {
  return (
    resourcesStore.error ||
    reservationsStore.availabilityLoadingError ||
    activitiesStore.error ||
    policyStore.error ||
    ''
  )
})

const availabilityItems = computed(() => {
  return reservationsStore.availabilityReservations
})

const policy = computed(() => policyStore.policy)

const eligibleResources = computed(() => {
  return getEligibleResources(
    resourcesStore.resources,
    policy.value,
    authStore.user?.isAdmin === true
  )
})

const readRouteQueryValue = (value) => {
  if (Array.isArray(value)) {
    return String(value[0] || '')
  }

  return value === null || value === undefined
    ? ''
    : String(value)
}

const resourceFilterOptions = computed(() => {
  return resourcesStore.resources
})

const selectedResourceId = computed(() => {
  const requested =
    readRouteQueryValue(route.query.resource)

  if (!requested) {
    return ''
  }

  const exists = resourceFilterOptions.value.some(
    resource => String(resource.id) === requested
  )

  return exists ? requested : ''
})

const selectedResource = computed(() => {
  if (!selectedResourceId.value) {
    return null
  }

  return resourceFilterOptions.value.find(
    resource =>
      String(resource.id) ===
      selectedResourceId.value
  ) || null
})

const onlyAvailable = computed(() => {
  return (
    readRouteQueryValue(route.query.available) ===
    'true'
  )
})

const selectedResourceIsEligible = computed(() => {
  if (!selectedResource.value) {
    return false
  }

  return eligibleResources.value.some(
    resource =>
      Number(resource.id) ===
      Number(selectedResource.value.id)
  )
})

const formatMinuteToHour = (minuteOfDay) => {
  const hour =
    Math.floor(minuteOfDay / 60)

  const minute =
    minuteOfDay % 60

  return (
    `${String(hour).padStart(2, '0')}:` +
    `${String(minute).padStart(2, '0')}`
  )
}

const resourceHasAvailableBlock = (resource) => {
  if (!policy.value) {
    return false
  }

  const durations =
    (policy.value.allowedDurations || [])
      .map(Number)
      .filter(duration => duration > 0)

  if (durations.length === 0) {
    return false
  }

  const minimumDuration =
    Math.min(...durations)

  const openingMinute =
    Number(policy.value.openingMinute)

  const closingMinute =
    Number(policy.value.closingMinute)

  const intervalMinutes =
    Number(policy.value.slotIntervalMinutes) || 15

  for (
    let minute = openingMinute;
    minute + minimumDuration <= closingMinute;
    minute += intervalMinutes
  ) {
    const hour =
      formatMinuteToHour(minute)

    const start =
      buildLocalDateTime(
        selectedDate.value,
        hour
      )

    if (!start) {
      continue
    }

    if (start.getTime() <= Date.now()) {
      continue
    }

    const end = new Date(
      start.getTime() +
      minimumDuration * 60000
    )

    const conflict =
      hasReservationConflict({
        items: availabilityItems.value,
        resource,
        userId: authStore.user?.id,
        start,
        end
      })

    if (!conflict) {
      return true
    }
  }

  return false
}

const filteredResources = computed(() => {
  let resources =
    eligibleResources.value.slice()

  if (selectedResourceId.value) {
    resources = resources.filter(
      resource =>
        String(resource.id) ===
        selectedResourceId.value
    )
  }

  if (onlyAvailable.value) {
    resources = resources.filter(
      resource =>
        resourceHasAvailableBlock(resource)
    )
  }

  return resources
})

const resourceFilterEmptyMessage = computed(() => {
  if (
    selectedResource.value &&
    !selectedResourceIsEligible.value
  ) {
    return (
      `${selectedResource.value.name} no está habilitado ` +
      'para reservas particulares.'
    )
  }

  if (
    selectedResource.value &&
    onlyAvailable.value
  ) {
    return (
      `${selectedResource.value.name} no tiene bloques ` +
      'disponibles para la fecha seleccionada.'
    )
  }

  if (onlyAvailable.value) {
    return (
      'No hay recursos con bloques disponibles ' +
      'para la fecha seleccionada.'
    )
  }

  return 'No hay recursos disponibles.'
})

const replaceAvailabilityQuery = async ({
  resourceId = selectedResourceId.value,
  available = onlyAvailable.value
} = {}) => {
  const query = {
    ...route.query
  }

  if (resourceId) {
    query.resource = String(resourceId)
  } else {
    delete query.resource
  }

  if (available) {
    query.available = 'true'
  } else {
    delete query.available
  }

  await router.replace({
    query
  })
}

const handleResourceFilterChange = (event) => {
  void replaceAvailabilityQuery({
    resourceId: event.target.value,
    available: onlyAvailable.value
  })
}

const handleAvailabilityFilterChange = (event) => {
  void replaceAvailabilityQuery({
    resourceId: selectedResourceId.value,
    available:
      event.target.value === 'available'
  })
}

const normalizeResourceQuery = async () => {
  const requested =
    readRouteQueryValue(route.query.resource)

  if (
    !requested ||
    resourcesStore.loading ||
    resourcesStore.resources.length === 0
  ) {
    return
  }

  const exists =
    resourcesStore.resources.some(
      resource =>
        String(resource.id) === requested
    )

  if (!exists) {
    await replaceAvailabilityQuery({
      resourceId: '',
      available: onlyAvailable.value
    })
  }
}

const availabilityIsCurrent = computed(() => {
  return reservationsStore.availabilityRangeKey === selectedDate.value
})

const reservationBlockingError = computed(() => {
  if (
    !authStore.user
  ) {
    return 'No se pudo validar la identidad local.'
  }

  if (
    authStore.user &&
    authStore.user.isAdmin !== true &&
    !authStore.user.rut
  ) {
    return 'Debes registrar tu RUT antes de crear reservas.'
  }

  if (resourcesStore.error) {
    return 'No se puede crear la reserva porque no se pudieron cargar los recursos.'
  }

  if (!policy.value || policyStore.error) {
    return 'No se puede crear la reserva porque no se pudo cargar la política vigente.'
  }

  if (
    reservationsStore.availabilityLoading ||
    reservationsStore.availabilityLoadingError ||
    !availabilityIsCurrent.value
  ) {
    return 'No se puede crear la reserva porque no se pudo validar la disponibilidad actual.'
  }

  if (eligibleResources.value.length === 0) {
    return 'No hay recursos habilitados por la política vigente.'
  }

  return ''
})

/* SELECTED SLOT */
const selectedSlot = ref(null)

/* SELECTED RESERVATION */
const selectedReservation = ref(null)

const canCancelSelectedReservation = computed(() => {
  if (!authStore.user || !selectedReservation.value) {
    return false
  }

  if (
    selectedReservation.value.isWorkshop ||
    selectedReservation.value.isScheduledActivity
  ) {
    return false
  }

  return (
    authStore.user.isAdmin === true ||
    selectedReservation.value.userId === authStore.user.id
  )
})

/* MODAL */
const showReservationForm = ref(false)
const isCreatingReservation = ref(false)
const focusReservationId = ref(null)
const createdReservationResult = ref(null)

/* LOAD DATA */
onMounted(async () => {
  await Promise.all([
    authStore.loadAuthUser(),
    resourcesStore.fetchResources(),
    activitiesStore.fetchActivities(),
    policyStore.fetchCurrentPolicy()
  ])

  await normalizeResourceQuery()
  await loadAvailabilityForDate(selectedDate.value)
})

const loadAvailabilityForDate = async (dateKey) => {
  const range = availabilityRangeForDate(dateKey)
  return reservationsStore.fetchAvailabilityReservations(range)
}

watch(selectedDate, async (dateKey, previousDate) => {
  if (!previousDate || dateKey === previousDate) {
    return
  }

  selectedSlot.value = null
  createdReservationResult.value = null
  showReservationForm.value = false
  await loadAvailabilityForDate(dateKey)
})

watch(
  () => route.query.resource,
  () => {
    void normalizeResourceQuery()
  }
)

/* SLOT SELECT */
const handleSlotSelected = (slot) => {
  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()

  if (reservationBlockingError.value) {
    reservationsStore.setActionError(
      reservationBlockingError.value
    )

    return
  }

  if (isPastStart(selectedDate.value, slot.hour)) {
    reservationsStore.setActionError(
      'No puedes crear reservas en fechas u horarios pasados.'
    )

    return
  }

  selectedReservation.value = null
  createdReservationResult.value = null

  selectedSlot.value = {
    resource: slot.resource,
    hour: slot.hour,
    date: selectedDate.value
  }

  showReservationForm.value = true
}

/* RESERVATION SELECT */
const handleReservationSelected = (reservation) => {
  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()

  selectedSlot.value = null
  createdReservationResult.value = null
  showReservationForm.value = false
  selectedReservation.value = reservation
}

/* CLOSE */
const closeReservationForm = () => {
  showReservationForm.value = false
  selectedSlot.value = null
  createdReservationResult.value = null

  reservationsStore.clearActionError?.()
}

const closeReservationDetail = () => {
  selectedReservation.value = null

  reservationsStore.clearActionError?.()
}

const clearReservationFocus = (reservationId) => {
  if (String(focusReservationId.value) === String(reservationId)) {
    focusReservationId.value = null
  }
}

/* SUBMIT */
const submitReservation = async (reservation) => {
  if (isCreatingReservation.value) {
    return
  }

  isCreatingReservation.value = true

  try {
    reservationsStore.clearActionSuccess?.()

    if (!authStore.user) {
      await authStore.loadAuthUser()
    }

    if (!authStore.user) {
      reservationsStore.setActionError(
        'No se pudo identificar al usuario autenticado'
      )

      return
    }

    if (reservationBlockingError.value) {
      reservationsStore.setActionError(
        reservationBlockingError.value
      )

      return
    }

    if (!reservation.resource?.id) {
      reservationsStore.setActionError(
        'No se pudo identificar el recurso seleccionado'
      )

      return
    }

    if (isPastStart(reservation.date, reservation.hour)) {
      reservationsStore.setActionError(
        'No puedes crear reservas en fechas u horarios pasados.'
      )

      return
    }

    if (
      hasAvailabilityConflict({
        resourceId: reservation.resource.id,
        date: reservation.date,
        hour: reservation.hour,
        durationMinutes: reservation.durationMinutes
      })
    ) {
      reservationsStore.setActionError(
        'Ese horario se cruza con una ocupación existente.'
      )

      return
    }

    const activityId =
      Number(reservation.activityId)

    const payload = {
      resourceId:
        reservation.resource.id,

      startTime:
        buildStartTime(
          reservation.date,
          reservation.hour
        ),

      durationMinutes:
        Number(reservation.durationMinutes)
    }

    if (activityId > 0) {
      payload.activityId = activityId
    }

    const createdReservation =
      await reservationsStore.createReservation(payload)

    createdReservationResult.value = createdReservation
    viewMode.value = 'resources'

    reservationsStore.clearActionError?.()

    await loadAvailabilityForDate(selectedDate.value)
    await nextTick()

    focusReservationId.value =
      createdReservation?.id || null
  } catch {
    // El store mantiene el error visible dentro del formulario.
  } finally {
    isCreatingReservation.value = false
  }
}

const cancelSelectedReservation = async () => {
  try {
    reservationsStore.clearActionSuccess?.()

    if (!selectedReservation.value?.id) {
      reservationsStore.setActionError(
        'No se pudo identificar la reserva seleccionada'
      )

      return
    }

    if (!canCancelSelectedReservation.value) {
      reservationsStore.setActionError(
        'No tienes permisos para cancelar esta reserva'
      )

      return
    }

    await reservationsStore.cancelReservation(
      selectedReservation.value.id
    )

    selectedReservation.value = null

    await loadAvailabilityForDate(selectedDate.value)

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store mantiene el error visible dentro del modal.
  }
}

/* BUILD DATETIME */
const buildStartTime = (date, hour) => {
  return `${date}T${hour}:00`
}

/* CALENDAR */
const handleDateSelect = (date) => {
  const month =
    String(date.month + 1).padStart(2, '0')

  const day =
    String(date.day).padStart(2, '0')

  const nextDate =
    `${date.year}-${month}-${day}`

  if (nextDate < todayKey()) {
    reservationsStore.setActionError(
      'No puedes seleccionar una fecha pasada.'
    )

    return
  }

  if (
    policy.value &&
    !isDateWithinPolicyWindow(
      nextDate,
      todayKey(),
      policy.value.reservableWindowDays
    )
  ) {
    reservationsStore.setActionError(
      'La fecha seleccionada está fuera de la ventana de reservas vigente.'
    )
    return
  }

  selectedDate.value = nextDate

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

/* TOOLBAR */
const previousDay = () => {
  const date =
    new Date(`${selectedDate.value}T00:00:00`)

  date.setDate(date.getDate() - 1)

  const nextDate = formatDateKey(date)

  if (nextDate < todayKey()) {
    selectedDate.value = todayKey()
    return
  }

  selectedDate.value = nextDate

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

const nextDay = () => {
  const date =
    new Date(`${selectedDate.value}T00:00:00`)

  date.setDate(date.getDate() + 1)

  const nextDate = formatDateKey(date)

  if (
    policy.value &&
    !isDateWithinPolicyWindow(
      nextDate,
      todayKey(),
      policy.value.reservableWindowDays
    )
  ) {
    reservationsStore.setActionError(
      'Llegaste al último día disponible para reservar.'
    )
    return
  }

  selectedDate.value = nextDate

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

const goToday = () => {
  selectedDate.value =
    todayKey()

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}
</script>

<template>
  <section class="availability-section">

    <!-- HEADER -->
    <div class="section-header">

      <div>

        <h2>
          Disponibilidad
        </h2>

        <p>
          Revisa horarios y selecciona cualquier punto de la línea de tiempo.
        </p>

      </div>

    </div>

    <!-- LOADING -->
    <div
      v-if="isLoadingAvailability"
      aria-label="Cargando disponibilidad"
    >
      <SkeletonLoader variant="availability" />
    </div>

    <template v-else>

      <!-- WARNING -->
      <div
        v-if="loadWarning"
        class="state-card warning"
        role="status"
      >
        {{ loadWarning }}
      </div>

      <!-- SUCCESS -->
      <div
        v-if="reservationsStore.actionSuccess"
        class="state-card success"
        role="status"
        aria-live="polite"
      >
        {{ reservationsStore.actionSuccess }}
      </div>
      <!-- ACTION ERROR -->
      <div
        v-if="!showReservationForm && reservationsStore.actionError"
        class="state-card error"
        role="alert"
      >
        {{ reservationsStore.actionError }}
      </div>

      <!-- TOOLBAR -->
      <CalendarToolbar
        :current-date="currentDateLabel"
        @prev-day="previousDay"
        @next-day="nextDay"
        @today="goToday"
      />

      <!-- CONTENT -->
      <div class="content">

        <!-- LEFT -->
        <div class="calendar-container">

          <CalendarMini
            :selected-date="selectedDate"
            @select-date="handleDateSelect"
          />

          <!-- SLOT INFO -->
          <div
            v-if="selectedSlot"
            class="selection-card"
          >

            <h3>
              Selección actual
            </h3>

            <p>
              <strong>
                {{ selectedSlot.resource.name }}
              </strong>
            </p>

            <span>
              {{ selectedSlot.date }} · {{ selectedSlot.hour }}
            </span>

          </div>

        </div>

        <!-- RIGHT -->
        <div class="grid-container">

          <div
            class="view-switch"
            aria-label="Modo de vista"
          >
            <button
              type="button"
              :class="{ active: viewMode === 'resources' }"
              @click="viewMode = 'resources'"
            >
              Por recurso
            </button>

            <button
              type="button"
              :class="{ active: viewMode === 'general' }"
              @click="viewMode = 'general'"
            >
              Agenda del día
            </button>
          </div>

          <div
            v-if="viewMode === 'resources'"
            class="availability-filters"
            aria-label="Filtros de disponibilidad"
          >
            <label class="filter-field">
              <span>
                Recurso
              </span>

              <select
                :value="selectedResourceId"
                @change="handleResourceFilterChange"
              >
                <option value="">
                  Todos
                </option>

                <option
                  v-for="resource in resourceFilterOptions"
                  :key="resource.id"
                  :value="String(resource.id)"
                >
                  {{ resource.name }}
                </option>
              </select>
            </label>

            <label class="filter-field">
              <span>
                Mostrar
              </span>

              <select
                :value="onlyAvailable ? 'available' : 'all'"
                @change="handleAvailabilityFilterChange"
              >
                <option value="all">
                  Todos
                </option>

                <option value="available">
                  Con bloques disponibles
                </option>
              </select>
            </label>
          </div>

          <ScheduleGrid
            v-if="policy && viewMode === 'resources' && filteredResources.length > 0"
            :resources="filteredResources"
            :reservations="availabilityItems"
            :selected-date="selectedDate"
            :start-hour="Math.floor(policy.openingMinute / 60)"
            :end-hour="Math.ceil(policy.closingMinute / 60)"
            :pixels-per-minute="1"
            :current-user-id="authStore.user?.id"
            :slot-interval-minutes="policy.slotIntervalMinutes"
            :focus-reservation-id="focusReservationId"
            @slot-selected="handleSlotSelected"
            @reservation-selected="handleReservationSelected"
            @reservation-focused="clearReservationFocus"
          />

          <div
            v-else-if="policy && viewMode === 'resources'"
            class="filter-empty"
          >
            {{ resourceFilterEmptyMessage }}
          </div>

          <GeneralCalendarView
            v-else-if="policy"
            :resources="eligibleResources"
            :reservations="availabilityItems"
            :selected-date="selectedDate"
            @reservation-selected="handleReservationSelected"
          />

        </div>

      </div>

    </template>

    <!-- FORM -->
    <ReservationForm
      v-if="policy"
      :visible="showReservationForm"
      :slot="selectedSlot"
      :resources="eligibleResources"
      :activities="activitiesStore.activities"
      :policy="policy"
      :error-message="reservationsStore.actionError"
      :submitting="isCreatingReservation"
      :created-reservation="createdReservationResult"
      @close="closeReservationForm"
      @submit="submitReservation"
    />

    <!-- DETAIL -->
    <ReservationDetailModal
      :visible="Boolean(selectedReservation)"
      :reservation="selectedReservation"
      :can-cancel="canCancelSelectedReservation"
      :error-message="reservationsStore.actionError"
      @close="closeReservationDetail"
      @cancel="cancelSelectedReservation"
    />

  </section>
</template>

<style scoped>
.availability-section {
  display: flex;
  flex-direction: column;

  gap: 24px;
}
.created-group-code {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  padding: var(--space-4);

  border: 1px solid #bfdbfe;
  border-radius: var(--radius-lg);

  background: #eff6ff;
}

.created-group-code div {
  display: flex;
  flex-direction: column;

  gap: 5px;
}

.created-group-code strong {
  color: var(--color-text);
}

.created-group-code span {
  color: #1d4ed8;

  font-family: monospace;
  font-size: 21px;
  font-weight: 800;
  letter-spacing: 1px;
}

.created-group-code small {
  color: var(--color-text-muted);

  font-size: 13px;
}

.copy-created-code-button {
  flex-shrink: 0;

  border: none;
  border-radius: var(--radius-md);

  cursor: pointer;

  padding: 10px 14px;

  background: var(--color-primary);

  color: white;

  font-weight: 750;
}
/* HEADER */
.section-header h2 {
  margin: 0;

  font-size: 28px;
  font-weight: 800;

  color: var(--color-text);
}

.section-header p {
  margin-top: 6px;

  font-size: 15px;

  color: var(--color-text-muted);
}

/* STATE */
.state-card {
  border-radius: var(--radius-lg);
  padding: var(--space-5);
}

.state-card.error {
  background: var(--color-error-soft);
  color: var(--color-error);
  border-color: var(--color-error-border);
}

.state-card.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border-color: var(--color-success-border);
}

.state-card.warning {
  background: var(--color-warning-soft);
  color: var(--color-warning);
  border-color: var(--color-warning-border);
}

/* CONTENT */
.content {
  display: grid;

  grid-template-columns: 340px 1fr;

  gap: 24px;

  align-items: start;
}

/* LEFT */
.calendar-container {
  display: flex;
  flex-direction: column;

  gap: 20px;

  position: sticky;

  top: 90px;
}

/* RIGHT */
.grid-container {
  min-width: 0;
}

.availability-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;

  gap: 12px;

  margin-bottom: 16px;
}

.filter-field {
  display: flex;
  flex-direction: column;

  gap: 6px;

  min-width: 220px;
}

.filter-field span {
  color: var(--color-text-muted);

  font-size: 12px;
  font-weight: 800;
}

.filter-field select {
  min-height: 40px;

  padding: 0 12px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  background: var(--color-surface);
  color: var(--color-text);

  font: inherit;

  cursor: pointer;
}

.filter-field select:focus {
  border-color: var(--color-primary);
  outline: 2px solid var(--color-primary-soft);
  outline-offset: 1px;
}

.filter-empty {
  padding: var(--space-5);

  border: 1px dashed var(--color-border);
  border-radius: var(--radius-lg);

  background: var(--color-surface);
  color: var(--color-text-muted);

  font-weight: 700;
}

.view-switch {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 16px;
  padding: 4px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  box-shadow: var(--shadow-card);
}

.view-switch button {
  min-height: 34px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-muted);
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
}

.view-switch button:hover {
  color: var(--color-primary);
  background: var(--color-primary-soft);
}

.view-switch button.active {
  color: var(--color-primary-contrast);
  background: var(--color-primary);
  box-shadow: 0 6px 14px rgba(37, 99, 235, 0.18);
}

/* CARD */
.selection-card {
  background: var(--color-surface);

  border-radius: var(--radius-lg);

  padding: 20px;

  border: 1px solid var(--color-border);

  box-shadow: var(--shadow-card);
}

.selection-card h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 700;

  color: var(--color-text);
}

.selection-card p {
  margin: 14px 0 6px;

  color: #334155;
}

.selection-card span {
  color: var(--color-primary);

  font-size: 14px;
  font-weight: 700;
}

/* TABLET */
@media (max-width: 1024px) {
  .content {
    grid-template-columns: 1fr;
  }

  .calendar-container {
    position: relative;

    top: 0;
  }
}

/* MOBILE */
@media (max-width: 768px) {
  .availability-section {
    gap: 20px;
  }
  .created-group-code {
    align-items: stretch;
    flex-direction: column;
  }

  .copy-created-code-button {
    width: 100%;
  }
  .section-header h2 {
    font-size: 24px;
  }

  .section-header p {
    font-size: 14px;
  }
}
@media (max-width: 768px) {
  .availability-filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-field {
    width: 100%;
    min-width: 0;
  }
}
</style>
