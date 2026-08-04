// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({
  policy: vi.fn(),
  workshopService: {
    getAll: vi.fn(),
    getMine: vi.fn(),
    enroll: vi.fn(),
    withdraw: vi.fn()
  },
  auth: {
    loading: false,
    user: { id: 4, isAdmin: false, rut: '12345678-5' },
    loadAuthUser: vi.fn()
  },
  resources: {
    resources: [{ id: 2, name: 'Cancha 2' }],
    loading: false,
    hasLoaded: true,
    initialLoading: false,
    error: null,
    fetchResources: vi.fn()
  },
  reservations: {
    availabilityReservations: [],
    availabilityLoading: false,
    availabilityHasLoaded: true,
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
    hasLoaded: true,
    initialLoading: false,
    error: null,
    fetchActivities: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({ useAuthStore: () => state.auth }))
vi.mock('@/stores/resources', () => ({ useResourcesStore: () => state.resources }))
vi.mock('@/stores/reservations', () => ({ useReservationsStore: () => state.reservations }))
vi.mock('@/stores/activities', () => ({ useActivitiesStore: () => state.activities }))
vi.mock('@/services/workshops.service', () => ({ workshopsService: state.workshopService }))
vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    getCurrentPolicy: (...args) => state.policy(...args)
  }
}))

import AvailabilitySection from './AvailabilitySection.vue'

const activeWorkshop = {
  id: 3,
  resourceId: 2,
  title: 'Entrenamiento funcional',
  description: 'Taller deportivo',
  location: 'Cancha 2',
  dayText: 'Lunes a domingo',
  scheduleText: '17:00 a 18:00',
  capacity: 20,
  enrolledCount: 9,
  isActive: true,
  isEnrolled: true,
  instructorName: 'Equipo deportivo',
  createdAt: '2026-07-01T12:00:00Z',
  updatedAt: '2026-08-01T12:00:00Z'
}

const ScheduleGridStub = {
  props: ['reservations'],
  emits: ['reservation-selected'],
  computed: {
    workshop() {
      return this.reservations.find((item) => item.isWorkshop)
    }
  },
  template: `
    <div>
      <output class="workshop-block-count">{{ reservations.filter(item => item.isWorkshop).length }}</output>
      <button v-if="workshop" class="open-workshop" type="button" @click="$emit('reservation-selected', workshop)">
        Abrir taller
      </button>
    </div>
  `
}

const findButton = (text) => Array.from(
  document.body.querySelectorAll('button')
).find((button) => button.textContent.includes(text))

const mountAvailability = async () => {
  const wrapper = mount(AvailabilitySection, {
    attachTo: document.body,
    global: {
      plugins: [createPinia()],
      stubs: {
        Teleport: true,
        SkeletonLoader: true,
        CalendarToolbar: true,
        CalendarMini: true,
        AvailabilityTypeLegend: true,
        ScheduleGrid: ScheduleGridStub,
        GeneralCalendarView: true,
        ReservationForm: true
      }
    }
  })
  await flushPromises()
  await wrapper.get('.open-workshop').trigger('click')
  await flushPromises()
  return wrapper
}

describe('AvailabilitySection - inscripcion y desinscripcion de taller', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    state.policy.mockResolvedValue({
      groupResourceIds: [],
      minimumParticipants: 10
    })
    state.auth.loadAuthUser.mockResolvedValue(state.auth.user)
    state.resources.fetchResources.mockResolvedValue(state.resources.resources)
    state.reservations.fetchAvailabilityReservations.mockResolvedValue([])
    state.activities.fetchActivities.mockResolvedValue([])
    state.workshopService.getAll.mockResolvedValue([activeWorkshop])
    state.workshopService.getMine.mockResolvedValue([{
      id: 41,
      workshopId: 3,
      status: 'CONFIRMED',
      isActive: true
    }])
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('muestra desinscripcion con isEnrolled real aunque el historial este rezagado', async () => {
    state.workshopService.getMine.mockResolvedValueOnce([])
    const wrapper = await mountAvailability()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.text()).not.toContain('Inscribirme')
    expect(state.workshopService.getAll).toHaveBeenCalledOnce()
  })

  it('completa alta y baja en el mismo detalle usando las respuestas reales', async () => {
    const enrollmentRequest = {}
    enrollmentRequest.promise = new Promise((resolve) => {
      enrollmentRequest.resolve = resolve
    })
    const notEnrolledWorkshop = {
      ...activeWorkshop,
      isEnrolled: false,
      enrolledCount: 8
    }
    state.workshopService.getAll.mockResolvedValueOnce([notEnrolledWorkshop])
    state.workshopService.getMine.mockReset()
    state.workshopService.getMine
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([{
        id: 42,
        workshopId: 3,
        status: 'CONFIRMED',
        isActive: true
      }])
    state.workshopService.enroll.mockReturnValueOnce(enrollmentRequest.promise)
    const wrapper = await mountAvailability()

    expect(wrapper.text()).toContain('Inscribirme')
    await wrapper.get('.detail-actions .primary').trigger('click')
    await wrapper.get('.detail-actions .primary').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Inscribiendo…')
    expect(state.workshopService.enroll).toHaveBeenCalledOnce()
    expect(state.workshopService.enroll).toHaveBeenCalledWith(3)

    enrollmentRequest.resolve({
      ...notEnrolledWorkshop,
      isEnrolled: true,
      enrolledCount: 9
    })
    await flushPromises()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Inscribirme')
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.text()).toContain('Inscripción registrada correctamente.')
    expect(wrapper.get('.workshop-action-feedback').attributes('role'))
      .toBe('status')
    expect(wrapper.get('.workshop-block-count').text()).toBe('1')
    expect(state.workshopService.getAll).toHaveBeenCalledOnce()
    expect(state.workshopService.getMine).toHaveBeenCalledTimes(2)

    state.workshopService.withdraw.mockResolvedValueOnce({
      workshopId: 3,
      isEnrolled: false,
      enrolledCount: 8,
      changed: true
    })
    await wrapper.get('.detail-actions .danger').trigger('click')
    await findButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(state.workshopService.withdraw).toHaveBeenCalledOnce()
    expect(state.workshopService.withdraw).toHaveBeenCalledWith(3)
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Inscribirme')
    expect(wrapper.text()).not.toContain('Desinscribirme')
    expect(wrapper.get('.workshop-block-count').text()).toBe('1')
  })

  it.each([
    ['RUT', 403, { error: 'Debes registrar tu RUT antes de inscribirte.' }],
    ['capacidad', 409, { error: 'El taller no tiene cupos disponibles.' }],
    ['solape', 409, {
      code: 'WORKSHOP_SCHEDULE_CONFLICT',
      error: 'Ya tienes otro taller en este horario.',
      conflict: { workshopId: 8 }
    }],
    ['inactivo', 404, { error: 'El taller ya no está disponible.' }],
    ['solicitud invalida', 400, { error: 'No se pudo procesar la inscripción.' }],
    ['generico', 500, {}]
  ])('preserva detalle y accion ante error de %s', async (_label, status, data) => {
    state.workshopService.getAll.mockResolvedValueOnce([{
      ...activeWorkshop,
      isEnrolled: false,
      enrolledCount: 8
    }])
    state.workshopService.getMine.mockResolvedValueOnce([])
    state.workshopService.enroll.mockRejectedValueOnce({
      response: { status, data }
    })
    const wrapper = await mountAvailability()

    await wrapper.get('.detail-actions .primary').trigger('click')
    await flushPromises()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Inscribirme')
    expect(wrapper.text()).toContain(
      data.error || 'No se pudo registrar la inscripción.'
    )
    expect(wrapper.get('.workshop-block-count').text()).toBe('1')
    expect(state.workshopService.enroll).toHaveBeenCalledOnce()
  })

  it('cancela la confirmacion sin enviar la baja', async () => {
    const wrapper = await mountAvailability()

    await wrapper.get('.detail-actions .danger').trigger('click')
    expect(document.body.textContent).toContain(
      '¿Desinscribirte de Entrenamiento funcional?'
    )

    await findButton('Mantener inscripción').click()
    await flushPromises()

    expect(state.workshopService.withdraw).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
  })

  it('mantiene el detalle y el bloque, actualiza el cupo y retira la accion sin recargar', async () => {
    let resolveWithdrawal
    state.workshopService.withdraw.mockReturnValueOnce(new Promise((resolve) => {
      resolveWithdrawal = resolve
    }))
    const wrapper = await mountAvailability()

    await wrapper.get('.detail-actions .danger').trigger('click')
    await findButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(wrapper.text()).toContain('Desinscribiendo…')
    expect(state.workshopService.withdraw).toHaveBeenCalledWith(3)

    resolveWithdrawal({
      workshopId: 3,
      isEnrolled: false,
      enrolledCount: 8,
      changed: true
    })
    await flushPromises()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Desinscribirme')
    expect(wrapper.text()).toContain(
      'Te desinscribiste de Entrenamiento funcional. Tu cupo quedó disponible.'
    )
    expect(wrapper.get('.workshop-action-feedback').attributes('role'))
      .toBe('status')
    expect(wrapper.get('.workshop-block-count').text()).toBe('1')
    expect(state.workshopService.getAll).toHaveBeenCalledOnce()
  })

  it('mantiene el detalle y la accion ante un cierre del taller', async () => {
    state.workshopService.withdraw.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { code: 'WORKSHOP_ENROLLMENT_CLOSED' }
      }
    })
    const wrapper = await mountAvailability()

    await wrapper.get('.detail-actions .danger').trigger('click')
    await findButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(wrapper.text()).toContain(
      'Este taller ya no admite cambios en la inscripción.'
    )
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.get('.workshop-block-count').text()).toBe('1')
  })

  it('presenta un error generico seguro sin cerrar el detalle', async () => {
    state.workshopService.withdraw.mockRejectedValueOnce({
      response: { status: 500, data: {} }
    })
    const wrapper = await mountAvailability()

    await wrapper.get('.detail-actions .danger').trigger('click')
    await findButton('Sí, desinscribirme').click()
    await flushPromises()

    expect(wrapper.text()).toContain(
      'No pudimos cancelar tu inscripción. Intenta nuevamente.'
    )
    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
  })
})
