import assert from 'node:assert/strict'
import test from 'node:test'

import {
  formatBusinessTime,
  getBusinessDateKey,
  parseBusinessDateTime
} from './businessTime.js'

test('interpreta una hora sin offset como America/Santiago', () => {
  const winter = parseBusinessDateTime('2026-07-14T10:30:00')
  const summer = parseBusinessDateTime('2026-01-14T10:30:00')

  assert.equal(winter.toISOString(), '2026-07-14T14:30:00.000Z')
  assert.equal(summer.toISOString(), '2026-01-14T13:30:00.000Z')
})

test('respeta un offset explicito y muestra la hora de Santiago', () => {
  const date = parseBusinessDateTime('2026-07-14T14:30:00Z')

  assert.equal(getBusinessDateKey(date), '2026-07-14')
  assert.equal(formatBusinessTime(date), '10:30')
})

test('clasifica correctamente un instante UTC cercano a medianoche', () => {
  const date = parseBusinessDateTime('2026-07-15T02:30:00Z')

  assert.equal(getBusinessDateKey(date), '2026-07-14')
  assert.equal(formatBusinessTime(date), '22:30')
})
