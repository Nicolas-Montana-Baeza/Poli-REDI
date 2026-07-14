const dateTimePattern =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?/

const datePattern =
  /^(\d{4})-(\d{2})-(\d{2})/

const pad = (value) => {
  return String(value).padStart(2, '0')
}

export const parseReservationDateTime = (startTime) => {
  if (!startTime) {
    return null
  }

  const value = String(startTime)
  const match = value.match(dateTimePattern)

  if (match) {
    const [
      ,
      year,
      month,
      day,
      hour,
      minute,
      second = '0'
    ] = match

    return new Date(
      Number(year),
      Number(month) - 1,
      Number(day),
      Number(hour),
      Number(minute),
      Number(second)
    )
  }

  const parsed = new Date(value)

  if (Number.isNaN(parsed.getTime())) {
    return null
  }

  return parsed
}

export const getReservationDateKey = (startTime) => {
  if (!startTime) {
    return ''
  }

  const value = String(startTime)
  const match = value.match(datePattern)

  if (match) {
    return `${match[1]}-${match[2]}-${match[3]}`
  }

  const parsed = parseReservationDateTime(value)

  if (!parsed) {
    return ''
  }

  return [
    parsed.getFullYear(),
    pad(parsed.getMonth() + 1),
    pad(parsed.getDate())
  ].join('-')
}

export const getReservationStartMinutes = (startTime) => {
  const parsed = parseReservationDateTime(startTime)

  if (!parsed) {
    return null
  }

  return parsed.getHours() * 60 + parsed.getMinutes()
}

export const formatReservationTimeRange = (
  startTime,
  durationMinutes = 60
) => {
  const parsed = parseReservationDateTime(startTime)

  if (!parsed) {
    return ''
  }

  const endDate =
    new Date(parsed.getTime() + durationMinutes * 60000)

  const start =
    `${pad(parsed.getHours())}:${pad(parsed.getMinutes())}`

  const end =
    `${pad(endDate.getHours())}:${pad(endDate.getMinutes())}`

  return `${start} - ${end}`
}

export const formatReservationDate = (startTime) => {
  const parsed = parseReservationDateTime(startTime)

  if (!parsed) {
    return ''
  }

  return parsed.toLocaleDateString('es-CL', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
}

export const getReservationEndDate = (reservation) => {
  const start = parseReservationDateTime(reservation?.startTime)

  if (!start || !reservation) {
    return null
  }

  return new Date(
    start.getTime() +
    Number(reservation.durationMinutes || 0) * 60000
  )
}

export const getReservationTemporalState = (reservation) => {
  const start = parseReservationDateTime(reservation?.startTime)
  const end = getReservationEndDate(reservation)

  if (!start || !end) {
    return 'unknown'
  }

  const now = Date.now()

  if (end.getTime() < now) {
    return 'past'
  }

  if (start.getTime() <= now && end.getTime() >= now) {
    return 'ongoing'
  }

  return 'upcoming'
}

export const isReservationHistorical = (reservation) => {
  return (
    reservation?.status === 'CANCELLED' ||
    getReservationTemporalState(reservation) === 'past'
  )
}

export const isReservationActionable = (reservation) => {
  return !isReservationHistorical(reservation)
}

export const isReservationCancelable = (reservation) => {
  return (
    reservation?.status !== 'CANCELLED' &&
    getReservationTemporalState(reservation) !== 'past'
  )
}

export const getReservationDisplayStatus = (reservation) => {
  if (reservation?.isScheduledActivity) {
    return {
      label: 'Programación institucional',
      className: 'scheduled'
    }
  }

  const temporalState = getReservationTemporalState(reservation)

  switch (reservation?.status) {
    case 'CONFIRMED':
      if (temporalState === 'past') {
        return {
          label: 'Finalizada',
          className: 'completed'
        }
      }

      if (temporalState === 'ongoing') {
        return {
          label: 'En curso',
          className: 'ongoing'
        }
      }

      return {
        label: 'Confirmada',
        className: 'confirmed'
      }

    case 'PENDING':
      return {
        label: 'Pendiente',
        className: 'pending'
      }

    case 'CANCELLED':
      return {
        label: 'Cancelada',
        className: 'cancelled'
      }

    case 'REJECTED':
      return {
        label: 'Rechazada',
        className: 'rejected'
      }

    case 'EXPIRED':
      return {
        label: 'Expirada',
        className: 'expired'
      }

    default:
      return {
        label: reservation?.status || 'Reserva',
        className: String(reservation?.status || 'default').toLowerCase()
      }
  }
}
