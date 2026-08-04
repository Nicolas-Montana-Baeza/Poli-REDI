import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const services = vi.hoisted(() => ({
  resources: {
    getAll: vi.fn(),
    updateImage: vi.fn()
  },
  reservations: {
    getAll: vi.fn(),
    getAvailability: vi.fn(),
    getMine: vi.fn(),
    create: vi.fn(),
    cancel: vi.fn()
  },
  workshops: {
    getAll: vi.fn(),
    getMine: vi.fn(),
    enroll: vi.fn(),
    withdraw: vi.fn()
  }
}))

vi.mock('@/services/resources.service', () => ({
  resourcesService: services.resources
}))
vi.mock('@/services/reservations.service', () => ({
  reservationsService: services.reservations
}))
vi.mock('@/services/workshops.service', () => ({
  workshopsService: services.workshops
}))

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import { useWorkshopsStore } from '@/stores/workshops'

const deferred = () => {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

describe('contrato asincrono de stores', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('deduplica recursos y conserva datos ante un error de refresco', async () => {
    const firstRequest = deferred()
    services.resources.getAll.mockReturnValueOnce(firstRequest.promise)
    const store = useResourcesStore()

    const first = store.fetchResources()
    const duplicate = store.fetchResources()

    expect(services.resources.getAll).toHaveBeenCalledOnce()
    expect(store.initialLoading).toBe(true)

    firstRequest.resolve([{ id: 1, name: 'Cancha 1' }])
    await Promise.all([first, duplicate])

    expect(store.hasLoaded).toBe(true)
    expect(store.resources).toEqual([{ id: 1, name: 'Cancha 1' }])

    const refreshRequest = deferred()
    services.resources.getAll.mockReturnValueOnce(refreshRequest.promise)
    const refresh = store.fetchResources()

    expect(store.initialLoading).toBe(false)
    expect(store.refreshing).toBe(true)
    expect(store.resources).toHaveLength(1)

    refreshRequest.reject(new Error('sin red'))
    await refresh

    expect(store.resources).toEqual([{ id: 1, name: 'Cancha 1' }])
    expect(store.status).toBe('error')
    expect(store.error).toBe('No se pudieron cargar los recursos')
  })

  it('ignora una respuesta forzada tardia y conserva la mas reciente', async () => {
    const oldRequest = deferred()
    const newRequest = deferred()
    services.resources.getAll
      .mockReturnValueOnce(oldRequest.promise)
      .mockReturnValueOnce(newRequest.promise)
    const store = useResourcesStore()

    const oldFetch = store.fetchResources()
    const newFetch = store.fetchResources({ force: true })

    newRequest.resolve([{ id: 2, name: 'Nueva' }])
    await newFetch
    oldRequest.resolve([{ id: 1, name: 'Antigua' }])
    await oldFetch

    expect(store.resources).toEqual([{ id: 2, name: 'Nueva' }])
    expect(store.loading).toBe(false)
  })

  it('separa crear y cancelar de las cargas de consulta', async () => {
    const createRequest = deferred()
    const cancelRequest = deferred()
    services.reservations.create.mockReturnValueOnce(createRequest.promise)
    services.reservations.cancel.mockReturnValueOnce(cancelRequest.promise)
    const store = useReservationsStore()

    const creating = store.createReservation({ resourceId: 1 })

    expect(store.creating).toBe(true)
    expect(store.loading).toBe(false)
    expect(store.initialLoading).toBe(false)

    createRequest.resolve({ id: 8, resourceId: 1, status: 'PENDING' })
    await creating

    const cancelling = store.cancelReservation(8)

    expect(store.cancellingId).toBe(8)
    expect(store.loading).toBe(false)
    expect(store.myLoading).toBe(false)

    cancelRequest.resolve({ id: 8, resourceId: 1, status: 'CANCELLED' })
    await cancelling

    expect(store.cancellingId).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('deduplica el historial de talleres y preserva el ultimo exito', async () => {
    const initialRequest = deferred()
    services.workshops.getMine.mockReturnValueOnce(initialRequest.promise)
    const store = useWorkshopsStore()

    const first = store.fetchMyEnrollments()
    const duplicate = store.fetchMyEnrollments()

    expect(services.workshops.getMine).toHaveBeenCalledOnce()

    initialRequest.resolve([{ id: 3, title: 'Esgrima' }])
    await Promise.all([first, duplicate])

    const refreshRequest = deferred()
    services.workshops.getMine.mockReturnValueOnce(refreshRequest.promise)
    const refresh = store.fetchMyEnrollments()
    refreshRequest.reject(new Error('sin red'))
    await refresh

    expect(store.myEnrollments).toEqual([{ id: 3, title: 'Esgrima' }])
    expect(store.historyStatus).toBe('error')
    expect(store.historyInitialLoading).toBe(false)
  })

  it('desinscribe una sola vez y conserva los episodios del historial', async () => {
    const withdrawalRequest = deferred()
    services.workshops.withdraw.mockReturnValueOnce(withdrawalRequest.promise)
    const store = useWorkshopsStore()
    store.workshops = [{
      id: 3,
      title: 'Esgrima',
      isActive: true,
      isEnrolled: true,
      enrolledCount: 9,
      capacity: 20
    }]
    store.myEnrollments = [
      { id: 40, workshopId: 3, status: 'CANCELLED' },
      { id: 41, workshopId: 3, status: 'CONFIRMED' },
      { id: 42, workshopId: 8, status: 'CONFIRMED' }
    ]

    const first = store.withdraw(3)
    const duplicate = store.withdraw(3)

    expect(store.withdrawingId).toBe(3)
    expect(services.workshops.withdraw).toHaveBeenCalledOnce()
    await expect(duplicate).resolves.toBeNull()

    withdrawalRequest.resolve({
      id: 3,
      title: 'Esgrima',
      isActive: true,
      isEnrolled: false,
      enrolledCount: 8,
      capacity: 20
    })
    await first

    expect(store.withdrawingId).toBeNull()
    expect(store.workshops[0]).toMatchObject({
      isEnrolled: false,
      enrolledCount: 8
    })
    expect(store.myEnrollments.map((item) => item.id))
      .toEqual([40, 41, 42])
    expect(store.myEnrollments.map((item) => item.status))
      .toEqual(['CANCELLED', 'CANCELLED', 'CONFIRMED'])
    expect(store.actionSuccess)
      .toBe('Te desinscribiste de Esgrima. Tu cupo quedó disponible.')
  })

  it('conserva talleres e historial si falla la desinscripcion', async () => {
    services.workshops.withdraw.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { code: 'WORKSHOP_ENROLLMENT_CLOSED' }
      }
    })
    const store = useWorkshopsStore()
    const workshop = {
      id: 3,
      isActive: true,
      isEnrolled: true,
      enrolledCount: 9
    }
    const enrollment = {
      id: 41,
      workshopId: 3,
      status: 'CONFIRMED'
    }
    store.workshops = [workshop]
    store.myEnrollments = [enrollment]

    await expect(store.withdraw(3)).rejects.toThrow(
      'Este taller ya no admite cambios en la inscripción.'
    )

    expect(store.withdrawingId).toBeNull()
    expect(store.workshops[0]).toEqual(workshop)
    expect(store.myEnrollments[0]).toEqual(enrollment)
    expect(store.actionError).toEqual({
      message: 'Este taller ya no admite cambios en la inscripción.'
    })
  })
})
