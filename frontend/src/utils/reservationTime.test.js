import assert from 'node:assert/strict'
import test from 'node:test'

import {
  canUserCancelReservation,
  canUserEditReservationTarget
} from './reservationTime.js'

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

test('solo permite editar el objetivo al propietario autorizado dentro del plazo', () => {
  const reservation = {
    userId: 7,
    status: 'PENDING',
    canEditTarget: true,
    confirmationDeadline: '2099-01-01T10:00:00'
  }

  assert.equal(
    canUserEditReservationTarget(reservation, { id: 7, isAdmin: false }),
    true
  )
  assert.equal(
    canUserEditReservationTarget(reservation, { id: 99, isAdmin: false }),
    false
  )
  assert.equal(
    canUserEditReservationTarget(reservation, { id: 99, isAdmin: true }),
    false
  )
})

test('oculta la edición sin capacidad, fuera de plazo o en estado terminal', () => {
  const baseReservation = {
    userId: 7,
    status: 'PENDING',
    canEditTarget: true,
    confirmationDeadline: '2099-01-01T10:00:00'
  }
  const user = { id: 7, isAdmin: false }

  assert.equal(
    canUserEditReservationTarget({ ...baseReservation, canEditTarget: undefined }, user),
    false
  )
  assert.equal(
    canUserEditReservationTarget({ ...baseReservation, confirmationDeadline: '2000-01-01T10:00:00' }, user),
    false
  )
  assert.equal(
    canUserEditReservationTarget({ ...baseReservation, status: 'CANCELLED' }, user),
    false
  )
})
