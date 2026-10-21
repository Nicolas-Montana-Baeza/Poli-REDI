import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PolicyAdministrationView from './PolicyAdministrationView.vue'

const mocks = vi.hoisted(() => ({
  resourcesStore: null,
  getHistory: null,
  publish: null
}))

vi.mock('@/stores/resources', () => ({
  useResourcesStore: () => mocks.resourcesStore
}))

vi.mock('@/services/reservationPolicy.service', () => ({
  reservationPolicyService: {
    getHistory: (...args) => mocks.getHistory(...args),
    publish: (...args) => mocks.publish(...args)
  }
}))

const policy = (overrides = {}) => ({
  id: 17,
  reservableWindowDays: 7,
  requestFrequencyDays: 7,
  confirmationDeadlineMinutes: 60,
  minimumParticipants: 10,
  openingMinute: 480,
  closingMinute: 1320,
  slotIntervalMinutes: 15,
  allowedDurations: [30, 60],
  resourceIds: [1, 2],
  groupResources: [
    { resourceId: 1, minimumParticipants: 10 }
  ],
  effectiveFrom: '2026-10-20T16:00:00Z',
  effectiveTo: null,
  createdByUserId: 1,
  createdAt: '2026-10-20T16:00:00Z',
  ...overrides
})

let wrapper

const mountView = async () => {
  wrapper = mount(PolicyAdministrationView)
  await flushPromises()
  return wrapper
}

const click = async (label) => {
  const button = wrapper.findAll('button').find(
    item => item.text().trim() === label
  )

  expect(button, `Botón no encontrado: ${label}`).toBeDefined()

  await button.trigger('click')
  await flushPromises()
}

const changeMinimum = async (value) => {
  await wrapper
    .get('input[aria-label="Mínimo de participantes Cancha 1"]')
    .setValue(String(value))

  await flushPromises()
}

beforeEach(() => {
  mocks.resourcesStore = {
    resources: [
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
    ],
    error: null,
    fetchResources: vi.fn().mockResolvedValue(undefined)
  }

  mocks.getHistory = vi.fn().mockResolvedValue([policy()])
  mocks.publish = vi.fn().mockResolvedValue({})

  vi.stubGlobal('crypto', {
    randomUUID: vi.fn(() => 'policy-test-idempotency-key')
  })
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  vi.unstubAllGlobals()
})

describe('ADMIN-006 — PolicyAdministrationView', () => {
  it('carga la política vigente y sus mínimos grupales', async () => {
    await mountView()

    expect(mocks.getHistory).toHaveBeenCalledOnce()
    expect(mocks.resourcesStore.fetchResources).toHaveBeenCalledOnce()

    expect(wrapper.text()).toContain('Versión #17')
    expect(wrapper.text()).toContain('Cancha 1: 10')

    // El historial ahora tiene su propia pestaña.
    expect(wrapper.text()).not.toContain('Historial de políticas')

    await click('Historial')

    expect(wrapper.text()).toContain('Historial de políticas')
    expect(wrapper.text()).toContain('Versión #17')
  })

  it('no permite publicar si no existen cambios', async () => {
    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await click('Revisar cambios')

    expect(wrapper.text()).toContain(
      'No hay modificaciones respecto a la política vigente.'
    )

    expect(mocks.publish).not.toHaveBeenCalled()
    expect(wrapper.find('.review').exists()).toBe(false)
  })

  it('compara los mínimos grupales antes de publicar', async () => {
    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)
    await click('Revisar cambios')

    expect(wrapper.find('.review').exists()).toBe(true)
    expect(wrapper.text()).toContain('Confirmación grupal por recurso')
    expect(wrapper.text()).toContain('Cancha 1: 12')

    expect(mocks.publish).not.toHaveBeenCalled()
  })

  it('rechaza mínimos superiores a la capacidad', async () => {
    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(23)
    await click('Revisar cambios')

    expect(wrapper.text()).toContain('supera su capacidad')
    expect(wrapper.find('.review').exists()).toBe(false)
    expect(mocks.publish).not.toHaveBeenCalled()
  })

  it('publica solo tras confirmación y usa idempotencia', async () => {
    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)
    await click('Revisar cambios')

    expect(mocks.publish).not.toHaveBeenCalled()

    await click('Confirmar y publicar nueva versión')

    expect(mocks.publish).toHaveBeenCalledOnce()

    expect(mocks.publish).toHaveBeenCalledWith(
      expect.objectContaining({
        expectedPolicyId: 17,
        groupResources: [
          { resourceId: 1, minimumParticipants: 12 }
        ]
      }),
      'policy-test-idempotency-key'
    )

    expect(wrapper.text()).toContain(
      'Nueva política publicada correctamente.'
    )
  })

  it('rechaza una publicación basada en una versión obsoleta', async () => {
    mocks.getHistory
      .mockResolvedValueOnce([policy()])
      .mockResolvedValueOnce([policy({ id: 18 })])

    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)
    await click('Revisar cambios')
    await click('Confirmar y publicar nueva versión')

    expect(mocks.publish).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(
      'La política vigente cambió.'
    )
  })

  it('reutiliza la clave de idempotencia después de un fallo', async () => {
    mocks.publish
      .mockRejectedValueOnce(new Error('Fallo simulado'))
      .mockResolvedValueOnce({})

    await mountView()

    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)
    await click('Revisar cambios')

    await click('Confirmar y publicar nueva versión')

    expect(mocks.publish).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain(
      'No se pudo confirmar la publicación.'
    )

    await click('Confirmar y publicar nueva versión')

    expect(mocks.publish).toHaveBeenCalledTimes(2)

    const firstKey = mocks.publish.mock.calls[0][1]
    const secondKey = mocks.publish.mock.calls[1][1]

    expect(firstKey).toBe(secondKey)

    expect(wrapper.text()).toContain(
      'Nueva política publicada correctamente.'
    )
  })
})


describe('ADMIN-006 — navegación y protección del borrador', () => {
  it('inicia en consulta y entra explícitamente al editor', async () => {
    await mountView()

    expect(wrapper.text()).toContain('Política vigente')
    expect(wrapper.text()).not.toContain('Configura las reglas generales')

    await click('Crear nueva versión')

    expect(wrapper.text()).toContain('Configura las reglas generales')
    expect(wrapper.text()).toContain('1')
    expect(mocks.publish).not.toHaveBeenCalled()
  })

  it('advierte antes de abandonar una edición sin publicar', async () => {
    await mountView()
    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)

    await click('Historial')

    expect(wrapper.text()).toContain('Hay cambios sin publicar')

    await click('Seguir editando')

    expect(wrapper.get(
      'input[aria-label="Mínimo de participantes Cancha 1"]'
    ).element.value).toBe('12')

    expect(mocks.publish).not.toHaveBeenCalled()
  })
})


describe('ADMIN-006 — alcance de recursos sin duplicar configuraciones', () => {
  it('no duplica la habilitación de recursos en políticas', async () => {
    await mountView()
    await click('Crear nueva versión')
    await click('Continuar a recursos')

    expect(
      wrapper.find('input[aria-label="Permitir Cancha 1"]').exists()
    ).toBe(false)

    expect(
      wrapper.find(
        'input[aria-label="Confirmación grupal Cancha 1"]'
      ).exists()
    ).toBe(true)

    expect(
      wrapper.find(
        'input[aria-label="Confirmación grupal Piscina"]'
      ).exists()
    ).toBe(false)
  })

  it('conserva el alcance al publicar cambios grupales', async () => {
    await mountView()
    await click('Crear nueva versión')
    await click('Continuar a recursos')
    await changeMinimum(12)
    await click('Revisar cambios')
    await click('Confirmar y publicar nueva versión')

    expect(mocks.publish).toHaveBeenCalledWith(
      expect.objectContaining({
        resourceIds: [1, 2],
        groupResources: [
          { resourceId: 1, minimumParticipants: 12 }
        ]
      }),
      expect.any(String)
    )
  })
})
