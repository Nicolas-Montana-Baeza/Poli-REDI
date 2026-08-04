import { parseReservationDateTime } from './reservationTime.js'

const activeStatus = item => (
  String(item?.status || '').toUpperCase() !== 'CANCELLED'
)

export const intervalsOverlap = (startA, endA, startB, endB) => (
  startA < endB && startB < endA
)

export const getEligibleResources = ({
  resources = [],
  policy = null,
  isAdmin = false
}) => {
  const allowedIds = new Set(
    (policy?.resourceIds || []).map(Number)
  )

  return resources.filter((resource) => {
    const mode = String(resource?.reservationMode || '').toUpperCase()

    if (
      resource?.isActive === false ||
      ['inactive', 'maintenance'].includes(
        String(resource?.status || '').toLowerCase()
      ) ||
      !allowedIds.has(Number(resource?.id)) ||
      !['RESERVABLE', 'OPEN_USE', 'ADMIN_ONLY'].includes(mode)
    ) {
      return false
    }

    return mode !== 'ADMIN_ONLY' || isAdmin
  })
}

export const getAvailabilityBlockingFlags = ({
  item,
  resource = null,
  currentUserId = null
}) => {
  const kind = String(item?.availabilityKind || '').toUpperCase()
  const mode = String(resource?.reservationMode || '').toUpperCase()
  const explicitBlocksResource = typeof item?.blocksResource === 'boolean'
  const explicitCurrentConflict =
    typeof item?.isCurrentUserConflict === 'boolean'
  const isWorkshop = kind === 'WORKSHOP' || item?.isWorkshop === true
  const isScheduled = kind === 'SCHEDULED_ACTIVITY' ||
    item?.isScheduledActivity === true
  const isAvailabilityBlock = kind === 'AVAILABILITY_BLOCK'
  const legacyReservation = !isWorkshop && !isScheduled && !isAvailabilityBlock

  return {
    blocksResource: activeStatus(item) && (
      explicitBlocksResource
        ? item.blocksResource
        : isAvailabilityBlock || mode !== 'OPEN_USE'
    ),
    isCurrentUserConflict: activeStatus(item) && (
      explicitCurrentConflict
        ? item.isCurrentUserConflict
        : legacyReservation &&
          Number(currentUserId) > 0 &&
          Number(item?.userId) === Number(currentUserId)
    )
  }
}

export const doesAvailabilityItemBlockInterval = ({
  item,
  candidateResourceId,
  candidateStart,
  candidateEnd,
  resources = [],
  currentUserId = null
}) => {
  if (!activeStatus(item)) {
    return false
  }

  const itemStart = parseReservationDateTime(item?.startTime)

  if (!itemStart) {
    return false
  }

  const itemEnd = new Date(
    itemStart.getTime() + Number(item?.durationMinutes || 0) * 60000
  )

  if (
    !Number.isFinite(itemEnd.getTime()) ||
    !intervalsOverlap(candidateStart, candidateEnd, itemStart, itemEnd)
  ) {
    return false
  }

  const resource = resources.find(
    candidate => Number(candidate.id) === Number(item.resourceId)
  )
  const flags = getAvailabilityBlockingFlags({
    item,
    resource,
    currentUserId
  })

  return (
    flags.isCurrentUserConflict ||
    (
      Number(item.resourceId) === Number(candidateResourceId) &&
      flags.blocksResource
    )
  )
}

export const hasAvailabilityIntervalConflict = ({
  items = [],
  resourceId,
  start,
  end,
  resources = [],
  currentUserId = null
}) => items.some(item => doesAvailabilityItemBlockInterval({
  item,
  candidateResourceId: resourceId,
  candidateStart: start,
  candidateEnd: end,
  resources,
  currentUserId
}))
