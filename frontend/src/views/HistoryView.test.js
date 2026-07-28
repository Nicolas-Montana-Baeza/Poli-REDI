// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const authStore = {
  user: { id: 4, isAdmin: false },
  loadAuthUser: vi.fn()
}
const reservationsStore = {
  myReservations: [],
  reservations: [],
  myLoading: false,
  loading: false,
  myLoadingError: null,
  loadingError: null,
  fetchMyReservations: vi.fn(),
  fetchReservations: vi.fn()
}
const workshopsStore = {
  myEnrollments: [],
  historyLoading: false,
  historyLoadingError: null,
  fetchMyEnrollments: vi.fn()
}

vi.mock('@/stores/auth', () => ({ useAuthStore: () => authStore }))
vi.mock('@/stores/reservations', () => ({ useReservationsStore: () => reservationsStore }))
vi.mock('@/stores/workshops', () => ({ useWorkshopsStore: () => workshopsStore }))

import HistoryView from './HistoryView.vue'

const mountHistory = (options = {}) => mount(HistoryView, {
  ...options,
  global: {
    ...options.global,
    stubs: {
      Teleport: true,
      RouterLink: true,
      ...options.global?.stubs
    }
  }
})

describe('HistoryView', () => {
  beforeEach(() => {
    authStore.user = { id: 4, isAdmin: false }
    authStore.loadAuthUser.mockReset().mockResolvedValue(authStore.user)
    reservationsStore.fetchMyReservations.mockReset().mockResolvedValue()
    workshopsStore.fetchMyEnrollments.mockReset().mockResolvedValue()
    reservationsStore.myReservations = [{
      id: 1,
      title: 'Fútbol',
      resourceName: 'Cancha 1',
      startTime: '2026-07-10T10:00:00-04:00',
      durationMinutes: 60,
      status: 'CANCELLED'
    }]
    workshopsStore.myEnrollments = [{
      id: 5,
      workshopId: 2,
      title: 'Escalada',
      location: 'Gimnasio',
      dayText: 'Martes',
      scheduleText: '18:00 - 19:00',
      instructorName: 'Ana',
      status: 'CONFIRMED',
      isActive: true,
      enrolledAt: '2026-07-20T15:00:00Z'
    }]
  })

  it('combina reservas e inscripciones y permite filtrar por tipo', async () => {
    const wrapper = mountHistory()
    await flushPromises()

    expect(reservationsStore.fetchMyReservations).toHaveBeenCalledOnce()
    expect(workshopsStore.fetchMyEnrollments).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('Escalada')
    expect(wrapper.text()).toContain('Fútbol')
    expect(wrapper.text()).toContain('Para los talleres, el rango considera la fecha de inscripción.')

    await wrapper.findAll('select')[0].setValue('WORKSHOP')
    expect(wrapper.text()).toContain('Escalada')
    expect(wrapper.text()).not.toContain('Fútbol')
  })

  it('abre la tarjeta completa de reserva en el modal compartido y restaura el foco', async () => {
    const wrapper = mountHistory({ attachTo: document.body })
    await flushPromises()
    const card = wrapper.get('[data-history-id="reservation-1"]')

    expect(card.element.tagName).toBe('BUTTON')
    card.element.focus()
    await card.trigger('click')

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.text()).toContain('Detalle de reserva')
    expect(dialog.text()).toContain('Fútbol')
    expect(dialog.text()).not.toContain('Cancelar reserva')
    expect(dialog.text()).not.toContain('Consultar código')
    expect(dialog.text()).not.toContain('Guardar cambios')

    await wrapper.get('.detail-actions button').trigger('click')
    await flushPromises()
    expect(document.activeElement).toBe(card.element)
    wrapper.unmount()
  })

  it('abre una inscripción como taller institucional en modo solo lectura', async () => {
    const wrapper = mountHistory()
    await flushPromises()
    const card = wrapper.get('[data-history-id="workshop-enrollment-5"]')

    expect(card.element.tagName).toBe('BUTTON')
    await card.trigger('click')

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.text()).toContain('Actividad institucional')
    expect(dialog.text()).toContain('Escalada')
    expect(dialog.text()).toContain('Martes')
    expect(dialog.text()).toContain('18:00 - 19:00')
    expect(dialog.text()).toContain('Fecha de inscripción')
    expect(dialog.text()).not.toContain('Objetivo de participantes')
    expect(dialog.text()).not.toContain('Cancelar reserva')
  })
})
