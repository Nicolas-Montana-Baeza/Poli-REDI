<script setup>
import { computed, ref } from 'vue'
import {
  formatReservationTimeRange,
  getReservationStartMinutes
} from '@/utils/reservationTime'
import { focusReservationBlock } from '@/utils/reservationFocus'

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

const blockElement = ref(null)

const focusAndScroll = () => {
  return focusReservationBlock(blockElement.value)
}

defineExpose({ focusAndScroll })

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

const isCompact = computed(() => {
  return duration.value < 60
})

const statusClass = computed(() => {
  if (
    props.reservation.type === 'blocked' ||
    props.reservation.isAvailabilityBlock
  ) {
    return 'blocked'
  }

  if (props.reservation.isWorkshop) {
    return 'workshop'
  }

  if (props.reservation.isScheduledActivity) {
    return 'scheduled'
  }

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
    ref="blockElement"
    class="reservation-block"
    type="button"
    :class="[
      statusClass,
      { compact: isCompact }
    ]"
    :style="blockStyle"
    :aria-label="`${statusClass === 'blocked' ? 'Ver bloqueo' : 'Ver reserva'} ${reservationTitle}`"
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

  left: 10px;
  right: 10px;

  border-radius: var(--radius-md);

  padding: 9px 11px;

  display: flex;
  flex-direction: column;
  gap: 4px;

  overflow: hidden;

  z-index: 3;

  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.14);

  border: 1px solid transparent;

  text-align: left;
  cursor: pointer;
}

.reservation-block:focus-visible {
  outline: 3px solid rgba(37,99,235,0.35);
  outline-offset: 2px;
}

.reservation-block.compact {
  justify-content: center;

  gap: 2px;

  padding: 5px 9px;
}

.reservation-block strong {
  font-size: 13px;
  font-weight: 800;

  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reservation-block.compact strong {
  font-size: 12px;

  line-height: 1.15;
}

.reservation-block span {
  font-size: 12px;
  font-weight: 600;

  opacity: 0.9;
}

.reservation-block.compact span {
  font-size: 11px;

  line-height: 1.15;
}

/* STATUS */
.confirmed {
  background: #d8e8ff;

  color: #1e4fb8;

  border-color: #a9c9ff;
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

.workshop {
  background: #fff3e6;

  color: #c2410c;

  border-color: #fdba74;
}

.scheduled {
  background: #ffedd5;

  color: #c2410c;

  border-color: #fb923c;
}

.blocked {
  background: #ffedd5;
  color: #9a3412;
  border-color: #fdba74;
}

.cancelled {
  background: #e2e8f0;

  color: #64748b;

  border-color: #cbd5e1;

  opacity: 0.7;
}
</style>
