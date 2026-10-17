import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import VenuesAdminPanel from './VenuesAdminPanel.vue'

const state = vi.hoisted(() => ({
  store: null
}))

vi.mock('@/stores/venues', () => ({
  useVenuesStore: () => state.store
}))

const existingVenue = () => ({
  id: 7,
  name: 'Sede Central',
  country: 'Chile',
  addressLine: 'Dirección anterior',
  commune: 'Santiago',
  city: 'Santiago',
  region: 'Metropolitana',
  latitude: -33.45,
  longitude: -70.66,
  isActive: true,
  resourceCount: 2,
  activeResourceCount: 2
})

describe('VenuesAdminPanel', () => {
  beforeEach(() => {
    state.store = {
      venues: [existingVenue()],
      loading: false,
      saving: false,
      error: null,
      actionError: null,
      actionSuccess: null,
      fetchVenues: vi.fn().mockResolvedValue([]),
      createVenue: vi.fn().mockResolvedValue({}),
      updateVenue: vi.fn().mockResolvedValue({})
    }
  })

  it('prioriza el listado y abre la creación bajo demanda', async () => {
    const wrapper = mount(VenuesAdminPanel)
    expect(wrapper.find('form.venue-form').exists()).toBe(false)
    expect(wrapper.findAll('.venue-item').length).toBe(1)
    await wrapper.find('.venues-heading button').trigger('click')
    expect(wrapper.find('form.venue-form').exists()).toBe(true)
  })

  it('crea una sede con ubicación y coordenadas', async () => {
    const wrapper = mount(VenuesAdminPanel)
    await wrapper.find('.venues-heading button').trigger('click')

    const form = wrapper.find('form.venue-form')
    const inputs = form.findAll('input')

    await inputs[0].setValue('Sede Norte')
    await inputs[1].setValue('Chile')
    await inputs[2].setValue('Av. Principal 123')
    await inputs[3].setValue('Independencia')
    await inputs[4].setValue('Santiago')
    await inputs[5].setValue('Metropolitana')
    await inputs[6].setValue('-33.4489')
    await inputs[7].setValue('-70.6693')

    await form.trigger('submit')
    await flushPromises()

    expect(state.store.createVenue).toHaveBeenCalledWith({
      name: 'Sede Norte',
      country: 'Chile',
      addressLine: 'Av. Principal 123',
      commune: 'Independencia',
      city: 'Santiago',
      region: 'Metropolitana',
      latitude: -33.4489,
      longitude: -70.6693
    })
  })

  it('rechaza una coordenada incompleta', async () => {
    const wrapper = mount(VenuesAdminPanel)
    await wrapper.find('.venues-heading button').trigger('click')

    const form = wrapper.find('form.venue-form')
    const inputs = form.findAll('input')

    await inputs[0].setValue('Sede Norte')
    await inputs[6].setValue('-33.4489')
    await inputs[7].setValue('')

    await form.trigger('submit')
    await flushPromises()

    expect(state.store.createVenue).not.toHaveBeenCalled()

    expect(wrapper.text()).toContain(
      'Debes ingresar ambas coordenadas'
    )
  })

  it('edita una sede y permite limpiar las coordenadas', async () => {
    const wrapper = mount(VenuesAdminPanel)

    const editButton = wrapper.findAll('button').find(
      button => button.text() === 'Editar sede'
    )

    expect(editButton).toBeDefined()
    await editButton.trigger('click')

    const editForm = wrapper.find('.venue-item form.venue-form')
    const inputs = editForm.findAll('input')

    await inputs[0].setValue('Sede Central Renovada')
    await inputs[2].setValue('Nueva dirección 456')
    await inputs[6].setValue('')
    await inputs[7].setValue('')

    await editForm.trigger('submit')
    await flushPromises()

    expect(state.store.updateVenue).toHaveBeenCalledWith(
      7,
      expect.objectContaining({
        name: 'Sede Central Renovada',
        addressLine: 'Nueva dirección 456',
        latitude: null,
        longitude: null,
        isActive: true
      })
    )
  })
})
