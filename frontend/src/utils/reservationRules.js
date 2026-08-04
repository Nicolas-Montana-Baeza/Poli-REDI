export const RESERVATION_OPENING_HOUR = 8
export const RESERVATION_CLOSING_HOUR = 22
export const RESERVATION_SLOT_MINUTES = 15
export const RESERVATION_ALLOWED_DURATIONS = [30, 60, 90, 120, 150, 180]

export const RESERVATION_DURATION_OPTIONS = [
  { label: '30 minutos', value: 30 },
  { label: '1 hora', value: 60 },
  { label: '1 hora 30 min', value: 90 },
  { label: '2 horas', value: 120 },
  { label: '2 horas 30 min', value: 150 },
  { label: '3 horas', value: 180 }
]

const hourPattern = /^(\d{2}):(\d{2})$/

const positiveInteger = (value, fallback) => {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback
}

export const getReservationPolicyRules = (policy = null) => {
  const openingMinute = Number.isInteger(Number(policy?.openingMinute))
    ? Number(policy.openingMinute)
    : RESERVATION_OPENING_HOUR * 60
  const closingMinute = Number.isInteger(Number(policy?.closingMinute))
    ? Number(policy.closingMinute)
    : RESERVATION_CLOSING_HOUR * 60
  const slotIntervalMinutes = positiveInteger(
    policy?.slotIntervalMinutes,
    RESERVATION_SLOT_MINUTES
  )
  const configuredDurations = Array.isArray(policy?.allowedDurations)
    ? policy.allowedDurations
        .map(Number)
        .filter(value => Number.isInteger(value) && value > 0)
    : []

  return {
    openingMinute,
    closingMinute,
    slotIntervalMinutes,
    allowedDurations: configuredDurations.length
      ? [...new Set(configuredDurations)].sort((a, b) => a - b)
      : [...RESERVATION_ALLOWED_DURATIONS],
    reservableWindowDays: positiveInteger(policy?.reservableWindowDays, 7)
  }
}

export const isCompleteReservationPolicy = (policy) => {
  if (!policy || !Array.isArray(policy.resourceIds)) {
    return false
  }

  const rules = getReservationPolicyRules(policy)

  return (
    Array.isArray(policy.groupResourceIds) &&
    Array.isArray(policy.allowedDurations) &&
    policy.allowedDurations.length > 0 &&
    Number.isInteger(Number(policy.openingMinute)) &&
    Number.isInteger(Number(policy.closingMinute)) &&
    Number.isInteger(Number(policy.slotIntervalMinutes)) &&
    Number.isInteger(Number(policy.reservableWindowDays)) &&
    rules.openingMinute >= 0 &&
    rules.closingMinute <= 1440 &&
    rules.openingMinute < rules.closingMinute &&
    rules.slotIntervalMinutes > 0 &&
    rules.allowedDurations.every(duration => (
      duration <= rules.closingMinute - rules.openingMinute
    )) &&
    rules.reservableWindowDays > 0
  )
}

export const formatScheduleMinute = (minuteOfDay) => {
  const hour = Math.floor(minuteOfDay / 60)
  const minute = minuteOfDay % 60

  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

export const snapToReservationSlot = (
  minuteOfDay,
  slotIntervalMinutes = RESERVATION_SLOT_MINUTES
) => {
  const interval = positiveInteger(
    slotIntervalMinutes,
    RESERVATION_SLOT_MINUTES
  )

  return Math.round(minuteOfDay / interval) * interval
}

export const getLatestReservationStart = (
  durationMinutes,
  policy = null
) => {
  const duration = Number(durationMinutes)
  const { closingMinute } = getReservationPolicyRules(policy)

  return formatScheduleMinute(closingMinute - duration)
}

export const getReservationScheduleError = ({
  hour,
  durationMinutes,
  policy = null
}) => {
  const duration = Number(durationMinutes)
  const rules = getReservationPolicyRules(policy)

  if (!rules.allowedDurations.includes(duration)) {
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
  const { openingMinute, closingMinute, slotIntervalMinutes } = rules

  if ((startMinute - openingMinute) % slotIntervalMinutes !== 0) {
    return {
      field: 'hour',
      message: `La hora de inicio debe usar intervalos de ${slotIntervalMinutes} minutos.`
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

export const getReservableDateRange = (policy = null, today) => {
  const start = new Date(`${today}T00:00:00`)

  if (Number.isNaN(start.getTime())) {
    return { min: '', max: '' }
  }

  const { reservableWindowDays } = getReservationPolicyRules(policy)
  const end = new Date(start)
  end.setDate(end.getDate() + reservableWindowDays - 1)

  const format = value => [
    value.getFullYear(),
    String(value.getMonth() + 1).padStart(2, '0'),
    String(value.getDate()).padStart(2, '0')
  ].join('-')

  return { min: format(start), max: format(end) }
}

export const getReservationDateError = ({ date, policy, today }) => {
  const range = getReservableDateRange(policy, today)

  if (!date || !range.min || date < range.min || date > range.max) {
    return `La fecha debe estar entre ${range.min} y ${range.max}.`
  }

  return ''
}
