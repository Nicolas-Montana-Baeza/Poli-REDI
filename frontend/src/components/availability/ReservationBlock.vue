<script setup>
import { computed } from 'vue'
import {
  formatReservationTimeRange,
  getReservationStartMinutes
} from '@/utils/reservationTime'

const props = defineProps({
  reservation: {
    type: Object,
    required: true
  },

  startHour: {
    type: Number,
    default: 8
  },

  pixelsPerMinute: {
    type: Number,
    default: 1
  },

  topOffset: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits([
  'select'
])

const startMinutes = computed(() => {
  const minutesFromMidnight =
    getReservationStartMinutes(
      props.reservation.startTime
    )

  if (minutesFromMidnight === null) {
    return 0
  }

  return minutesFromMidnight -
    props.startHour * 60
})

const duration = computed(() => {
  return props.reservation.durationMinutes || 60
})

const blockStyle = computed(() => {
  return {
    top: `${props.topOffset + startMinutes.value * props.pixelsPerMinute}px`,
    height: `${duration.value * props.pixelsPerMinute}px`
  }
})

const reservationTitle = computed(() => {
  return props.reservation.title || 'Reserva'
})

const reservationTime = computed(() => {
  return formatReservationTimeRange(
    props.reservation.startTime,
    duration.value
  )
})

const statusClass = computed(() => {
  if (props.reservation.status === 'PENDING') {
    return 'pending'
  }

  if (props.reservation.status === 'CANCELLED') {
    return 'cancelled'
  }

  if (props.reservation.type === 'priority') {
    return 'priority'
  }

  return 'confirmed'
})
</script>

<template>
  <button
    class="reservation-block"
    type="button"
    :class="statusClass"
    :style="blockStyle"
    :aria-label="`Ver reserva ${reservationTitle}`"
    @click.stop="emit('select', reservation)"
  >
    <strong>
      {{ reservationTitle }}
    </strong>

    <span>
      {{ reservationTime }}
    </span>
  </button>
</template>

<style scoped>
.reservation-block {
  position: absolute;

  left: 8px;
  right: 8px;

  border-radius: 14px;

  padding: 10px 12px;

  display: flex;
  flex-direction: column;
  gap: 4px;

  overflow: hidden;

  z-index: 3;

  box-shadow:
    0 8px 18px rgba(0,0,0,0.12);

  border: 1px solid transparent;

  text-align: left;
  cursor: pointer;
}

.reservation-block:focus-visible {
  outline: 3px solid rgba(37,99,235,0.35);
  outline-offset: 2px;
}

.reservation-block strong {
  font-size: 13px;
  font-weight: 800;

  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reservation-block span {
  font-size: 12px;
  font-weight: 600;

  opacity: 0.9;
}

/* STATUS */
.confirmed {
  background: #dbeafe;

  color: #1d4ed8;

  border-color: #bfdbfe;
}

.pending {
  background: #fef3c7;

  color: #b45309;

  border-color: #fde68a;
}

.priority {
  background: #ffedd5;

  color: #c2410c;

  border-color: #fed7aa;
}

.cancelled {
  background: #e2e8f0;

  color: #64748b;

  border-color: #cbd5e1;

  opacity: 0.7;
}
</style>
