import assert from 'node:assert/strict'
import test from 'node:test'

import {
  getFeaturesForScope,
  resolveMvpScope
} from './appScope.js'

test('usa MVP1 como alcance seguro por defecto', () => {
  assert.equal(resolveMvpScope(), 'mvp1')
  assert.equal(resolveMvpScope('unexpected'), 'mvp1')
})

test('solo habilita el alcance completo de forma explicita', () => {
  assert.equal(resolveMvpScope('full'), 'full')
  assert.equal(resolveMvpScope(' FULL '), 'full')
})

test('MVP1 no habilita autenticacion online ni llamadas posteriores', () => {
  assert.deepEqual(getFeaturesForScope('mvp1'), {
    onlineAuth: false,
    notifications: false,
    workshops: false,
    groupReservations: false,
    resourceAdministration: false,
    policyAdministration: false,
    reports: false
  })
})
