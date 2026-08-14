import assert from 'node:assert/strict'
import test from 'node:test'

import {
  addCalendarDays,
  availabilityRangeForDate,
  getResourceEligibility,
  hasReservationConflict,
  isDateWithinPolicyWindow
} from './availabilityRules.js'

const policy = { resourceIds: [1, 2] }
const openUse = { id: 1, status: 'available', reservationMode: 'OPEN_USE' }
const reservable = { id: 2, status: 'available', reservationMode: 'RESERVABLE' }
const interval = {
  start: new Date('2026-08-20T12:00:00'),
  end: new Date('2026-08-20T13:00:00')
}

test('calcula un rango diario semiabierto sin depender de DST', () => {
  assert.equal(addCalendarDays('2026-08-31', 1), '2026-09-01')
  assert.deepEqual(availabilityRangeForDate('2026-09-05'), {
    from: '2026-09-05',
    to: '2026-09-06'
  })
})

test('aplica la ventana reservable como limite exclusivo', () => {
  assert.equal(isDateWithinPolicyWindow('2026-08-14', '2026-08-14', 14), true)
  assert.equal(isDateWithinPolicyWindow('2026-08-27', '2026-08-14', 14), true)
  assert.equal(isDateWithinPolicyWindow('2026-08-28', '2026-08-14', 14), false)
})

test('filtra recursos inactivos, informativos o fuera de politica', () => {
  assert.equal(getResourceEligibility(openUse, policy).eligible, true)
  assert.equal(getResourceEligibility({ ...openUse, status: 'busy' }, policy).eligible, false)
  assert.equal(getResourceEligibility({ ...openUse, reservationMode: 'INFORMATIVE' }, policy).eligible, false)
  assert.equal(getResourceEligibility({ ...openUse, id: 99 }, policy).eligible, false)
})

test('OPEN_USE permite uso ajeno pero no bloqueos ni solape propio', () => {
  const otherUse = [{
    resourceId: 1,
    userId: 9,
    startTime: '2026-08-20T12:15:00',
    durationMinutes: 30,
    status: 'CONFIRMED'
  }]
  assert.equal(hasReservationConflict({ items: otherUse, resource: openUse, userId: 7, ...interval }), false)

  const ownOtherResource = [{ ...otherUse[0], resourceId: 2, userId: 7 }]
  assert.equal(hasReservationConflict({ items: ownOtherResource, resource: openUse, userId: 7, ...interval }), true)

  const blocked = [{ ...otherUse[0], type: 'blocked', userId: 0 }]
  assert.equal(hasReservationConflict({ items: blocked, resource: openUse, userId: 7, ...interval }), true)
})

test('RESERVABLE bloquea una reserva ajena del mismo recurso', () => {
  const items = [{
    resourceId: 2,
    userId: 9,
    startTime: '2026-08-20T12:15:00',
    durationMinutes: 30,
    status: 'PENDING'
  }]
  assert.equal(hasReservationConflict({ items, resource: reservable, userId: 7, ...interval }), true)
  assert.equal(hasReservationConflict({ items: [{ ...items[0], status: 'CANCELLED' }], resource: reservable, userId: 7, ...interval }), false)
})
