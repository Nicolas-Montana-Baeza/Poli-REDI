<script setup>
import { ref, computed, onMounted, watch } from 'vue'

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

/* LOAD DATA */
onMounted(async () => {
  await Promise.all([
    authStore.loadAuthUser(),
    resourcesStore.fetchResources(),
    activitiesStore.fetchActivities(),
    policyStore.fetchCurrentPolicy()
  ])

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
  showReservationForm.value = false
  await loadAvailabilityForDate(dateKey)
})

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
  showReservationForm.value = false
  selectedReservation.value = reservation
}

/* CLOSE */
const closeReservationForm = () => {
  showReservationForm.value = false
  selectedSlot.value = null

  reservationsStore.clearActionError?.()
}

const closeReservationDetail = () => {
  selectedReservation.value = null

  reservationsStore.clearActionError?.()
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

    await reservationsStore.createReservation(payload)

    showReservationForm.value = false
    selectedSlot.value = null

    reservationsStore.clearActionError?.()

    await loadAvailabilityForDate(selectedDate.value)

    reservationsStore.setActionSuccess(
      'Reserva creada correctamente'
    )
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

          <ScheduleGrid
            v-if="policy && viewMode === 'resources'"
            :resources="eligibleResources"
            :reservations="availabilityItems"
            :selected-date="selectedDate"
            :start-hour="Math.floor(policy.openingMinute / 60)"
            :end-hour="Math.ceil(policy.closingMinute / 60)"
            :pixels-per-minute="1"
            :current-user-id="authStore.user?.id"
            :slot-interval-minutes="policy.slotIntervalMinutes"
            @slot-selected="handleSlotSelected"
            @reservation-selected="handleReservationSelected"
          />

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

  .section-header h2 {
    font-size: 24px;
  }

  .section-header p {
    font-size: 14px;
  }
}
</style>
