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

test('reconoce MVP2 de forma explicita', () => {
  assert.equal(resolveMvpScope('mvp2'), 'mvp2')
  assert.equal(resolveMvpScope(' MVP2 '), 'mvp2')
})

test('reconoce FULL de forma explicita', () => {
  assert.equal(resolveMvpScope('full'), 'full')
  assert.equal(resolveMvpScope(' FULL '), 'full')
})

test('MVP1 mantiene deshabilitadas las funciones posteriores', () => {
  assert.deepEqual(getFeaturesForScope('mvp1'), {
    groupReservations: false,
    schedulingConflictAdministration: false,
    institutionalUnitAdministration: false,
    onlineAuth: false,
    notifications: false,
    workshops: false,
    resourceAdministration: false,
    policyAdministration: false,
    reports: false
  })
})

test('MVP2 habilita reservas grupales y workshops institucionales', () => {
  assert.deepEqual(getFeaturesForScope('mvp2'), {
    groupReservations: true,
    schedulingConflictAdministration: true,
    institutionalUnitAdministration: true,
    onlineAuth: false,
    notifications: false,
    workshops: true,
    resourceAdministration: false,
    policyAdministration: false,
    reports: false
  })
})

test('FULL habilita todas las funcionalidades', () => {
  assert.deepEqual(getFeaturesForScope('full'), {
    groupReservations: true,
    schedulingConflictAdministration: true,
    institutionalUnitAdministration: true,
    onlineAuth: true,
    notifications: true,
    workshops: true,
    resourceAdministration: true,
    policyAdministration: true,
    reports: true
  })
})
