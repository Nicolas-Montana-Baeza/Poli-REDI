import assert from 'node:assert/strict'
import test from 'node:test'

import {
  getReservationCancellationMessage
} from './reservationTime.js'

test('explica la cancelación por mínimo no alcanzado', () => {
  assert.equal(
    getReservationCancellationMessage({
      status: 'CANCELLED',
      cancellationReason: 'MINIMUM_NOT_MET'
    }),
    'La reserva se canceló porque no alcanzó el mínimo de participantes antes del plazo.'
  )
})

test('no inventa un motivo cuando no corresponde', () => {
  assert.equal(
    getReservationCancellationMessage({
      status: 'CONFIRMED',
      cancellationReason: 'MINIMUM_NOT_MET'
    }),
    ''
  )

  assert.equal(
    getReservationCancellationMessage({
      status: 'CANCELLED'
    }),
    ''
  )
})
