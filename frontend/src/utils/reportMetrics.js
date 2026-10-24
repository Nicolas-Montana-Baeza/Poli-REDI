import { getBusinessMinutes } from './businessTime.js'
import { parseReservationDateTime } from './reservationTime.js'

const ACTIVE_STATUSES = new Set(['PENDING', 'CONFIRMED'])

const statusLabels = {
  CONFIRMED: 'Confirmadas',
  PENDING: 'Pendientes',
  CANCELLED: 'Canceladas',
  REJECTED: 'Rechazadas',
  EXPIRED: 'Expiradas',
  OTHER: 'Otros estados'
}

export const formatReservedMinutes = (value) => {
  const minutes = Math.max(0, Math.trunc(Number(value) || 0))
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60

  return rest ? `${hours} h ${rest} min` : `${hours} h`
}

export const calculateReportMetrics = (
  reservations = [],
  resources = [],
  now = Date.now()
) => {
  const all = Array.isArray(reservations) ? reservations : []
  const catalog = Array.isArray(resources) ? resources : []

  const active = all.filter((reservation) => {
    if (!ACTIVE_STATUSES.has(reservation?.status)) return false

    const start = parseReservationDateTime(reservation.startTime)
    const duration = Number(reservation.durationMinutes)

    if (!start || !Number.isFinite(duration) || duration <= 0) {
      return false
    }

    return start.getTime() + duration * 60000 > now
  })

  const usage = new Map()

  for (const resource of catalog) {
    if (resource?.id == null) continue

    usage.set(String(resource.id), {
      id: resource.id,
      name: resource.name || 'Recurso',
      count: 0,
      minutes: 0
    })
  }

  const starts = new Map()
  let totalMinutes = 0

  for (const reservation of active) {
    const minutes = Number(reservation.durationMinutes)
    totalMinutes += minutes

    if (reservation.resourceId != null) {
      const key = String(reservation.resourceId)
      const current = usage.get(key) || {
        id: reservation.resourceId,
        name: reservation.resourceName || 'Recurso',
        count: 0,
        minutes: 0
      }

      current.count += 1
      current.minutes += minutes
      usage.set(key, current)
    }

    const startMinutes = getBusinessMinutes(reservation.startTime)

    if (startMinutes != null) {
      const hour = `${String(Math.floor(startMinutes / 60)).padStart(2, '0')}:00`
      starts.set(hour, (starts.get(hour) || 0) + 1)
    }
  }

  const usedResources = [...usage.values()]
    .filter((resource) => resource.count > 0)
    .sort((a, b) => b.count - a.count || String(a.name).localeCompare(String(b.name)))

  const statuses = new Map()

  for (const reservation of all) {
    const status = reservation?.status || 'OTHER'
    statuses.set(status, (statuses.get(status) || 0) + 1)
  }

  return {
    activeCount: active.length,
    totalMinutes,
    resourceCount: usedResources.length,
    resourceUsage: usedResources.slice(0, 8),
    statusSummary: [...statuses.entries()].map(([status, count]) => ({
      label: statusLabels[status] || status,
      count
    })),
    peakHours: [...starts.entries()]
      .map(([hour, count]) => ({ hour, count }))
      .sort((a, b) => b.count - a.count || a.hour.localeCompare(b.hour))
      .slice(0, 6)
  }
}
