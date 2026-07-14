import assert from 'node:assert/strict'
import test from 'node:test'

import {
  getLatestReservationStart,
  getReservationScheduleError,
  snapToReservationSlot
} from './reservationRules.js'

test('acepta los limites validos de la jornada', () => {
  assert.equal(getReservationScheduleError({ hour: '08:00', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '21:30', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '19:00', durationMinutes: 180 }), null)
})

test('rechaza horas fuera de jornada o fuera del paso', () => {
  assert.equal(getReservationScheduleError({ hour: '07:30', durationMinutes: 30 })?.field, 'hour')
  assert.equal(getReservationScheduleError({ hour: '10:10', durationMinutes: 30 })?.field, 'hour')
  assert.equal(getReservationScheduleError({ hour: '22:00', durationMinutes: 30 })?.field, 'hour')
})

test('acepta inicios en cada cuarto de hora', () => {
  assert.equal(getReservationScheduleError({ hour: '10:15', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '10:30', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '10:45', durationMinutes: 30 }), null)
})

test('rechaza duraciones manipuladas y cierres excedidos', () => {
  assert.equal(getReservationScheduleError({ hour: '10:00', durationMinutes: 45 })?.field, 'durationMinutes')
  assert.equal(getReservationScheduleError({ hour: '21:30', durationMinutes: 60 })?.field, 'durationMinutes')
})

test('calcula la ultima hora segun duracion', () => {
  assert.equal(getLatestReservationStart(30), '21:30')
  assert.equal(getLatestReservationStart(180), '19:00')
})

test('ajusta los clics de la linea de tiempo a intervalos de 15 minutos', () => {
  assert.equal(snapToReservationSlot(8 * 60 + 2), 8 * 60)
  assert.equal(snapToReservationSlot(8 * 60 + 8), 8 * 60 + 15)
  assert.equal(snapToReservationSlot(8 * 60 + 16), 8 * 60 + 15)
  assert.equal(snapToReservationSlot(21 * 60 + 29), 21 * 60 + 30)
})
