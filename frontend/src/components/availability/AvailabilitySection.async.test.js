// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({
  policy: vi.fn(),
  auth: {
    loading: false,
    user: { id: 4, isAdmin: false, rut: '12345678-5' },
    loadAuthUser: vi.fn()
  },
  resources: {
    resources: [],
    loading: false,
    hasLoaded: false,
    initialLoading: false,
    error: null,
    fetchResources: vi.fn()
  },
  reservations: {
    availabilityReservations: [],
    availabilityLoading: false,
    availabilityHasLoaded: false,
    availabilityInitialLoading: false,
    availabilityLoadingError: null,
    actionError: null,
    actionSuccess: null,
    fetchAvailabilityReservations: vi.fn(),
    clearActionError: vi.fn(),
    clearActionSuccess: vi.fn(),
    setActionError: vi.fn(),
    setActionSuccess: vi.fn()
  },
  activities: {
    activities: [],
    loading: false,
    hasLoaded: false,
    initialLoading: false,
    error: null,
    fetchActivities: vi.fn()
  },
  workshops: {
    workshops: [],
    myEnrollments: [],
    loading: false,
    hasLoaded: false,
    initialLoading: false,
    historyLoading: false,
    historyHasLoaded: false,
    historyLoadingError: null,
    withdrawingId: null,
    loadingError: null,
    fetchWorkshops: vi.fn(),
    fetchMyEnrollments: vi.fn(),
    clearMessages: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({ useAuthStore: () => state.auth }))
vi.mock('@/stores/resources', () => ({ useResourcesStore: () => state.resources }))
vi.mock('@/stores/reservations', () => ({ useReservationsStore: () => state.reservations }))
vi.mock('@/stores/activities', () => ({ useActivitiesStore: () => state.activities }))
vi.mock('@/stores/workshops', () => ({ useWorkshopsStore: () => state.workshops }))
vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    getCurrentPolicy: (...args) => state.policy(...args)
  }
}))

import AvailabilitySection from './AvailabilitySection.vue'

const deferred = () => {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

const mountAvailability = () => mount(AvailabilitySection, {
  global: {
    stubs: {
      SkeletonLoader: { template: '<div class="skeleton-stub" />' },
      CalendarToolbar: true,
      CalendarMini: true,
      ScheduleGrid: true,
      GeneralCalendarView: true,
      ReservationForm: true,
      ReservationDetailModal: true
    }
  }
})

describe('AvailabilitySection readiness', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    state.auth.loadAuthUser.mockResolvedValue(state.auth.user)
    state.resources.fetchResources.mockResolvedValue([])
    state.reservations.fetchAvailabilityReservations.mockResolvedValue([])
    state.activities.fetchActivities.mockResolvedValue([])
    state.workshops.fetchWorkshops.mockResolvedValue([])
    state.workshops.fetchMyEnrollments.mockResolvedValue([])
    state.resources.error = null
    state.reservations.availabilityLoadingError = null
    state.activities.error = null
    state.workshops.loadingError = null
  })

  it('mantiene el skeleton hasta que finaliza la politica cargada en paralelo', async () => {
    const policyRequest = deferred()
    state.policy.mockReturnValueOnce(policyRequest.promise)

    const wrapper = mountAvailability()
    await flushPromises()

    expect(state.policy).toHaveBeenCalledOnce()
    expect(state.resources.fetchResources).toHaveBeenCalledOnce()
    expect(state.workshops.fetchMyEnrollments).toHaveBeenCalledOnce()
    expect(wrapper.find('.skeleton-stub').exists()).toBe(true)

    policyRequest.resolve({ groupResourceIds: [], minimumParticipants: 10 })
    await flushPromises()

    expect(wrapper.find('.skeleton-stub').exists()).toBe(false)
  })

  it('sale del skeleton con un error visible y mantiene el flujo cerrado', async () => {
    state.policy.mockRejectedValueOnce(new Error('sin politica'))

    const wrapper = mountAvailability()
    await flushPromises()

    expect(wrapper.find('.skeleton-stub').exists()).toBe(false)
    expect(wrapper.text()).toContain('No se pudo validar la política de reservas')
    expect(state.reservations.setActionError).not.toHaveBeenCalled()
  })

  it('no bloquea la disponibilidad por una carga tardia del historial de talleres', async () => {
    const historyRequest = deferred()
    state.policy.mockResolvedValueOnce({
      groupResourceIds: [],
      minimumParticipants: 10
    })
    state.workshops.fetchMyEnrollments.mockReturnValueOnce(historyRequest.promise)

    const wrapper = mountAvailability()
    await flushPromises()

    expect(state.workshops.fetchMyEnrollments).toHaveBeenCalledOnce()
    expect(wrapper.find('.skeleton-stub').exists()).toBe(false)

    historyRequest.resolve([])
    await flushPromises()
  })
})
