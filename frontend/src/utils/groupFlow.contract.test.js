import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

test('rutas join están autenticadas y admiten código manual o URL', () => {
  const router = read('src/router/index.js')
  assert.match(router, /path:\s*'\/join\/:code\?'/)
  assert.match(router, /requiresAuth:\s*true/)
  const view = read('src/views/JoinReservationView.vue')
  assert.match(view, /Código de invitación/)
  assert.match(view, /confirmGroup/)
  assert.match(view, /withdrawGroup/)
})

test('progreso muestra contrato completo y protege retiro del owner', () => {
  const component = read('src/components/ui/ParticipantsProgress.vue')
  for (const label of ['Mínimo requerido', 'Objetivo', 'Capacidad', 'Estado', 'Plazo']) {
    assert.ok(component.includes(label), label)
  }
  assert.match(component, /!props\.progress\.isOwner/)
})

test('código recuperado permanece en estado local del detalle', () => {
  const modal = read('src/components/availability/ReservationDetailModal.vue')
  assert.match(modal, /const joinCode = ref\(''\)/)
  assert.match(modal, /getJoinCode/)
  assert.match(modal, /rotateJoinCode/)
  assert.doesNotMatch(modal, /localStorage|sessionStorage/)
})

test('detalle propio integra progreso, objetivo y código bajo demanda', () => {
  const detail = read('src/views/ReservationDetailView.vue')
  assert.match(detail, /ParticipantsProgress/)
  assert.match(detail, /updateTarget/)
  assert.match(detail, /toggleJoinCode/)
  assert.match(detail, /reservation\.value\?\.userId === authStore\.user\?\.id/)
})
