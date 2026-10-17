import assert from 'node:assert/strict'
import test from 'node:test'

import {
  parsePolicyDurations,
  validatePolicyDraft,
  getPolicyChanges
} from './policyAdministration.js'

const resources = [
  {
    id: 1,
    name: 'Cancha 1',
    isActive: true,
    reservationMode: 'RESERVABLE',
    capacity: 22
  },
  {
    id: 2,
    name: 'Piscina',
    isActive: true,
    reservationMode: 'OPEN_USE',
    capacity: 20
  }
]

const draft = () => ({
  reservableWindowDays: 7,
  requestFrequencyDays: 7,
  confirmationDeadlineMinutes: 60,
  minimumParticipants: 10,
  openingMinute: 480,
  closingMinute: 1320,
  slotIntervalMinutes: 15,
  resourceIds: [1, 2],
  groupResources: [
    { resourceId: 1, minimumParticipants: 10 }
  ]
})

test('normaliza duraciones sin duplicados', () => {
  assert.deepEqual(
    parsePolicyDurations('90, 30, 60, 30'),
    [30, 60, 90]
  )
})

test('rechaza duraciones vacías o inválidas', () => {
  assert.equal(parsePolicyDurations('30, '), null)
  assert.equal(parsePolicyDurations('30, abc'), null)
  assert.equal(parsePolicyDurations('0, 60'), null)
})

test('construye un payload válido completo', () => {
  const result = validatePolicyDraft(
    draft(),
    '60, 30',
    resources
  )

  assert.deepEqual(result.errors, [])
  assert.deepEqual(result.payload.allowedDurations, [30, 60])
  assert.deepEqual(result.payload.groupResources, [
    { resourceId: 1, minimumParticipants: 10 }
  ])
})

test('rechaza mínimos superiores a capacidad', () => {
  const value = draft()
  value.groupResources[0].minimumParticipants = 23

  assert.equal(
    validatePolicyDraft(value, '30, 60', resources).payload,
    null
  )
})

test('rechaza recursos OPEN_USE como grupales', () => {
  const value = draft()
  value.groupResources = [
    { resourceId: 2, minimumParticipants: 10 }
  ]

  assert.equal(
    validatePolicyDraft(value, '30, 60', resources).payload,
    null
  )
})

test('rechaza jornadas inválidas', () => {
  const value = draft()
  value.openingMinute = 1350
  value.closingMinute = 1200

  assert.equal(
    validatePolicyDraft(value, '30, 60', resources).payload,
    null
  )
})

test('detecta cambios por recurso sin modificar el original', () => {
  const original = {
    ...draft(),
    allowedDurations: [30, 60]
  }

  const updated = {
    ...original,
    groupResources: [
      { resourceId: 1, minimumParticipants: 12 }
    ]
  }

  const changes = getPolicyChanges(original, updated, resources)

  assert.equal(changes.length, 1)
  assert.equal(changes[0].label, 'Confirmación grupal por recurso')
  assert.equal(original.groupResources[0].minimumParticipants, 10)
})

test('no detecta cambios cuando los datos son equivalentes', () => {
  const original = {
    ...draft(),
    allowedDurations: [30, 60]
  }

  const next = {
    ...original,
    allowedDurations: [60, 30]
  }

  assert.deepEqual(getPolicyChanges(original, next, resources), [])
})
