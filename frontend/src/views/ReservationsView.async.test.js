// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({
  auth: null,
  reservations: null
}))

vi.mock('@/stores/auth', async () => {
  const { reactive } = await import('vue')
  state.auth = reactive({
    user: { id: 1, isAdmin: true },
    loadAuthUser: vi.fn()
  })
  return { useAuthStore: () => state.auth }
})

vi.mock('@/stores/reservations', async () => {
  const { reactive } = await import('vue')
  state.reservations = reactive({
    reservations: [],
    myReservations: [],
    loading: false,
    hasLoaded: false,
    initialLoading: false,
    myLoading: false,
    myHasLoaded: false,
    myInitialLoading: false,
    loadingError: null,
    myLoadingError: null,
    actionError: null,
    actionSuccess: null,
    fetchReservations: vi.fn(),
    fetchMyReservations: vi.fn(),
    cancelReservation: vi.fn(),
    clearActionError: vi.fn(),
    clearActionSuccess: vi.fn(),
    setActionError: vi.fn(),
    setActionSuccess: vi.fn()
  })
  return { useReservationsStore: () => state.reservations }
})

vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    updateTarget: vi.fn()
  }
}))

import ReservationsView from './ReservationsView.vue'

const mountReservations = () => mount(ReservationsView, {
  global: {
    stubs: {
      SkeletonLoader: { template: '<div class="skeleton-stub" />' },
      ReservationListCard: {
        props: ['reservation'],
        template: '<div class="reservation-stub">{{ reservation.title }}</div>'
      },
      ReservationDetailModal: true,
      ConfirmModal: true
    }
  }
})

describe('ReservationsView async state', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    state.auth.user = { id: 1, isAdmin: true }
    state.auth.loadAuthUser.mockResolvedValue(state.auth.user)
    state.reservations.fetchReservations.mockResolvedValue([])
    state.reservations.reservations = [{
      id: 9,
      userId: 1,
      title: 'Basquetbol',
      status: 'PENDING',
      startTime: '2099-07-30T18:00:00-04:00',
      durationMinutes: 60
    }]
    state.reservations.loading = false
    state.reservations.hasLoaded = true
    state.reservations.initialLoading = false
    state.reservations.loadingError = null
  })

  it('mantiene el contenido durante refresh o mutaciones y no muestra skeleton global', async () => {
    state.reservations.loading = true
    state.reservations.initialLoading = false

    const wrapper = mountReservations()
    await flushPromises()

    expect(wrapper.find('.skeleton-stub').exists()).toBe(false)
    expect(wrapper.text()).toContain('Basquetbol')
  })
})
