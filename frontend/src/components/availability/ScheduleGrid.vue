<script setup>
import ResourceTimeline from './ResourceTimeline.vue'
import { getReservationDateKey } from '@/utils/reservationTime'
import {
  RESERVATION_CLOSING_HOUR,
  RESERVATION_OPENING_HOUR
} from '@/utils/reservationRules'

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
    default: ''
  },

  startHour: {
    type: Number,
    default: RESERVATION_OPENING_HOUR
  },

  endHour: {
    type: Number,
    default: RESERVATION_CLOSING_HOUR
  },

  pixelsPerMinute: {
    type: Number,
    default: 1
  }
})

const emit = defineEmits([
  'slot-selected',
  'reservation-selected'
])

/* DATE FILTER */
const getDateFromReservation = (reservation) => {
  return getReservationDateKey(reservation.startTime)
}

const filteredReservations = (resourceId) => {
  return props.reservations.filter((reservation) => {
    const sameResource =
      reservation.resourceId === resourceId

    const notCancelled =
      reservation.status !== 'CANCELLED'

    const sameDate =
      !props.selectedDate ||
      getDateFromReservation(reservation) === props.selectedDate

    return sameResource && notCancelled && sameDate
  })
}

/* SLOT SELECT */
const handleSlotSelected = (slot) => {
  emit('slot-selected', slot)
}

const handleReservationSelected = (reservation) => {
  emit('reservation-selected', reservation)
}
</script>

<template>
  <section class="schedule-section">

    <!-- HEADER -->
    <div class="section-header">

      <div>

        <h2>
          Disponibilidad
        </h2>

        <p>
          Selecciona cualquier punto de la línea de tiempo para crear una reserva.
        </p>

      </div>

    </div>

    <!-- EMPTY -->
    <div
      v-if="resources.length === 0"
      class="empty"
    >
      No hay recursos disponibles.
    </div>

    <!-- TIMELINES -->
    <div
      v-else
      class="timelines-wrapper"
    >

      <ResourceTimeline
        v-for="resource in resources"
        :key="resource.id"

        :resource="resource"

        :reservations="
          filteredReservations(resource.id)
        "

        :start-hour="startHour"

        :end-hour="endHour"

        :selected-date="selectedDate"

        :pixels-per-minute="pixelsPerMinute"

        @slot-selected="
          handleSlotSelected
        "

        @reservation-selected="
          handleReservationSelected
        "
      />

    </div>

  </section>
</template>

<style scoped>
.schedule-section {
  display: flex;
  flex-direction: column;

  gap: 20px;
}

/* HEADER */
.section-header h2 {
  margin: 0;

  font-size: 24px;
  font-weight: 800;

  color: var(--color-text);
}

.section-header p {
  margin-top: 4px;

  font-size: 14px;

  color: var(--color-text-muted);
}

/* EMPTY */
.empty {
  background: var(--color-surface);

  border-radius: var(--radius-lg);

  padding: var(--space-5);

  border: 1px dashed var(--color-border);

  color: #64748b;

  font-weight: 600;
}

/* WRAPPER */
.timelines-wrapper {
  display: flex;

  gap: var(--space-4);

  overflow-x: auto;

  padding: 2px 2px 12px;

  scroll-behavior: smooth;
}

/* SCROLLBAR */
.timelines-wrapper::-webkit-scrollbar {
  height: 8px;
}

.timelines-wrapper::-webkit-scrollbar-thumb {
  background: #b9c7da;

  border-radius: 999px;
}

/* MOBILE */
@media (max-width: 768px) {
  .timelines-wrapper {
    gap: 16px;
  }
}
</style>
