import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import InstitutionalOccurrencesPanel from './InstitutionalOccurrencesPanel.vue'

const mocks = vi.hoisted(() => ({ getExceptions: vi.fn(), patchOccurrence: vi.fn() }))
vi.mock('../../services/institutionalActivities.service', () => ({
  institutionalActivitiesService: mocks
}))
const activity = { id: 20, enrollmentCount: 0, schedules: [{
  id: 31, scheduleType: 'WEEKLY', isActive: true, dayOfWeek: 2,
  validFrom: '2026-01-01', validTo: '2027-12-31', startTime: '10:00', endTime: '11:00'
}] }
let wrapper

afterEach(() => { wrapper?.unmount(); wrapper = null; vi.clearAllMocks() })
describe('ADMIN-005 sesiones semanales', () => {
  it('muestra próximas fechas sin generar ocupaciones nuevas', async () => {
    mocks.getExceptions.mockResolvedValue([])
    wrapper = mount(InstitutionalOccurrencesPanel, { props: { activity } })
    await wrapper.get('.toggle').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.session-row').length).toBeGreaterThan(0)
    expect(mocks.patchOccurrence).not.toHaveBeenCalled()
  })

  it('suspende únicamente la fecha elegida y exige motivo', async () => {
    mocks.getExceptions.mockResolvedValue([])
    mocks.patchOccurrence.mockResolvedValue({ id: 88 })
    wrapper = mount(InstitutionalOccurrencesPanel, { props: { activity } })
    await wrapper.get('.toggle').trigger('click')
    await flushPromises()
    await wrapper.get('.session-row .session-actions button').trigger('click')
    await wrapper.get('.session-form textarea').setValue('Suspensión de prueba')
    await wrapper.get('.session-form').trigger('submit')
    await flushPromises()
    expect(mocks.patchOccurrence).toHaveBeenCalledOnce()
    expect(mocks.patchOccurrence.mock.calls[0][0]).toBe(activity.id)
    expect(mocks.patchOccurrence.mock.calls[0][1]).toEqual(expect.objectContaining({
      scheduleId: 31, action: 'CANCEL', reason: 'Suspensión de prueba'
    }))
  })

  it('reprograma solo una fecha y envía nuevo horario', async () => {
    mocks.getExceptions.mockResolvedValue([])
    mocks.patchOccurrence.mockResolvedValue({ id: 89 })
    wrapper = mount(InstitutionalOccurrencesPanel, { props: { activity } })
    await wrapper.get('.toggle').trigger('click')
    await flushPromises()
    await wrapper.findAll('.session-row .session-actions button')[1].trigger('click')
    await wrapper.get('.session-form textarea').setValue('Cambio de horario')
    await wrapper.get('.session-form input[type="time"]').setValue('12:00')
    await wrapper.findAll('.session-form input[type="time"]')[1].setValue('13:00')
    await wrapper.get('.session-form').trigger('submit')
    await flushPromises()
    expect(mocks.patchOccurrence).toHaveBeenCalledWith(activity.id, expect.objectContaining({
      scheduleId: 31, action: 'RESCHEDULE', newStartTime: '12:00', newEndTime: '13:00'
    }))
  })

  it('permite reprogramar por segunda vez una sesión con excepción', async () => {
    const date = new Date()
    date.setUTCHours(12, 0, 0, 0)
    date.setUTCDate(date.getUTCDate() + 14)
    while (date.getUTCDay() !== 2) date.setUTCDate(date.getUTCDate() + 1)
    const originalDate = date.toISOString().slice(0, 10)
    const next = new Date(date)
    next.setUTCDate(next.getUTCDate() + 1)
    const newDate = next.toISOString().slice(0, 10)
    mocks.getExceptions.mockResolvedValue([{
      id: 40, scheduleId: 31, originalDate, action: 'RESCHEDULE',
      newDate, newStartTime: '12:00', newEndTime: '13:00', reason: 'Primera modificación'
    }])
    mocks.patchOccurrence.mockResolvedValue({ id: 40 })
    wrapper = mount(InstitutionalOccurrencesPanel, { props: { activity } })
    await wrapper.get('.toggle').trigger('click')
    await flushPromises()
    const row = wrapper.findAll('.session-row').find(item => item.text().includes(originalDate))
    expect(row).toBeDefined()
    const actions = row.findAll('button')
    expect(actions.length).toBe(2)
    await actions[1].trigger('click')
    expect(wrapper.get('.session-form input[type="date"]').element.value).toBe(newDate)
    expect(wrapper.findAll('.session-form input[type="time"]')[0].element.value).toBe('12:00')
    await wrapper.get('.session-form textarea').setValue('Segundo cambio solicitado')
    await wrapper.get('.session-form').trigger('submit')
    await flushPromises()
    expect(mocks.patchOccurrence).toHaveBeenCalledWith(activity.id, expect.objectContaining({
      scheduleId: 31, originalDate, action: 'RESCHEDULE', reason: 'Segundo cambio solicitado'
    }))
  })

  it('impide edición cuando hay inscritos', async () => {
    mocks.getExceptions.mockResolvedValue([])
    wrapper = mount(InstitutionalOccurrencesPanel, { props: { activity: { ...activity, enrollmentCount: 2 } } })
    await wrapper.get('.toggle').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('inscripciones vigentes')
    expect(wrapper.find('.session-actions').exists()).toBe(false)
  })
})
