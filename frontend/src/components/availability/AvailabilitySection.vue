<script setup>
import { ref, computed, onMounted } from 'vue'

import CalendarToolbar from './CalendarToolbar.vue'
import CalendarMini from './CalendarMini.vue'
import ScheduleGrid from './ScheduleGrid.vue'
import GeneralCalendarView from './GeneralCalendarView.vue'
import AvailabilityTypeLegend from './AvailabilityTypeLegend.vue'
import ReservationDetailModal from './ReservationDetailModal.vue'
import ReservationForm from '../forms/ReservationForm.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import WorkshopWithdrawalConfirm from '@/components/workshops/WorkshopWithdrawalConfirm.vue'

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import { useAuthStore } from '@/stores/auth'
import { useActivitiesStore } from '@/stores/activities'
import { useWorkshopsStore } from '@/stores/workshops'
import { reservationsService } from '@/services/reservations.service'
import { buildWorkshopAvailabilityItems } from '@/utils/workshopSchedule'
import { hasRut } from '@/utils/validators'
import {
  canUserEditReservationTarget,
  getBusinessDateKey,
  parseReservationDateTime
} from '@/utils/reservationTime'
import {
  RESERVATION_CLOSING_HOUR,
  RESERVATION_OPENING_HOUR
} from '@/utils/reservationRules'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()
const authStore = useAuthStore()
const activitiesStore = useActivitiesStore()
const workshopsStore = useWorkshopsStore()

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

const getEndDateTime = (date, hour, durationMinutes) => {
  const start = buildLocalDateTime(date, hour)

  return new Date(
    start.getTime() + Number(durationMinutes || 0) * 60000
  )
}

const hasAvailabilityConflict = ({
  resourceId,
  date,
  hour,
  durationMinutes
}) => {
  const resource = resourcesStore.resources.find(
    item => item.id === resourceId
  )

  if (resource?.reservationMode === 'OPEN_USE') {
    return false
  }

  const nextStart = buildLocalDateTime(date, hour)
  const nextEnd = getEndDateTime(date, hour, durationMinutes)

  return availabilityItems.value.some((item) => {
    if (
      item.resourceId !== resourceId ||
      item.status === 'CANCELLED'
    ) {
      return false
    }

    const itemStart = parseReservationDateTime(item.startTime)

    if (!itemStart) {
      return false
    }

    const itemEnd = new Date(
      itemStart.getTime() +
      Number(item.durationMinutes || 0) * 60000
    )

    return nextStart < itemEnd && nextEnd > itemStart
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
    policyLoading.value ||
    authStore.loading ||
    (resourcesStore.initialLoading ??
      (resourcesStore.loading && !resourcesStore.hasLoaded)) ||
    (reservationsStore.availabilityInitialLoading ??
      (reservationsStore.availabilityLoading &&
        !reservationsStore.availabilityHasLoaded)) ||
    (workshopsStore.initialLoading ??
      (workshopsStore.loading && !workshopsStore.hasLoaded)) ||
    (activitiesStore.initialLoading ??
      (activitiesStore.loading && !activitiesStore.hasLoaded))
  )
})

const loadWarning = computed(() => {
  return (
    resourcesStore.error ||
    reservationsStore.availabilityLoadingError ||
    activitiesStore.error ||
    workshopsStore.loadingError ||
    policyError.value ||
    ''
  )
})

const availabilityItems = computed(() => {
  return [
    ...reservationsStore.availabilityReservations,
    ...buildWorkshopAvailabilityItems({
      workshops: workshopsStore.workshops,
      resources: resourcesStore.resources,
      selectedDate: selectedDate.value
    })
  ]
})

const reservationBlockingError = computed(() => {
  if (
    authStore.user &&
    authStore.user.isAdmin !== true &&
    !hasRut(authStore.user.rut)
  ) {
    return 'Debes registrar tu RUT antes de crear reservas.'
  }

  if (resourcesStore.error) {
    return 'No se puede crear la reserva porque no se pudieron cargar los recursos.'
  }

  if (reservationsStore.availabilityLoadingError) {
    return 'No se puede crear la reserva porque no se pudo validar la disponibilidad actual.'
  }

  return ''
})

/* SELECTED SLOT */
const selectedSlot = ref(null)

/* SELECTED RESERVATION */
const selectedReservation = ref(null)
const pendingWorkshopWithdrawal = ref(null)

const selectedWorkshopId = computed(() => {
  if (!selectedReservation.value?.isWorkshop) {
    return null
  }

  const nestedId = selectedReservation.value?.workshop?.id
  const directId = selectedReservation.value?.workshopId
  const availabilityId = String(selectedReservation.value?.id || '')
    .match(/^workshop-(\d+)-/)?.[1]
  const workshopId = Number(nestedId ?? directId ?? availabilityId)

  if (!Number.isInteger(workshopId) || workshopId <= 0) {
    return null
  }

  return workshopId
})
const currentWorkshop = computed(() => {
  if (!selectedWorkshopId.value) {
    return null
  }

  return workshopsStore.workshops.find(
    (workshop) => Number(workshop.id) === selectedWorkshopId.value
  ) || selectedReservation.value.workshop
})
const hasConfirmedSelectedWorkshopEnrollment = computed(() => {
  if (currentWorkshop.value?.isEnrolled === true) {
    return true
  }

  return workshopsStore.myEnrollments?.some(
    (enrollment) => (
      Number(enrollment.workshopId) === selectedWorkshopId.value &&
      String(enrollment.status || '').toUpperCase() === 'CONFIRMED'
    )
  ) === true
})
const canWithdrawSelectedWorkshop = computed(() => Boolean(
  authStore.user &&
  selectedReservation.value?.isWorkshop &&
  currentWorkshop.value?.isActive !== false &&
  hasConfirmedSelectedWorkshopEnrollment.value
))
const canEnrollSelectedWorkshop = computed(() => Boolean(
  authStore.user &&
  selectedReservation.value?.isWorkshop &&
  currentWorkshop.value?.isActive !== false &&
  !hasConfirmedSelectedWorkshopEnrollment.value &&
  Number(currentWorkshop.value?.capacity || 0) >
    Number(currentWorkshop.value?.enrolledCount || 0)
))
const selectedWorkshopEnrolling = computed(() => (
  workshopsStore.enrollingId !== null
))
const selectedWorkshopWithdrawing = computed(() => (
  Number(workshopsStore.withdrawingId) === selectedWorkshopId.value
))
const selectedDetailError = computed(() => (
  selectedReservation.value?.isWorkshop
    ? workshopsStore.actionError?.message || ''
    : reservationsStore.actionError || ''
))
const selectedWorkshopActionMessage = computed(() => (
  selectedReservation.value?.isWorkshop
    ? workshopsStore.actionSuccess || ''
    : ''
))

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
const canManageSelectedJoinCode = computed(() =>
  Boolean(selectedReservation.value?.targetParticipants) &&
  selectedReservation.value?.userId === authStore.user?.id
)
const canEditSelectedTarget = computed(() =>
  canUserEditReservationTarget(selectedReservation.value, authStore.user)
)

/* MODAL */
const showReservationForm = ref(false)
const isCreatingReservation = ref(false)
const currentPolicy = ref(null)
const policyLoading = ref(true)
const policyError = ref('')

const loadCurrentPolicy = async () => {
  policyLoading.value = true
  policyError.value = ''

  try {
    const policy = await reservationsService.getCurrentPolicy()

    if (!policy || !Array.isArray(policy.groupResourceIds)) {
      throw new Error('invalid policy')
    }

    currentPolicy.value = policy
    return policy
  } catch {
    currentPolicy.value = null
    policyError.value = 'No se pudo validar la política de reservas. Por seguridad, la creación de reservas está temporalmente deshabilitada.'
    return null
  } finally {
    policyLoading.value = false
  }
}

/* LOAD DATA */
onMounted(async () => {
  await Promise.all([
    authStore.loadAuthUser(),
    resourcesStore.fetchResources(),
    reservationsStore.fetchAvailabilityReservations(),
    activitiesStore.fetchActivities(),
    workshopsStore.fetchWorkshops(),
    workshopsStore.fetchMyEnrollments(),
    loadCurrentPolicy()
  ])
})

/* SLOT SELECT */
const handleSlotSelected = (slot) => {
  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()

  if (policyLoading.value || !currentPolicy.value) {
    reservationsStore.setActionError(
      policyError.value || 'Espera mientras se valida la política de reservas.'
    )
    return
  }

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
  workshopsStore.clearMessages?.()

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
  pendingWorkshopWithdrawal.value = null

  reservationsStore.clearActionError?.()
  workshopsStore.clearMessages?.()
}

const requestSelectedWorkshopWithdrawal = () => {
  if (!canWithdrawSelectedWorkshop.value || selectedWorkshopWithdrawing.value) {
    return
  }

  pendingWorkshopWithdrawal.value = currentWorkshop.value
}

const enrollSelectedWorkshop = async () => {
  const workshop = currentWorkshop.value

  if (
    !canEnrollSelectedWorkshop.value ||
    workshopsStore.enrollingId !== null
  ) {
    return
  }

  try {
    const updatedWorkshop = await workshopsStore.enroll(workshop.id)

    if (selectedReservation.value?.isWorkshop && updatedWorkshop) {
      selectedReservation.value = {
        ...selectedReservation.value,
        workshop: {
          ...selectedReservation.value.workshop,
          ...updatedWorkshop
        }
      }
    }
  } catch {
    // El store conserva el mensaje específico sin cerrar el detalle.
  }
}

const confirmSelectedWorkshopWithdrawal = async () => {
  const workshop = pendingWorkshopWithdrawal.value

  if (!workshop || workshopsStore.withdrawingId !== null) {
    return
  }

  try {
    const change = await workshopsStore.withdraw(workshop.id)

    if (selectedReservation.value?.isWorkshop && change) {
      selectedReservation.value = {
        ...selectedReservation.value,
        workshop: {
          ...selectedReservation.value.workshop,
          isEnrolled: change.isEnrolled,
          enrolledCount: change.enrolledCount
        }
      }
    }
  } catch {
    // El store mantiene un mensaje seguro dentro del detalle.
  } finally {
    pendingWorkshopWithdrawal.value = null
  }
}
const updateSelectedTarget = async (targetParticipants) => {
  if (!canEditSelectedTarget.value) {
    return
  }

  try {
    const updated = await reservationsService.updateTarget(selectedReservation.value.id, Number(targetParticipants))
    selectedReservation.value = { ...selectedReservation.value, ...updated }
    await reservationsStore.fetchAvailabilityReservations()
    reservationsStore.setActionSuccess('Objetivo de participantes actualizado.')
  } catch (error) {
    reservationsStore.setActionError(error.response?.data?.error || 'No se pudo actualizar el objetivo.')
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
        'Ese horario se cruza con una reserva o taller existente.'
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
    if (reservation.targetParticipants != null) {
      payload.targetParticipants = Number(reservation.targetParticipants)
    }

    await reservationsStore.createReservation(payload)

    showReservationForm.value = false
    selectedSlot.value = null

    reservationsStore.clearActionError?.()

    await reservationsStore.fetchAvailabilityReservations()

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

    await reservationsStore.fetchAvailabilityReservations()

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

  selectedDate.value = formatDateKey(date)

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

          <div class="view-controls">
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

            <AvailabilityTypeLegend
              :resources="resourcesStore.resources"
              :reservations="availabilityItems"
              :selected-date="selectedDate"
            />
          </div>

          <ScheduleGrid
            v-if="viewMode === 'resources'"
            :resources="resourcesStore.resources"
            :reservations="availabilityItems"
            :selected-date="selectedDate"
            :start-hour="RESERVATION_OPENING_HOUR"
            :end-hour="RESERVATION_CLOSING_HOUR"
            :pixels-per-minute="1"
            @slot-selected="handleSlotSelected"
            @reservation-selected="handleReservationSelected"
          />

          <GeneralCalendarView
            v-else
            :resources="resourcesStore.resources"
            :reservations="availabilityItems"
            :selected-date="selectedDate"
            @reservation-selected="handleReservationSelected"
          />

        </div>

      </div>

    </template>

    <!-- FORM -->
    <ReservationForm
      :visible="showReservationForm"
      :slot="selectedSlot"
      :resources="resourcesStore.resources"
      :activities="activitiesStore.activities"
      :error-message="reservationsStore.actionError"
      :submitting="isCreatingReservation"
      :policy="currentPolicy"
      :is-admin="authStore.user?.isAdmin === true"
      @close="closeReservationForm"
      @submit="submitReservation"
    />

    <!-- DETAIL -->
    <ReservationDetailModal
      :visible="Boolean(selectedReservation)"
      :reservation="selectedReservation"
      :can-edit-target="canEditSelectedTarget"
      :can-manage-join-code="canManageSelectedJoinCode"
      :can-cancel="canCancelSelectedReservation"
      :error-message="selectedDetailError"
      :can-enroll-workshop="canEnrollSelectedWorkshop"
      :workshop-enrolling="selectedWorkshopEnrolling"
      :can-withdraw-workshop="canWithdrawSelectedWorkshop"
      :workshop-withdrawing="selectedWorkshopWithdrawing"
      :workshop-action-message="selectedWorkshopActionMessage"
      @close="closeReservationDetail"
      @cancel="cancelSelectedReservation"
      @update-target="updateSelectedTarget"
      @enroll-workshop="enrollSelectedWorkshop"
      @withdraw-workshop="requestSelectedWorkshopWithdrawal"
    />

    <WorkshopWithdrawalConfirm
      :workshop="pendingWorkshopWithdrawal"
      :loading="selectedWorkshopWithdrawing"
      @confirm="confirmSelectedWorkshopWithdrawal"
      @cancel="pendingWorkshopWithdrawal = null"
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

.view-controls {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.view-switch {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 4px;
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

  .view-controls {
    align-items: stretch;
    flex-direction: column;
    gap: 10px;
  }

  .section-header h2 {
    font-size: 24px;
  }

  .section-header p {
    font-size: 14px;
  }
}
</style>
