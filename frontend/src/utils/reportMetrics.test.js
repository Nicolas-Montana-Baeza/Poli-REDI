import assert from 'node:assert/strict'
import test from 'node:test'

import {
  calculateReportMetrics,
  formatReservedMinutes
} from './reportMetrics.js'

const now = Date.parse('2026-10-24T12:00:00-03:00')

test('excluye reservas finalizadas y estados terminales', () => {
  const reservations = [
    { status: 'CONFIRMED', startTime: '2026-10-24T11:30:00-03:00',
      durationMinutes: 90, resourceId: 1 },
    { status: 'PENDING', startTime: '2026-10-25T08:15:00-03:00',
      durationMinutes: 45, resourceId: '1' },
    { status: 'CONFIRMED', startTime: '2026-10-24T10:00:00-03:00',
      durationMinutes: 60, resourceId: 2 },
    { status: 'CANCELLED', startTime: '2026-10-25T09:00:00-03:00',
      durationMinutes: 60, resourceId: 2 },
    { status: 'REJECTED', startTime: '2026-10-25T09:00:00-03:00',
      durationMinutes: 60, resourceId: 2 },
    { status: 'EXPIRED', startTime: '2026-10-25T09:00:00-03:00',
      durationMinutes: 60, resourceId: 2 }
  ]

  const result = calculateReportMetrics(
    reservations,
    [{ id: 1, name: 'Cancha 1' }, { id: 2, name: 'Cancha 2' }],
    now
  )

  assert.equal(result.activeCount, 2)
  assert.equal(result.totalMinutes, 135)
  assert.equal(result.resourceCount, 1)
  assert.equal(result.resourceUsage[0].count, 2)
  assert.equal(result.resourceUsage[0].minutes, 135)
  assert.equal(result.statusSummary.reduce((sum, item) => sum + item.count, 0), 6)
})

test('conserva la precisión de las duraciones', () => {
  assert.equal(formatReservedMinutes(90), '1 h 30 min')
  assert.equal(formatReservedMinutes(135), '2 h 15 min')
  assert.equal(formatReservedMinutes(45), '0 h 45 min')
  assert.equal(formatReservedMinutes(120), '2 h')
})

test('utiliza la hora de Santiago y no la hora del navegador', () => {
  const result = calculateReportMetrics([{
    status: 'CONFIRMED',
    startTime: '2026-10-25T13:15:00Z',
    durationMinutes: 60,
    resourceId: 1
  }], [{ id: 1, name: 'Cancha' }], now)

  assert.equal(result.peakHours[0].hour, '10:00')
})

test('no limita el total de recursos a los ocho mostrados', () => {
  const resources = Array.from({ length: 9 }, (_, i) => ({
    id: i + 1,
    name: `Recurso ${i + 1}`
  }))

  const reservations = resources.map((resource) => ({
    status: 'CONFIRMED',
    startTime: '2026-10-25T10:00:00-03:00',
    durationMinutes: 60,
    resourceId: resource.id
  }))

  const result = calculateReportMetrics(reservations, resources, now)

  assert.equal(result.resourceCount, 9)
  assert.equal(result.resourceUsage.length, 8)
})

test('excluye fechas y duraciones inválidas', () => {
  const result = calculateReportMetrics([
    { status: 'CONFIRMED', startTime: 'fecha inválida',
      durationMinutes: 60, resourceId: 1 },
    { status: 'CONFIRMED', startTime: '2026-10-25T10:00:00-03:00',
      durationMinutes: 0, resourceId: 1 }
  ], [], now)

  assert.equal(result.activeCount, 0)
  assert.equal(result.totalMinutes, 0)
})
