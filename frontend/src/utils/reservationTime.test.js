import assert from 'node:assert/strict'
import test from 'node:test'

import { canUserCancelReservation } from './reservationTime.js'

test('solo permite cancelar al propietario o al administrador', () => {
  const reservation = {
    userId: 7,
    status: 'CONFIRMED',
    startTime: '2099-01-01T10:00:00',
    durationMinutes: 60
  }

  assert.equal(
    canUserCancelReservation(reservation, { id: 7, isAdmin: false }),
    true
  )

  assert.equal(
    canUserCancelReservation(reservation, { id: 99, isAdmin: false }),
    false
  )

  assert.equal(
    canUserCancelReservation(reservation, { id: 99, isAdmin: true }),
    true
  )
})

test('rechaza cancelaciones para reservas ya vencidas o canceladas', () => {
  const pastReservation = {
    userId: 7,
    status: 'CONFIRMED',
    startTime: '2000-01-01T10:00:00',
    durationMinutes: 30
  }

  const cancelledReservation = {
    userId: 7,
    status: 'CANCELLED',
    startTime: '2099-01-01T10:00:00',
    durationMinutes: 60
  }

  assert.equal(
    canUserCancelReservation(pastReservation, { id: 7, isAdmin: false }),
    false
  )

  assert.equal(
    canUserCancelReservation(cancelledReservation, { id: 7, isAdmin: false }),
    false
  )
})
