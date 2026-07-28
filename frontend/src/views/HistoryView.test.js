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

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore
}))
vi.mock('@/stores/reservations', () => ({
  useReservationsStore: () => reservationsStore
}))
vi.mock('@/stores/workshops', () => ({
  useWorkshopsStore: () => workshopsStore
}))

import HistoryView from './HistoryView.vue'

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
    const wrapper = mount(HistoryView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' }
        }
      }
    })
    await flushPromises()

    expect(reservationsStore.fetchMyReservations).toHaveBeenCalledOnce()
    expect(workshopsStore.fetchMyEnrollments).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('Escalada')
    expect(wrapper.text()).toContain('Fútbol')
    expect(wrapper.text()).toContain(
      'Para los talleres, el rango considera la fecha de inscripción.'
    )

    await wrapper.findAll('select')[0].setValue('WORKSHOP')
    expect(wrapper.text()).toContain('Escalada')
    expect(wrapper.text()).not.toContain('Fútbol')
  })
})
