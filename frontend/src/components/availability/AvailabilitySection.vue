<script setup>
import { ref, computed, onMounted } from 'vue'

import CalendarToolbar from './CalendarToolbar.vue'
import CalendarMini from './CalendarMini.vue'
import ScheduleGrid from './ScheduleGrid.vue'
import ReservationDetailModal from './ReservationDetailModal.vue'
import ReservationForm from '../forms/ReservationForm.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import { useAuthStore } from '@/stores/auth'
import { useActivitiesStore } from '@/stores/activities'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()
const authStore = useAuthStore()
const activitiesStore = useActivitiesStore()

/* DATE */
const selectedDate = ref(
  new Date().toISOString().slice(0, 10)
)

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
    reservationsStore.loading
  )
})

const loadWarning = computed(() => {
  return (
    resourcesStore.error ||
    reservationsStore.loadingError ||
    activitiesStore.error ||
    ''
  )
})

const reservationBlockingError = computed(() => {
  if (resourcesStore.error) {
    return 'No se puede crear la reserva porque no se pudieron cargar los recursos.'
  }

  if (reservationsStore.loadingError) {
    return 'No se puede crear la reserva porque no se pudo validar la disponibilidad actual.'
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
    reservationsStore.fetchReservations(),
    activitiesStore.fetchActivities()
  ])
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

    const activityId =
      Number(reservation.activityId)

    if (activityId > 0) {
      payload.activityId = activityId
    }

    await reservationsStore.createReservation(payload)

    showReservationForm.value = false
    selectedSlot.value = null

    reservationsStore.clearActionError?.()

    await reservationsStore.fetchReservations()

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

    await reservationsStore.fetchReservations()

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

  selectedDate.value =
    `${date.year}-${month}-${day}`

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

/* TOOLBAR */
const previousDay = () => {
  const date =
    new Date(`${selectedDate.value}T00:00:00`)

  date.setDate(date.getDate() - 1)

  selectedDate.value =
    date.toISOString().slice(0, 10)

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

const nextDay = () => {
  const date =
    new Date(`${selectedDate.value}T00:00:00`)

  date.setDate(date.getDate() + 1)

  selectedDate.value =
    date.toISOString().slice(0, 10)

  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()
}

const goToday = () => {
  selectedDate.value =
    new Date().toISOString().slice(0, 10)

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

          <ScheduleGrid
            :resources="resourcesStore.resources"
            :reservations="reservationsStore.reservations"
            :selected-date="selectedDate"
            :start-hour="8"
            :end-hour="22"
            :pixels-per-minute="1"
            @slot-selected="handleSlotSelected"
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

  color: #0f172a;
}

.section-header p {
  margin-top: 6px;

  font-size: 15px;

  color: #64748b;
}

/* STATE */
.state-card {
  background: white;

  border-radius: 22px;

  padding: 24px;

  border: 1px solid #e2e8f0;

  color: #334155;

  font-weight: 600;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.state-card.success {
  background: #dcfce7;

  color: #166534;

  border-color: #bbf7d0;
}

.state-card.warning {
  background: #fef3c7;

  color: #92400e;

  border-color: #fde68a;
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

/* CARD */
.selection-card {
  background: white;

  border-radius: 22px;

  padding: 20px;

  border: 1px solid #e2e8f0;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.selection-card h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 700;

  color: #0f172a;
}

.selection-card p {
  margin: 14px 0 6px;

  color: #334155;
}

.selection-card span {
  color: #2563eb;

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
