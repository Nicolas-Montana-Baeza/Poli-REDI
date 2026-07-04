<script setup>
import ResourceTimeline from './ResourceTimeline.vue'
import { getReservationDateKey } from '@/utils/reservationTime'

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
    default: 8
  },

  endHour: {
    type: Number,
    default: 22
  },

  pixelsPerMinute: {
    type: Number,
    default: 1
  }
})

const emit = defineEmits([
  'slot-selected'
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

        :pixels-per-minute="pixelsPerMinute"

        @slot-selected="
          handleSlotSelected
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
  font-weight: 700;

  color: #0f172a;
}

.section-header p {
  margin-top: 4px;

  font-size: 14px;

  color: #64748b;
}

/* EMPTY */
.empty {
  background: white;

  border-radius: 22px;

  padding: 24px;

  border: 1px dashed #cbd5e1;

  color: #64748b;

  font-weight: 600;
}

/* WRAPPER */
.timelines-wrapper {
  display: flex;

  gap: 20px;

  overflow-x: auto;

  padding-bottom: 12px;

  scroll-behavior: smooth;
}

/* SCROLLBAR */
.timelines-wrapper::-webkit-scrollbar {
  height: 8px;
}

.timelines-wrapper::-webkit-scrollbar-thumb {
  background: #cbd5e1;

  border-radius: 999px;
}

/* MOBILE */
@media (max-width: 768px) {
  .timelines-wrapper {
    gap: 16px;
  }
}
</style>
