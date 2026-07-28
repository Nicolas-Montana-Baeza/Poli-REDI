import {
  formatBusinessDate,
  formatBusinessTime,
  getBusinessDateKey,
  getBusinessMinutes,
  parseBusinessDateTime
} from './businessTime.js'

export { getBusinessDateKey }

export const parseReservationDateTime = (startTime) =>
  parseBusinessDateTime(startTime)

export const getReservationDateKey = (startTime) =>
  getBusinessDateKey(startTime)

export const getReservationStartMinutes = (startTime) =>
  getBusinessMinutes(startTime)

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

  const start = formatBusinessTime(parsed)
  const end = formatBusinessTime(endDate)

  return `${start} - ${end}`
}

export const formatReservationDate = (startTime) => {
  const parsed = parseReservationDateTime(startTime)

  if (!parsed) {
    return ''
  }

  return formatBusinessDate(parsed, {
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

export const canUserCancelReservation = (reservation, user) => {
  if (!reservation || !user) {
    return false
  }

  if (user.isAdmin === true) {
    return isReservationCancelable(reservation)
  }

  return (
    isReservationCancelable(reservation) &&
    reservation.userId === user.id
  )
}

export const canUserEditReservationTarget = (reservation, user) => {
  if (!reservation || !user || reservation.canEditTarget !== true) {
    return false
  }

  const isOwner = reservation.isOwner === true ||
    Number(reservation.userId) === Number(user.id)
  const status = String(reservation.status || '').toUpperCase()
  const deadline = new Date(reservation.confirmationDeadline).getTime()

  return (
    isOwner &&
    ['PENDING', 'CONFIRMED'].includes(status) &&
    Number.isFinite(deadline) &&
    Date.now() <= deadline
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
