// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const service = vi.hoisted(() => ({
  getAll: vi.fn(),
  getMine: vi.fn(),
  enroll: vi.fn(),
  withdraw: vi.fn()
}))

vi.mock('@/services/workshops.service', () => ({
  workshopsService: service
}))

import WorkshopsView from './WorkshopsView.vue'

const workshop = (overrides = {}) => ({
  id: 3,
  title: 'Esgrima',
  description: 'Taller deportivo',
  location: 'Cancha 1',
  dayText: 'Martes',
  scheduleText: '19:00 - 21:00',
  capacity: 20,
  enrolledCount: 9,
  isActive: true,
  isEnrolled: true,
  ...overrides
})

const findBodyButton = (text) => Array.from(
  document.body.querySelectorAll('button')
).find((button) => button.textContent.includes(text))

const mountView = async (workshops = [workshop()]) => {
  service.getAll.mockResolvedValueOnce(workshops)
  const wrapper = mount(WorkshopsView, {
    attachTo: document.body,
    global: { plugins: [createPinia()] }
  })
  await flushPromises()
  return wrapper
}

describe('WorkshopsView - desinscripcion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('pide confirmacion y cancelar no envia la solicitud', async () => {
    const wrapper = await mountView()

    await wrapper.get('button.withdraw').trigger('click')

    expect(document.body.textContent).toContain('¿Desinscribirte de Esgrima?')
    expect(document.body.textContent).toContain(
      'Podrás volver a inscribirte solo si el taller sigue activo, tiene cupos y no se cruza con otro taller.'
    )

    await findBodyButton('Mantener inscripción').click()
    await flushPromises()

    expect(service.withdraw).not.toHaveBeenCalled()
    expect(document.body.textContent).not.toContain('¿Desinscribirte de Esgrima?')
  })

  it('desinscribe, actualiza el cupo y permite volver a inscribirse', async () => {
    const wrapper = await mountView()
    service.withdraw.mockResolvedValueOnce(workshop({
      isEnrolled: false,
      enrolledCount: 8
    }))

    await wrapper.get('button.withdraw').trigger('click')
    await findBodyButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(service.withdraw).toHaveBeenCalledOnce()
    expect(service.withdraw).toHaveBeenCalledWith(3)
    expect(wrapper.text()).toContain('8 / 20 inscritos')
    expect(wrapper.text()).toContain('Inscribirme')
    expect(wrapper.text()).toContain(
      'Te desinscribiste de Esgrima. Tu cupo quedó disponible.'
    )
  })

  it('muestra el error de cierre y conserva la inscripcion', async () => {
    const wrapper = await mountView()
    service.withdraw.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { code: 'WORKSHOP_ENROLLMENT_CLOSED' }
      }
    })

    await wrapper.get('button.withdraw').trigger('click')
    await findBodyButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(wrapper.text()).toContain(
      'Este taller ya no admite cambios en la inscripción.'
    )
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.text()).toContain('9 / 20 inscritos')
  })

  it('no muestra talleres inactivos ni acciones de inscripcion', async () => {
    const wrapper = await mountView([
      workshop({ isActive: false })
    ])

    expect(wrapper.text()).toContain(
      'No hay talleres disponibles con esos filtros.'
    )
    expect(wrapper.text()).not.toContain('Esgrima')
    expect(wrapper.find('button.withdraw').exists()).toBe(false)
  })
})
