import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import InstitutionalActivitiesView from './InstitutionalActivitiesView.vue'

const mocks = vi.hoisted(() => ({
  store: null,
  getUnits: vi.fn(),
  getResources: vi.fn(),
  getPolicy: vi.fn()
}))

vi.mock('../services/institutionalUnits.service', () => ({
  institutionalUnitsService: {
    getAll: mocks.getUnits
  }
}))

vi.mock('../services/resources.service', () => ({
  resourcesService: {
    getAll: mocks.getResources
  }
}))

vi.mock('../services/reservationPolicy.service', () => ({
  reservationPolicyService: { getCurrent: mocks.getPolicy }
}))

vi.mock('../stores/institutionalActivities', () => ({
  useInstitutionalActivitiesStore: () => mocks.store
}))

let wrapper

beforeEach(() => {
  mocks.getUnits.mockReset()
  mocks.getResources.mockReset()
  mocks.getPolicy.mockReset()
  mocks.getPolicy.mockResolvedValue({ openingMinute: 480, closingMinute: 1320 })

  mocks.getUnits.mockResolvedValue([
    { id: 4, name: 'Educación Física', code: 'EFI' }
  ])

  mocks.getResources.mockResolvedValue([
    { id: 8, name: 'Cancha 1', isActive: true },
    { id: 9, name: 'Cancha inactiva', isActive: false }
  ])

  mocks.store = reactive({
    activitiesByUnit: {},
    loading: false,
    creating: false,
    saving: false,
    cancellingId: null,
    error: null,

    clearError: vi.fn(() => {
      mocks.store.error = null
    }),

    loadByUnit: vi.fn().mockResolvedValue([]),

    create: vi.fn(async payload => ({
      id: 100,
      title: payload.title
    })),
    update: vi.fn(async (_id,payload) => ({id:100,title:payload.title})),
    cancel: vi.fn(async id => ({id,status:'CANCELLED'}))
  })
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
})

const mountView = async () => {
  wrapper = mount(InstitutionalActivitiesView)
  await flushPromises()
  return wrapper
}

const setTitle = async (title) => {
  await wrapper
    .get('.activity-form input[type="text"]')
    .setValue(title)
}

describe('ADMIN-005: programación institucional', () => {
  it('carga unidades y recursos activos', async () => {
    await mountView()

    expect(mocks.getUnits).toHaveBeenCalledOnce()
    expect(mocks.getResources).toHaveBeenCalledOnce()
    expect(mocks.store.loadByUnit).toHaveBeenCalledWith(4)

    expect(wrapper.text()).toContain('Educación Física')
    expect(wrapper.text()).toContain('Cancha 1')
    expect(wrapper.text()).not.toContain('Cancha inactiva')
  })

  it('crea programación SINGLE con los datos correctos', async () => {
    await mountView()
    await setTitle('Clase institucional')

    await wrapper
      .get('.schedule-grid input[type="date"]')
      .setValue('2026-11-02')

    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.create).toHaveBeenCalledWith({
      unitId: 4,
      resourceId: 8,
      activityType: 'ACADEMIC_CLASS',
      title: 'Clase institucional',
      description: '',
      requiresEnrollment: false,
      schedules: [{
        scheduleType: 'SINGLE',
        specificDate: '2026-11-02',
        startTime: '10:00',
        endTime: '11:00'
      }]
    })

    expect(wrapper.text()).toContain('creada correctamente')
  })

  it('crea programación WEEKLY con día y vigencia', async () => {
    await mountView()
    await setTitle('Entrenamiento semanal')

    await wrapper
      .get('.schedule-type input[value="WEEKLY"]')
      .setValue()

    await nextTick()

    await wrapper
      .get('.schedule-grid select')
      .setValue('3')

    const dates = wrapper.findAll(
      '.schedule-grid input[type="date"]'
    )

    await dates[0].setValue('2026-11-02')
    await dates[1].setValue('2026-11-30')

    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.create).toHaveBeenCalledWith(
      expect.objectContaining({
        schedules: [{
          scheduleType: 'WEEKLY',
          dayOfWeek: 3,
          startTime: '10:00',
          endTime: '11:00',
          validFrom: '2026-11-02',
          validTo: '2026-11-30'
        }]
      })
    )
  })

  it('exige inscripción y capacidad para WORKSHOP', async () => {
    await mountView()

    await wrapper
      .findAll('.activity-form select')[0]
      .setValue('WORKSHOP')

    await nextTick()

    const checkbox = wrapper.get(
      '.activity-form input[type="checkbox"]'
    )

    expect(checkbox.element.checked).toBe(true)
    expect(checkbox.element.disabled).toBe(true)

    expect(
      wrapper.get('.activity-form input[type="number"]').element.value
    ).toBe('20')

    await setTitle('Taller deportivo')

    await wrapper
      .get('.schedule-grid input[type="date"]')
      .setValue('2026-11-02')

    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.create).toHaveBeenCalledWith(
      expect.objectContaining({
        activityType: 'WORKSHOP',
        requiresEnrollment: true,
        capacity: 20
      })
    )
  })

  it('conserva el formulario si la API rechaza la creación', async () => {
    mocks.store.create = vi.fn(async () => {
      mocks.store.error = 'Conflicto de disponibilidad simulado'
      throw new Error('API rechazada')
    })

    await mountView()
    await setTitle('Actividad rechazada')

    await wrapper
      .get('.schedule-grid input[type="date"]')
      .setValue('2026-11-02')

    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain(
      'Conflicto de disponibilidad simulado'
    )

    expect(
      wrapper.get('.activity-form input[type="text"]').element.value
    ).toBe('Actividad rechazada')

    expect(wrapper.find('.message.success').exists()).toBe(false)
  })

  it('permite editar programación SINGLE existente', async () => {
    mocks.store.activitiesByUnit = {
      4: [{
        id: 100, unitId: 4, resourceId: 8, activityType: 'ACADEMIC_CLASS',
        title: 'Clase inicial', description: '', status: 'SCHEDULED',
        requiresEnrollment: false, enrollmentCount: 0,
        schedules: [{ id: 1, scheduleType: 'SINGLE',
          specificDate: '2026-11-02', startTime: '10:00', endTime: '11:00' }]
      }]
    }
    await mountView()
    await wrapper.get('.activity-card .secondary-button').trigger('click')
    await setTitle('Clase actualizada')
    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()
    expect(mocks.store.update).toHaveBeenCalledWith(100,
      expect.objectContaining({ title: 'Clase actualizada',
        schedules: [{scheduleType:'SINGLE',specificDate:'2026-11-02',
          startTime:'10:00',endTime:'11:00'}]
      }))
  })

  it('exige confirmación antes de cancelar una actividad', async () => {
    mocks.store.activitiesByUnit = {
      4: [{
        id: 101, unitId: 4, resourceId: 8, activityType: 'EVENT',
        title: 'Evento de prueba', status: 'SCHEDULED',
        enrollmentCount: 0, schedules: []
      }]
    }
    await mountView()
    await wrapper.get('.activity-card .danger-button').trigger('click')
    expect(mocks.store.cancel).not.toHaveBeenCalled()
    await wrapper.get('.cancel-confirm .danger-button').trigger('click')
    await flushPromises()
    expect(mocks.store.cancel).toHaveBeenCalledWith(101, 4)
  })


  it('filtra actividades por estado y nombre sin perder el historial', async () => {
    mocks.store.activitiesByUnit = {
      4: [
        { id: 150, unitId: 4, resourceId: 8, activityType: 'TRAINING',
          title: 'Entrenamiento de fútbol', status: 'SCHEDULED',
          schedules: [], enrollmentCount: 0 },
        { id: 151, unitId: 4, resourceId: 8, activityType: 'EVENT',
          title: 'Evento cultural', status: 'CANCELLED',
          schedules: [], enrollmentCount: 0 }
      ]
    }
    await mountView()

    expect(wrapper.findAll('.activity-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('Entrenamiento de fútbol')

    await wrapper.get('.status-filters button:nth-child(2)').trigger('click')
    expect(wrapper.findAll('.activity-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('Evento cultural')
    expect(wrapper.get('.status-badge').text()).toBe('Cancelada')

    await wrapper.get('.status-filters button:nth-child(3)').trigger('click')
    await wrapper.get('.search-field input').setValue('futbol')
    expect(wrapper.findAll('.activity-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('Entrenamiento de fútbol')

    await wrapper.get('.search-field input').setValue('sin resultados')
    expect(wrapper.findAll('.activity-card')).toHaveLength(0)
    expect(wrapper.text()).toContain('No encontramos coincidencias')
  })

  it('rechaza un taller fuera de la jornada vigente sin llamar al backend', async () => {
    await mountView()
    await setTitle('Taller jornada inválida')
    await wrapper.get('.schedule-grid input[type="date"]').setValue('2026-11-02')
    await wrapper.findAll('.schedule-grid input[type="time"]')[0].setValue('06:00')
    await wrapper.findAll('.schedule-grid input[type="time"]')[1].setValue('23:00')
    await wrapper.get('.activity-form').trigger('submit')
    await flushPromises()
    expect(mocks.store.create).not.toHaveBeenCalled()
    expect(mocks.store.error).toContain('08:00–22:00')
  })

  it('no habilita creación sin unidades institucionales', async () => {
    mocks.getUnits.mockResolvedValue([])

    await mountView()

    expect(wrapper.text()).toContain('Sin unidades disponibles')
    expect(wrapper.find('.activity-form').exists()).toBe(false)
    expect(mocks.store.create).not.toHaveBeenCalled()
  })
})
