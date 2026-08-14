import { parseReservationDateTime } from './reservationTime.js'

const ACTIVE_STATUSES = new Set(['PENDING', 'CONFIRMED'])
const RESERVABLE_MODES = new Set(['RESERVABLE', 'OPEN_USE'])

export const addCalendarDays = (dateKey, amount) => {
  const match = String(dateKey || '').match(/^(\d{4})-(\d{2})-(\d{2})$/)

  if (!match) {
    return ''
  }

  const date = new Date(Date.UTC(
    Number(match[1]),
    Number(match[2]) - 1,
    Number(match[3]) + Number(amount || 0)
  ))

  return date.toISOString().slice(0, 10)
}

export const availabilityRangeForDate = (dateKey) => ({
  from: dateKey,
  to: addCalendarDays(dateKey, 1)
})

export const isDateWithinPolicyWindow = (
  dateKey,
  todayKey,
  windowDays
) => {
  const endExclusive = addCalendarDays(todayKey, Number(windowDays || 0))
  return Boolean(endExclusive) && dateKey >= todayKey && dateKey < endExclusive
}

export const rangesOverlap = (startA, endA, startB, endB) => {
  return startA < endB && endA > startB
}

export const isActiveAvailabilityItem = (item) => {
  return item?.type === 'blocked' ||
    item?.isAvailabilityBlock === true ||
    ACTIVE_STATUSES.has(String(item?.status || '').toUpperCase())
}

export const getResourceEligibility = (resource, policy, isAdmin = false) => {
  if (!resource || resource.status !== 'available') {
    return { eligible: false, reason: 'El recurso no está activo.' }
  }

  const mode = String(resource.reservationMode || '').toUpperCase()
  if (!RESERVABLE_MODES.has(mode)) {
    if (mode === 'ADMIN_ONLY' && isAdmin) {
      return { eligible: false, reason: 'El MVP1 no permite reservas administrativas.' }
    }

    return { eligible: false, reason: 'El recurso no admite reservas particulares.' }
  }

  if (!policy?.resourceIds?.includes(Number(resource.id))) {
    return { eligible: false, reason: 'El recurso no está incluido en la política vigente.' }
  }

  return { eligible: true, reason: '' }
}

export const getEligibleResources = (resources, policy, isAdmin = false) => {
  return (resources || []).filter((resource) => {
    return getResourceEligibility(resource, policy, isAdmin).eligible
  })
}

export const hasReservationConflict = ({
  items,
  resource,
  userId,
  start,
  end
}) => {
  const candidateStart = start instanceof Date
    ? start
    : parseReservationDateTime(start)
  const candidateEnd = end instanceof Date
    ? end
    : parseReservationDateTime(end)

  if (!candidateStart || !candidateEnd || candidateEnd <= candidateStart) {
    return true
  }

  return (items || []).some((item) => {
    if (!isActiveAvailabilityItem(item)) {
      return false
    }

    const itemStart = parseReservationDateTime(item.startTime)
    if (!itemStart) {
      return false
    }

    const itemEnd = new Date(
      itemStart.getTime() + Number(item.durationMinutes || 0) * 60000
    )

    if (!rangesOverlap(candidateStart, candidateEnd, itemStart, itemEnd)) {
      return false
    }

    if (
      (item.type === 'blocked' || item.isAvailabilityBlock === true) &&
      Number(item.resourceId) === Number(resource?.id)
    ) {
      return true
    }

    if (Number(item.userId) > 0 && Number(item.userId) === Number(userId)) {
      return true
    }

    return resource?.reservationMode !== 'OPEN_USE' &&
      Number(item.resourceId) === Number(resource?.id)
  })
}
