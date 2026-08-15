import test from 'node:test'
import assert from 'node:assert/strict'

import {
  focusReservationBlock,
  scrollBehaviorFor
} from './reservationFocus.js'

test('usa desplazamiento suave cuando no se solicita reducir movimiento', () => {
  assert.equal(
    scrollBehaviorFor({ matchMedia: () => ({ matches: false }) }),
    'smooth'
  )
})

test('respeta la preferencia de reducir movimiento', () => {
  assert.equal(
    scrollBehaviorFor({ matchMedia: () => ({ matches: true }) }),
    'auto'
  )
})

test('centra y enfoca el bloque de reserva', () => {
  const calls = []
  const element = {
    scrollIntoView: options => calls.push(['scroll', options]),
    focus: options => calls.push(['focus', options])
  }

  assert.equal(
    focusReservationBlock(element, { matchMedia: () => ({ matches: false }) }),
    true
  )
  assert.deepEqual(calls, [
    ['scroll', { behavior: 'smooth', block: 'center', inline: 'center' }],
    ['focus', { preventScroll: true }]
  ])
})

test('no intenta enfocar cuando no existe bloque', () => {
  assert.equal(focusReservationBlock(null), false)
})
