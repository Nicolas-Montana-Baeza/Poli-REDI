export const RESERVATION_OPENING_HOUR = 8
export const RESERVATION_CLOSING_HOUR = 22
export const RESERVATION_SLOT_MINUTES = 15
export const RESERVATION_ALLOWED_DURATIONS = [30, 60, 90, 120, 150, 180]

export const DEFAULT_RESERVATION_POLICY = Object.freeze({
  reservableWindowDays: 14,
  openingMinute: RESERVATION_OPENING_HOUR * 60,
  closingMinute: RESERVATION_CLOSING_HOUR * 60,
  slotIntervalMinutes: RESERVATION_SLOT_MINUTES,
  allowedDurations: RESERVATION_ALLOWED_DURATIONS,
  resourceIds: []
})

export const RESERVATION_DURATION_OPTIONS = [
  { label: '30 minutos', value: 30 },
  { label: '1 hora', value: 60 },
  { label: '1 hora 30 min', value: 90 },
  { label: '2 horas', value: 120 },
  { label: '2 horas 30 min', value: 150 },
  { label: '3 horas', value: 180 }
]

const hourPattern = /^(\d{2}):(\d{2})$/

export const formatScheduleMinute = (minuteOfDay) => {
  const hour = Math.floor(minuteOfDay / 60)
  const minute = minuteOfDay % 60
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

export const normalizeReservationPolicy = (policy) => {
  const normalized = {
    reservableWindowDays: Number(policy?.reservableWindowDays),
    openingMinute: Number(policy?.openingMinute),
    closingMinute: Number(policy?.closingMinute),
    slotIntervalMinutes: Number(policy?.slotIntervalMinutes),
    allowedDurations: (policy?.allowedDurations || []).map(Number),
    resourceIds: (policy?.resourceIds || []).map(Number)
  }

  const valid =
    Number.isInteger(normalized.reservableWindowDays) &&
    normalized.reservableWindowDays > 0 &&
    Number.isInteger(normalized.openingMinute) &&
    normalized.openingMinute >= 0 &&
    Number.isInteger(normalized.closingMinute) &&
    normalized.closingMinute <= 1440 &&
    normalized.openingMinute < normalized.closingMinute &&
    Number.isInteger(normalized.slotIntervalMinutes) &&
    normalized.slotIntervalMinutes > 0 &&
    normalized.allowedDurations.length > 0 &&
    normalized.allowedDurations.every(duration => Number.isInteger(duration) && duration > 0) &&
    normalized.resourceIds.length > 0 &&
    normalized.resourceIds.every(id => Number.isInteger(id) && id > 0)

  return valid ? normalized : null
}

export const getDurationOptions = (policy = DEFAULT_RESERVATION_POLICY) => {
  return policy.allowedDurations.map((duration) => ({
    value: duration,
    label: duration % 60 === 0
      ? `${duration / 60} ${duration === 60 ? 'hora' : 'horas'}`
      : `${duration} minutos`
  }))
}

export const snapToReservationSlot = (
  minuteOfDay,
  intervalMinutes = RESERVATION_SLOT_MINUTES
) => {
  const interval = Number(intervalMinutes) || RESERVATION_SLOT_MINUTES
  return Math.round(minuteOfDay / interval) * interval
}

export const getLatestReservationStart = (
  durationMinutes,
  policy = DEFAULT_RESERVATION_POLICY
) => {
  return formatScheduleMinute(
    Number(policy.closingMinute) - Number(durationMinutes)
  )
}

export const getReservationScheduleError = ({
  hour,
  durationMinutes,
  policy = DEFAULT_RESERVATION_POLICY
}) => {
  const duration = Number(durationMinutes)

  if (!policy?.allowedDurations?.includes(duration)) {
    return {
      field: 'durationMinutes',
      message: 'Selecciona una duración permitida por la política vigente.'
    }
  }

  const match = String(hour || '').match(hourPattern)
  if (!match) {
    return {
      field: 'hour',
      message: 'Selecciona una hora de inicio válida.'
    }
  }

  const startMinute = Number(match[1]) * 60 + Number(match[2])
  const openingMinute = Number(policy.openingMinute)
  const closingMinute = Number(policy.closingMinute)
  const intervalMinutes = Number(policy.slotIntervalMinutes)

  if ((startMinute - openingMinute) % intervalMinutes !== 0) {
    return {
      field: 'hour',
      message: `La hora de inicio debe usar intervalos de ${intervalMinutes} minutos.`
    }
  }

  if (startMinute < openingMinute) {
    return {
      field: 'hour',
      message: `La jornada de reservas comienza a las ${formatScheduleMinute(openingMinute)}.`
    }
  }

  if (startMinute >= closingMinute) {
    return {
      field: 'hour',
      message: `La hora de inicio debe ser anterior a las ${formatScheduleMinute(closingMinute)}.`
    }
  }

  if (startMinute + duration > closingMinute) {
    return {
      field: 'durationMinutes',
      message: `La reserva debe finalizar a más tardar a las ${formatScheduleMinute(closingMinute)}.`
    }
  }

  return null
}
