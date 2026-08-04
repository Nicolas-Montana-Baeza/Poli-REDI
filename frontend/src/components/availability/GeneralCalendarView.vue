<script setup>
import { computed } from 'vue'
import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import {
  formatReservationTimeRange,
  getReservationDateKey,
  getReservationDisplayStatus,
  getReservationStartMinutes
} from '@/utils/reservationTime'
import {
  getAvailabilityDisplayTitle,
  getAvailabilityType
} from '@/utils/availabilityType'

const props = defineProps({
  resources: {
    type: Array,
    default: () => []
  },

  reservations: {
    type: Array,
    default: () => []
  },

  selectedDate: {
    type: String,
    required: true
  }
})

const emit = defineEmits([
  'reservation-selected'
])

const dayReservations = computed(() => {
  return props.reservations
    .filter((reservation) => {
      return (
        reservation.status !== 'CANCELLED' &&
        getReservationDateKey(reservation.startTime) === props.selectedDate
      )
    })
    .sort((a, b) => {
      return (
        getReservationStartMinutes(a.startTime) -
        getReservationStartMinutes(b.startTime)
      )
    })
})

const reservedResourceCount = computed(() => {
  return new Set(
    dayReservations.value.map(
      reservation => reservation.resourceId
    )
  ).size
})

const totalMinutes = computed(() => {
  return dayReservations.value.reduce((total, reservation) => {
    return total + Number(reservation.durationMinutes || 0)
  }, 0)
})

const resourceName = (reservation) => {
  if (reservation.resourceName) {
    return reservation.resourceName
  }

  const resource = props.resources.find(
    item => item.id === reservation.resourceId
  )

  return resource?.name || 'Recurso'
}

const resourceFor = (reservation) => {
  return props.resources.find(
    item => String(item.id) === String(reservation.resourceId)
  ) || null
}

const typeFor = (reservation) => {
  return getAvailabilityType(
    reservation,
    resourceFor(reservation)
  )
}

const titleFor = (reservation) =>
  getAvailabilityDisplayTitle(reservation)

const statusFor = (reservation) => {
  if (reservation.isScheduledActivity) {
    return {
      label: 'Programada',
      className: 'confirmed'
    }
  }

  return getReservationDisplayStatus(reservation)
}

const accessibleLabel = (reservation) => {
  return [
    typeFor(reservation).label,
    titleFor(reservation),
    statusFor(reservation).label,
    formatReservationTimeRange(
      reservation.startTime,
      reservation.durationMinutes
    ),
    'Abrir detalle'
  ].filter(Boolean).join('. ')
}

const userName = (reservation) => {
  if (reservation.isScheduledActivity) {
    return 'Programación institucional'
  }

  return (
    reservation.userFullName ||
    reservation.userEmail ||
    'Ocupado'
  )
}

const selectReservation = (reservation) => {
  emit('reservation-selected', reservation)
}
</script>

<template>
  <section class="daily-agenda">

    <div class="agenda-header">
      <div>
        <h2>
          Agenda del día
        </h2>

        <p>
          Reservas y programación institucional ordenadas por horario.
        </p>
      </div>
    </div>

    <div class="summary-row">
      <div class="summary-item">
        <span>
          Bloques
        </span>

        <strong>
          {{ dayReservations.length }}
        </strong>
      </div>

      <div class="summary-item">
        <span>
          Recursos usados
        </span>

        <strong>
          {{ reservedResourceCount }}
        </strong>
      </div>

      <div class="summary-item">
        <span>
          Tiempo ocupado
        </span>

        <strong>
          {{ totalMinutes }} min
        </strong>
      </div>
    </div>

    <div
      v-if="dayReservations.length === 0"
      class="empty"
    >
      No hay reservas ni programación institucional para este día.
    </div>

    <div
      v-else
      class="agenda-list"
    >
      <button
        v-for="reservation in dayReservations"
        :key="reservation.availabilityKey || reservation.id"
        type="button"
        class="agenda-item"
        :aria-label="accessibleLabel(reservation)"
        @click="selectReservation(reservation)"
      >
        <div class="time-block">
          <strong>
            {{ formatReservationTimeRange(
              reservation.startTime,
              reservation.durationMinutes
            ) }}
          </strong>

          <span>
            {{ reservation.durationMinutes }} min
          </span>
        </div>

        <div class="reservation-main">
          <div class="reservation-title-row">
            <h3>
              {{ titleFor(reservation) }}
            </h3>

            <div class="reservation-indicators">
              <span
                class="status-pill"
                :class="statusFor(reservation).className"
              >
                {{ statusFor(reservation).label }}
              </span>

              <AvailabilityTypeChip
                :item="reservation"
                :resource="resourceFor(reservation)"
                aria-hidden
              />
            </div>
          </div>

          <div class="reservation-meta">
            <span>
              {{ resourceName(reservation) }}
            </span>

            <span>
              {{ userName(reservation) }}
            </span>

            <span v-if="reservation.userRut">
              {{ reservation.userRut }}
            </span>
          </div>
        </div>
      </button>
    </div>

  </section>
</template>

<style scoped>
.daily-agenda {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.agenda-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 800;
  color: var(--color-text);
}

.agenda-header p {
  margin-top: 4px;
  font-size: 14px;
  color: var(--color-text-muted);
}

.summary-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.summary-item {
  padding: 14px 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}

.summary-item span,
.summary-item strong {
  display: block;
}

.summary-item span {
  font-size: 12px;
  font-weight: 800;
  color: var(--color-text-muted);
}

.summary-item strong {
  margin-top: 6px;
  font-size: 22px;
  color: var(--color-text);
}

.empty {
  padding: var(--space-5);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-weight: 700;
}

.agenda-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.agenda-item {
  display: grid;
  grid-template-columns: 132px 1fr;
  gap: 16px;
  width: 100%;
  padding: 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
  text-align: left;
  cursor: pointer;
}

.agenda-item:hover {
  border-color: #b7cdf7;
  background: #fbfdff;
}

.time-block {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 58px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
}

.time-block strong {
  font-size: 13px;
  font-weight: 900;
}

.time-block span {
  margin-top: 4px;
  font-size: 12px;
  font-weight: 800;
}

.reservation-main {
  min-width: 0;
}

.reservation-title-row {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.reservation-title-row h3 {
  flex: 1 1 auto;
  margin: 0;
  min-width: 0;
  overflow: hidden;
  color: var(--color-text);
  font-size: 17px;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reservation-indicators {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
}

.reservation-indicators :deep(.availability-type-chip) {
  flex: 0 0 auto;
  max-width: none;
}

.status-pill {
  flex: 0 0 auto;
  padding: 6px 10px;
  border-radius: var(--radius-pill);
  font-size: 12px;
  font-weight: 900;
}

.status-pill.confirmed,
.status-pill.completed,
.status-pill.ongoing {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.status-pill.pending {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.status-pill.scheduled {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.status-pill.cancelled,
.status-pill.rejected,
.status-pill.expired {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.reservation-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.reservation-meta span {
  padding: 5px 9px;
  border-radius: var(--radius-pill);
  background: var(--color-surface-soft);
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 800;
}

@media (max-width: 768px) {
  .summary-row,
  .agenda-item {
    grid-template-columns: 1fr;
  }

  .reservation-title-row {
    align-items: center;
  }

  .reservation-indicators {
    gap: 6px;
  }
}

@media (max-width: 480px) {
  .reservation-title-row {
    align-items: flex-start;
  }

  .reservation-indicators {
    flex-direction: column-reverse;
    align-items: flex-end;
  }

  .status-pill {
    padding: 4px 8px;
  }
}
</style>
