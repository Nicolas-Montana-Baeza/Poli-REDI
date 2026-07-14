export const RESERVATION_OPENING_HOUR = 8
export const RESERVATION_CLOSING_HOUR = 22
export const RESERVATION_SLOT_MINUTES = 30
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

export const formatScheduleMinute = (minuteOfDay) => {
  const hour = Math.floor(minuteOfDay / 60)
  const minute = minuteOfDay % 60

  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

export const getLatestReservationStart = (durationMinutes) => {
  const duration = Number(durationMinutes)
  const closingMinute = RESERVATION_CLOSING_HOUR * 60

  return formatScheduleMinute(closingMinute - duration)
}

export const getReservationScheduleError = ({ hour, durationMinutes }) => {
  const duration = Number(durationMinutes)

  if (!RESERVATION_ALLOWED_DURATIONS.includes(duration)) {
    return {
      field: 'durationMinutes',
      message: 'Selecciona una duración de 30 a 180 minutos en intervalos de 30.'
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
  const openingMinute = RESERVATION_OPENING_HOUR * 60
  const closingMinute = RESERVATION_CLOSING_HOUR * 60

  if (startMinute % RESERVATION_SLOT_MINUTES !== 0) {
    return {
      field: 'hour',
      message: 'La hora de inicio debe usar intervalos de 30 minutos.'
    }
  }

  if (startMinute < openingMinute) {
    return {
      field: 'hour',
      message: 'La jornada de reservas comienza a las 08:00.'
    }
  }

  if (startMinute >= closingMinute) {
    return {
      field: 'hour',
      message: 'La hora de inicio debe ser anterior a las 22:00.'
    }
  }

  if (startMinute + duration > closingMinute) {
    return {
      field: 'durationMinutes',
      message: 'La reserva debe finalizar a más tardar a las 22:00.'
    }
  }

  return null
}
