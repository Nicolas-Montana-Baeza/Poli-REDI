<script setup>
import { computed } from 'vue'
import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import WorkshopEnrollmentBadge from './WorkshopEnrollmentBadge.vue'
import {
  formatReservationTimeRange,
  getReservationDisplayStatus,
  getReservationStartMinutes
} from '@/utils/reservationTime'
import {
  getAvailabilityDisplayTitle,
  getAvailabilityType
} from '@/utils/availabilityType'
import {
  getWorkshopEnrollmentLabel,
  isWorkshopAvailabilityItem
} from '@/utils/workshopEnrollment'

const props = defineProps({
  reservation: {
    type: Object,
    required: true
  },

  resource: {
    type: Object,
    default: null
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

const reservationTitle = computed(() =>
  getAvailabilityDisplayTitle(props.reservation)
)

const reservationTime = computed(() => {
  return formatReservationTimeRange(
    props.reservation.startTime,
    duration.value
  )
})

const isCompact = computed(() => {
  return duration.value < 60
})

const availabilityType = computed(() =>
  getAvailabilityType(props.reservation, props.resource)
)

const reservationStatus = computed(() =>
  props.reservation.isScheduledActivity
    ? { label: 'Programada', className: 'confirmed' }
    : getReservationDisplayStatus(props.reservation)
)
const showWorkshopEnrollment = computed(() =>
  isWorkshopAvailabilityItem(props.reservation)
)
const workshopEnrollmentLabel = computed(() =>
  showWorkshopEnrollment.value
    ? getWorkshopEnrollmentLabel(props.reservation)
    : ''
)

const accessibleLabel = computed(() => {
  return [
    availabilityType.value.label,
    reservationTitle.value,
    reservationStatus.value.label,
    workshopEnrollmentLabel.value,
    reservationTime.value,
    'Abrir detalle'
  ].filter(Boolean).join('. ')
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
    :class="[
      statusClass,
      { compact: isCompact }
    ]"
    :style="blockStyle"
    :aria-label="accessibleLabel"
    @click.stop="emit('select', reservation)"
  >
    <span class="block-heading">
      <strong>
        {{ reservationTitle }}
      </strong>

      <span
        v-if="isCompact && showWorkshopEnrollment"
        class="block-indicators"
      >
        <WorkshopEnrollmentBadge
          :item="reservation"
          compact
          aria-hidden
        />

        <AvailabilityTypeChip
          :item="reservation"
          :resource="resource"
          :compact="isCompact"
          aria-hidden
        />
      </span>

      <AvailabilityTypeChip
        v-else
        :item="reservation"
        :resource="resource"
        :compact="isCompact"
        aria-hidden
      />
    </span>

    <span class="block-meta">
      <span class="reservation-time">
        {{ reservationTime }}
      </span>

      <WorkshopEnrollmentBadge
        v-if="!isCompact && showWorkshopEnrollment"
        :item="reservation"
        compact
        aria-hidden
      />
    </span>
  </button>
</template>

<style scoped>
.reservation-block {
  position: absolute;

  left: 10px;
  right: 10px;

  border-radius: var(--radius-md);

  min-height: 28px;

  box-sizing: border-box;

  padding: 7px 10px;

  display: flex;
  flex-direction: column;
  gap: 3px;

  overflow: hidden;

  z-index: 5;

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

  gap: 0;

  padding: 4px 8px;
}

.block-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
}

.block-heading strong {
  flex: 1 1 auto;
  min-width: 0;
  font-size: 13px;
  font-weight: 800;

  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.block-indicators {
  display: inline-flex;
  flex: 0 0 auto;
  min-width: 0;
  align-items: center;
  gap: 4px;
}

.block-heading :deep(.availability-type-chip) {
  flex: 0 0 auto;
  max-width: none;
}

.reservation-block.compact .block-heading strong {
  font-size: 12px;

  line-height: 1.15;
}

.reservation-time {
  min-width: 0;

  font-size: 12px;
  font-weight: 600;

  opacity: 0.9;

  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.block-meta {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.block-meta .reservation-time {
  flex: 1 1 auto;
}

.reservation-block.compact .block-meta {
  display: none;
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

.cancelled {
  background: #e2e8f0;

  color: #64748b;

  border-color: #cbd5e1;

  opacity: 0.7;
}
</style>
