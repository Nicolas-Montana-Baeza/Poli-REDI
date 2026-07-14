import assert from 'node:assert/strict'
import test from 'node:test'

import {
  getLatestReservationStart,
  getReservationScheduleError
} from './reservationRules.js'

test('acepta los limites validos de la jornada', () => {
  assert.equal(getReservationScheduleError({ hour: '08:00', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '21:30', durationMinutes: 30 }), null)
  assert.equal(getReservationScheduleError({ hour: '19:00', durationMinutes: 180 }), null)
})

test('rechaza horas fuera de jornada o fuera del paso', () => {
  assert.equal(getReservationScheduleError({ hour: '07:30', durationMinutes: 30 })?.field, 'hour')
  assert.equal(getReservationScheduleError({ hour: '10:15', durationMinutes: 30 })?.field, 'hour')
  assert.equal(getReservationScheduleError({ hour: '22:00', durationMinutes: 30 })?.field, 'hour')
})

test('rechaza duraciones manipuladas y cierres excedidos', () => {
  assert.equal(getReservationScheduleError({ hour: '10:00', durationMinutes: 45 })?.field, 'durationMinutes')
  assert.equal(getReservationScheduleError({ hour: '21:30', durationMinutes: 60 })?.field, 'durationMinutes')
})

test('calcula la ultima hora segun duracion', () => {
  assert.equal(getLatestReservationStart(30), '21:30')
  assert.equal(getLatestReservationStart(180), '19:00')
})
