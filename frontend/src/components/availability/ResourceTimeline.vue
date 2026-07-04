<script setup>
import { computed } from 'vue'

import ReservationBlock from './ReservationBlock.vue'
import {
  getReservationStartMinutes as getReservationStartMinutesFromTime
} from '@/utils/reservationTime'

const props = defineProps({
  resource: {
    type: Object,
    required: true
  },

  reservations: {
    type: Array,
    default: () => []
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

/* HEIGHT */
const totalMinutes = computed(() => {
  return (
    props.endHour - props.startHour
  ) * 60
})

const timelineHeight = computed(() => {
  return `${totalMinutes.value * props.pixelsPerMinute}px`
})

/* HOUR LINES */
const hourLines = computed(() => {
  const lines = []

  for (
    let hour = props.startHour;
    hour <= props.endHour;
    hour++
  ) {
    lines.push({
      hour,
      label: `${String(hour).padStart(2, '0')}:00`,
      top:
        (hour - props.startHour) *
        60 *
        props.pixelsPerMinute
    })
  }

  return lines
})

/* RESERVATIONS */
const resourceReservations = computed(() => {
  return props.reservations.filter(
    reservation =>
      reservation.resourceId === props.resource.id &&
      reservation.status !== 'CANCELLED'
  )
})

/* HELPERS */
const getReservationStartMinutes = (reservation) => {
  return getReservationStartMinutesFromTime(
    reservation.startTime
  )
}

const isMinuteReserved = (minuteOfDay) => {
  return resourceReservations.value.some(
    (reservation) => {
      const start =
        getReservationStartMinutes(reservation)

      if (start === null) {
        return false
      }

      const duration =
        reservation.durationMinutes || 60

      const end =
        start + duration

      return (
        minuteOfDay >= start &&
        minuteOfDay < end
      )
    }
  )
}

const formatMinuteToHour = (minuteOfDay) => {
  const hour =
    Math.floor(minuteOfDay / 60)

  const minutes =
    minuteOfDay % 60

  return `${String(hour).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

/* CLICK TO SELECT TIME */
const handleTimelineClick = (event) => {
  const rect =
    event.currentTarget.getBoundingClientRect()

  const y =
    event.clientY - rect.top

  const minutesFromStart =
    Math.round(y / props.pixelsPerMinute)

  const minuteOfDay =
    props.startHour * 60 +
    minutesFromStart

  const isOutsideRange =
    minuteOfDay < props.startHour * 60 ||
    minuteOfDay >= props.endHour * 60

  if (isOutsideRange) {
    return
  }

  if (isMinuteReserved(minuteOfDay)) {
    return
  }

  emit('slot-selected', {
    resource: props.resource,
    hour: formatMinuteToHour(minuteOfDay)
  })
}

const statusLabel = (status) => {
  switch (status) {
    case 'available':
      return 'Disponible'

    case 'busy':
      return 'Ocupado'

    case 'maintenance':
      return 'Mantención'

    default:
      return 'Disponible'
  }
}
</script>

<template>
  <div class="resource-timeline">

    <!-- HEADER -->
    <div class="resource-header">

      <div>

        <h3>
          {{ resource.name }}
        </h3>

        <p>
          {{ resource.type }}
        </p>

      </div>

      <span
        class="resource-status"
        :class="resource.status"
      >
        {{ statusLabel(resource.status) }}
      </span>

    </div>

    <!-- MODE -->
    <div
      v-if="resource.reservationMode"
      class="mode"
    >
      Modo:
      <strong>
        {{ resource.reservationMode }}
      </strong>
    </div>

    <!-- TIMELINE -->
    <div
      class="timeline"
      :style="{ height: timelineHeight }"
      @click="handleTimelineClick"
    >

      <!-- HOUR LINES -->
      <div
        v-for="line in hourLines"
        :key="line.label"
        class="hour-line"
        :style="{ top: `${line.top}px` }"
      >
        <span>
          {{ line.label }}
        </span>
      </div>

      <!-- RESERVATIONS -->
      <ReservationBlock
        v-for="reservation in resourceReservations"
        :key="reservation.id"
        :reservation="reservation"
        :start-hour="startHour"
        :pixels-per-minute="pixelsPerMinute"
      />

    </div>

  </div>
</template>

<style scoped>
.resource-timeline {
  min-width: 300px;

  background: white;

  border-radius: 24px;

  border: 1px solid #e2e8f0;

  overflow: hidden;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);

  display: flex;
  flex-direction: column;
}

/* HEADER */
.resource-header {
  padding: 18px;

  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 12px;

  border-bottom: 1px solid #e2e8f0;
}

.resource-header h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 800;

  color: #0f172a;
}

.resource-header p {
  margin-top: 4px;

  font-size: 13px;

  color: #64748b;
}

/* STATUS */
.resource-status {
  padding: 6px 12px;

  border-radius: 999px;

  font-size: 12px;
  font-weight: 700;

  white-space: nowrap;
}

.available {
  background: #dcfce7;
  color: #15803d;
}

.busy {
  background: #fee2e2;
  color: #dc2626;
}

.maintenance {
  background: #fef3c7;
  color: #b45309;
}

/* MODE */
.mode {
  margin: 14px 18px 0;

  padding: 10px 12px;

  border-radius: 14px;

  background: #f8fafc;

  color: #64748b;

  font-size: 13px;
}

.mode strong {
  color: #0f172a;

  text-transform: capitalize;
}

/* TIMELINE */
.timeline {
  position: relative;

  margin: 18px;

  background:
    linear-gradient(
      to bottom,
      #f8fafc 0,
      #f8fafc 1px,
      transparent 1px,
      transparent 60px
    );

  border-radius: 18px;

  border: 1px solid #e2e8f0;

  cursor: crosshair;

  overflow: hidden;
}

/* HOUR LINE */
.hour-line {
  position: absolute;

  left: 0;
  right: 0;

  height: 1px;

  background: #e2e8f0;

  z-index: 1;
}

.hour-line span {
  position: absolute;

  top: -9px;
  left: 10px;

  background: white;

  padding: 0 6px;

  font-size: 11px;
  font-weight: 700;

  color: #94a3b8;
}

/* MOBILE */
@media (max-width: 768px) {
  .resource-timeline {
    min-width: 85vw;
  }
}
</style>
