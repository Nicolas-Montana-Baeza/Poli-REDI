<script setup>
import { computed, ref } from 'vue'

import ReservationBlock from './ReservationBlock.vue'
import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import {
  getBusinessDateKey,
  getReservationStartMinutes as getReservationStartMinutesFromTime
} from '@/utils/reservationTime'
import {
  RESERVATION_CLOSING_HOUR,
  RESERVATION_OPENING_HOUR,
  RESERVATION_SLOT_MINUTES,
  snapToReservationSlot
} from '@/utils/reservationRules'
import { doesAvailabilityItemBlockInterval } from '@/utils/availabilityRules'

const props = defineProps({
  resource: {
    type: Object,
    required: true
  },

  reservations: {
    type: Array,
    default: () => []
  },

  allReservations: {
    type: Array,
    default: () => []
  },

  startHour: {
    type: Number,
    default: RESERVATION_OPENING_HOUR
  },

  endHour: {
    type: Number,
    default: RESERVATION_CLOSING_HOUR
  },

  openingMinute: { type: Number, default: null },
  closingMinute: { type: Number, default: null },
  slotIntervalMinutes: {
    type: Number,
    default: RESERVATION_SLOT_MINUTES
  },
  currentUserId: { type: [Number, String], default: null },

  selectedDate: {
    type: String,
    default: ''
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

const timelineTopPadding = 18
const timelineBottomPadding = 24
const timelineStartMinute = computed(() => (
  Number.isFinite(props.openingMinute)
    ? Number(props.openingMinute)
    : props.startHour * 60
))
const timelineEndMinute = computed(() => (
  Number.isFinite(props.closingMinute)
    ? Number(props.closingMinute)
    : props.endHour * 60
))

const now = computed(() => new Date())

const isToday = computed(() => {
  return props.selectedDate === getBusinessDateKey(now.value)
})

const currentMinuteOfDay = computed(() => {
  return getReservationStartMinutesFromTime(now.value)
})

const pastOverlayHeight = computed(() => {
  if (!isToday.value) {
    return 0
  }

  const startMinute = timelineStartMinute.value
  const endMinute = timelineEndMinute.value

  const clampedMinute =
    Math.min(
      Math.max(currentMinuteOfDay.value, startMinute),
      endMinute
    )

  return (
    timelineTopPadding +
    (clampedMinute - startMinute) *
    props.pixelsPerMinute
  )
})

const nowLineTop = computed(() => {
  if (
    !isToday.value ||
    currentMinuteOfDay.value < timelineStartMinute.value ||
    currentMinuteOfDay.value > timelineEndMinute.value
  ) {
    return null
  }

  return (
    timelineTopPadding +
    (currentMinuteOfDay.value - timelineStartMinute.value) *
    props.pixelsPerMinute
  )
})

const isPastMinute = (minuteOfDay) => {
  return (
    isToday.value &&
    minuteOfDay <= currentMinuteOfDay.value
  )
}

const isOpenUse = computed(() => {
  return props.resource.reservationMode === 'OPEN_USE'
})

/* HEIGHT */
const totalMinutes = computed(() => {
  return timelineEndMinute.value - timelineStartMinute.value
})

const timelineHeight = computed(() => {
  return `${
    totalMinutes.value * props.pixelsPerMinute +
    timelineTopPadding +
    timelineBottomPadding
  }px`
})

/* HOUR LINES */
const hourLines = computed(() => {
  const lines = []
  const firstFullHour = Math.ceil(timelineStartMinute.value / 60) * 60
  const minutes = [timelineStartMinute.value]

  for (
    let minute = firstFullHour;
    minute <= timelineEndMinute.value;
    minute += 60
  ) {
    if (!minutes.includes(minute)) minutes.push(minute)
  }

  if (!minutes.includes(timelineEndMinute.value)) {
    minutes.push(timelineEndMinute.value)
  }

  for (const minute of minutes) {
    lines.push({
      minute,
      label: formatMinuteToHour(minute),
      top:
        timelineTopPadding +
        (minute - timelineStartMinute.value) *
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
  const candidateStart = new Date(
    `${props.selectedDate}T${formatMinuteToHour(minuteOfDay)}:00`
  )
  const candidateEnd = new Date(
    candidateStart.getTime() + props.slotIntervalMinutes * 60000
  )
  const source = props.allReservations.length
    ? props.allReservations
    : props.reservations

  return source.some(item => doesAvailabilityItemBlockInterval({
    item,
    candidateResourceId: props.resource.id,
    candidateStart,
    candidateEnd,
    resources: [props.resource],
    currentUserId: props.currentUserId
  }))
}

const formatMinuteToHour = (minuteOfDay) => {
  const hour =
    Math.floor(minuteOfDay / 60)

  const minutes =
    minuteOfDay % 60

  return `${String(hour).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

const heatmapSegments = computed(() => {
  if (!isOpenUse.value) {
    return []
  }

  const segmentMinutes = props.slotIntervalMinutes
  const segments = []
  const startMinute = timelineStartMinute.value
  const endMinute = timelineEndMinute.value

  for (
    let minute = startMinute;
    minute < endMinute;
    minute += segmentMinutes
  ) {
    const segmentEnd = Math.min(
      minute + segmentMinutes,
      endMinute
    )

    const count = resourceReservations.value.filter((reservation) => {
      const reservationStart =
        getReservationStartMinutes(reservation)

      if (reservationStart === null) {
        return false
      }

      const reservationEnd =
        reservationStart + Number(reservation.durationMinutes || 60)

      return reservationStart < segmentEnd && reservationEnd > minute
    }).length

    segments.push({
      minute,
      count,
      top:
        timelineTopPadding +
        (minute - startMinute) *
        props.pixelsPerMinute,
      height:
        (segmentEnd - minute) *
        props.pixelsPerMinute
    })
  }

  return segments
})

const blockingSegments = computed(() => {
  const segments = []
  const interval = props.slotIntervalMinutes
  let active = null

  for (
    let minute = timelineStartMinute.value;
    minute < timelineEndMinute.value;
    minute += interval
  ) {
    const blocked = isMinuteReserved(minute)

    if (blocked && !active) {
      active = { start: minute, end: Math.min(minute + interval, timelineEndMinute.value) }
    } else if (blocked && active) {
      active.end = Math.min(minute + interval, timelineEndMinute.value)
    } else if (!blocked && active) {
      segments.push(active)
      active = null
    }
  }

  if (active) segments.push(active)

  return segments.map(segment => ({
    ...segment,
    top: timelineTopPadding +
      (segment.start - timelineStartMinute.value) * props.pixelsPerMinute,
    height: (segment.end - segment.start) * props.pixelsPerMinute
  }))
})

const heatmapClass = (count) => {
  if (count >= 4) {
    return 'high'
  }

  if (count >= 2) {
    return 'medium'
  }

  if (count === 1) {
    return 'low'
  }

  return 'empty'
}

const handleReservationSelected = (reservation) => {
  emit('reservation-selected', reservation)
}

/* CLICK TO SELECT TIME */
const handleTimelineClick = (event) => {
  const rect =
    event.currentTarget.getBoundingClientRect()

  const y =
    event.clientY - rect.top - timelineTopPadding

  const minutesFromStart = snapToReservationSlot(
    y / props.pixelsPerMinute,
    props.slotIntervalMinutes
  )

  const minuteOfDay =
    timelineStartMinute.value +
    minutesFromStart

  const isOutsideRange =
    minuteOfDay < timelineStartMinute.value ||
    minuteOfDay >= timelineEndMinute.value

  if (isOutsideRange) {
    return
  }

  if (isPastMinute(minuteOfDay)) {
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

const keyboardMinute = ref(timelineStartMinute.value)
const isSelectableMinute = minute => (
  minute >= timelineStartMinute.value &&
  minute < timelineEndMinute.value &&
  !isPastMinute(minute) &&
  !isMinuteReserved(minute)
)
const moveKeyboardCursor = direction => {
  const lower = timelineStartMinute.value
  const upper = timelineEndMinute.value - props.slotIntervalMinutes
  let candidate = Math.min(upper, Math.max(lower, keyboardMinute.value + direction * props.slotIntervalMinutes))
  while (candidate >= lower && candidate <= upper && !isSelectableMinute(candidate)) {
    candidate += direction * props.slotIntervalMinutes
  }
  if (candidate >= lower && candidate <= upper) keyboardMinute.value = candidate
}
const handleTimelineKeydown = event => {
  if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
    event.preventDefault()
    moveKeyboardCursor(1)
  } else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
    event.preventDefault()
    moveKeyboardCursor(-1)
  } else if (event.key === 'Home') {
    event.preventDefault()
    keyboardMinute.value = timelineStartMinute.value
    if (!isSelectableMinute(keyboardMinute.value)) moveKeyboardCursor(1)
  } else if (event.key === 'End') {
    event.preventDefault()
    keyboardMinute.value = timelineEndMinute.value - props.slotIntervalMinutes
    if (!isSelectableMinute(keyboardMinute.value)) moveKeyboardCursor(-1)
  } else if ((event.key === 'Enter' || event.key === ' ') && isSelectableMinute(keyboardMinute.value)) {
    event.preventDefault()
    emit('slot-selected', { resource: props.resource, hour: formatMinuteToHour(keyboardMinute.value) })
  }
}

const modeLabel = (mode) => {
  switch (mode) {
    case 'ADMIN_ONLY':
      return 'Solo administrador'

    case 'INFORMATIVE':
      return 'Informativo'

    case 'OPEN_USE':
      return 'Uso libre'

    default:
      return 'Reservable'
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
      :class="{ 'open-use-mode': isOpenUse }"
    >
      <template v-if="isOpenUse">
        <span class="mode-caption">
          La intensidad indica la cantidad de reservas simultáneas.
        </span>

        <AvailabilityTypeChip :resource="resource" />
      </template>

      <template v-else>
        Modo:
        <strong>
          {{ modeLabel(resource.reservationMode) }}
        </strong>
      </template>
    </div>

    <!-- TIMELINE -->
    <div
      class="timeline"
      :style="{ height: timelineHeight }"
      role="button"
      tabindex="0"
      :aria-label="`Disponibilidad de ${resource.name}. Horario seleccionado ${formatMinuteToHour(keyboardMinute)}. Usa las flechas para cambiar y Enter para reservar.`"
      @click="handleTimelineClick"
      @keydown="handleTimelineKeydown"
    >

      <div
        v-if="pastOverlayHeight > timelineTopPadding"
        class="past-overlay"
        :style="{ height: `${pastOverlayHeight}px` }"
      />

      <div
        v-if="nowLineTop !== null"
        class="now-line"
        :style="{ top: `${nowLineTop}px` }"
      >
        <span>
          Ahora
        </span>
      </div>

      <div class="blocking-layer" aria-hidden="true">
        <div
          v-for="segment in blockingSegments"
          :key="`${segment.start}-${segment.end}`"
          class="blocking-segment"
          :style="{
            top: `${segment.top}px`,
            height: `${segment.height}px`
          }"
        />
      </div>

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

      <!-- OPEN USE HEATMAP -->
      <div
        v-if="isOpenUse"
        class="heatmap-layer"
      >
        <div
          v-for="segment in heatmapSegments"
          :key="segment.minute"
          class="heatmap-segment"
          :class="heatmapClass(segment.count)"
          :style="{
            top: `${segment.top}px`,
            height: `${segment.height}px`
          }"
        >
          <span v-if="segment.count > 0">
            {{ segment.count }}
          </span>
        </div>
      </div>

      <!-- RESERVATIONS -->
      <template v-if="!isOpenUse">
        <ReservationBlock
          v-for="reservation in resourceReservations"
          :key="reservation.availabilityKey || reservation.id"
          :reservation="reservation"
          :resource="resource"
          :start-hour="timelineStartMinute / 60"
          :pixels-per-minute="pixelsPerMinute"
          :top-offset="timelineTopPadding"
          @select="handleReservationSelected"
        />
      </template>

    </div>

  </div>
</template>

<style scoped>
.resource-timeline {
  min-width: 300px;

  background: var(--color-surface);

  border-radius: var(--radius-xl);

  border: 1px solid var(--color-border);

  overflow: hidden;

  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);

  display: flex;
  flex-direction: column;
}

/* HEADER */
.resource-header {
  padding: var(--space-4);

  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 12px;

  border-bottom: 1px solid var(--color-border-soft);
}

.resource-header h3 {
  margin: 0;

  font-size: 17px;
  font-weight: 800;

  color: var(--color-text);
}

.resource-header p {
  margin-top: 4px;

  font-size: 13px;

  color: var(--color-text-muted);
}

/* STATUS */
.resource-status {
  padding: 6px 10px;

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
  margin: 12px 16px 0;

  padding: 8px 10px;

  border-radius: var(--radius-md);

  background: var(--color-surface-muted);

  color: var(--color-text-muted);

  font-size: 13px;
}

.mode strong {
  color: var(--color-text);
}

.mode-caption {
  min-width: 0;
  font-size: 12px;
}

.open-use-mode {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.open-use-mode :deep(.availability-type-chip) {
  flex: 0 0 auto;
  max-width: none;
}

/* TIMELINE */
.timeline {
  position: relative;

  margin: var(--space-4);

  background: #ffffff;

  border-radius: var(--radius-lg);

  border: 1px solid var(--color-border);

  cursor: crosshair;

  overflow: hidden;
}

.past-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 2;
  border-bottom: 1px solid rgba(100, 116, 139, 0.28);
  background:
    repeating-linear-gradient(
      -45deg,
      rgba(148, 163, 184, 0.16),
      rgba(148, 163, 184, 0.16) 7px,
      rgba(226, 232, 240, 0.28) 7px,
      rgba(226, 232, 240, 0.28) 14px
  );
  pointer-events: none;
}

.now-line {
  position: absolute;
  left: 0;
  right: 0;
  z-index: 4;
  height: 2px;
  background: #f97316;
  pointer-events: none;
}

.now-line::before {
  content: "";
  position: absolute;
  top: -4px;
  left: 10px;
  width: 10px;
  height: 10px;
  border-radius: var(--radius-pill);
  background: #f97316;
}

.now-line span {
  position: absolute;
  top: -13px;
  right: 10px;
  padding: 3px 7px;
  border-radius: var(--radius-pill);
  background: #f97316;
  color: #ffffff;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0;
}

/* HOUR LINE */
.hour-line {
  position: absolute;

  left: 0;
  right: 0;

  height: 1px;

  background: var(--color-border-soft);

  z-index: 3;
}

.hour-line span {
  position: absolute;

  top: -9px;
  left: 10px;

  background: #ffffff;

  padding: 0 6px;

  font-size: 11px;
  font-weight: 700;

  color: var(--color-text-soft);
}

.heatmap-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
}

.blocking-layer {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
}

.blocking-segment {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1px solid rgba(100, 116, 139, 0.18);
  border-bottom: 1px solid rgba(100, 116, 139, 0.18);
  background:
    repeating-linear-gradient(
      -45deg,
      rgba(148, 163, 184, 0.12),
      rgba(148, 163, 184, 0.12) 6px,
      rgba(226, 232, 240, 0.24) 6px,
      rgba(226, 232, 240, 0.24) 12px
    );
}

.heatmap-segment {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 10px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
}

.heatmap-segment.empty {
  background: rgba(22, 163, 74, 0.04);
}

.heatmap-segment.low {
  background: rgba(34, 197, 94, 0.16);
}

.heatmap-segment.medium {
  background: rgba(249, 115, 22, 0.18);
}

.heatmap-segment.high {
  background: rgba(239, 68, 68, 0.2);
}

.heatmap-segment span {
  min-width: 20px;
  height: 20px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.84);
  color: var(--color-text);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 900;
}

/* MOBILE */
@media (max-width: 768px) {
  .resource-timeline {
    min-width: 85vw;
  }
}
</style>
