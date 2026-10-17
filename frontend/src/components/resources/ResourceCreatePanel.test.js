import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ResourceCreatePanel from './ResourceCreatePanel.vue'

const stores = vi.hoisted(() => ({
  resources: null,
  venues: null
}))

vi.mock('@/stores/resources', () => ({
  useResourcesStore: () => stores.resources
}))

vi.mock('@/stores/venues', () => ({
  useVenuesStore: () => stores.venues
}))

describe('ResourceCreatePanel', () => {
  beforeEach(() => {
    stores.resources = {
      creatingResource: false,
      actionError: null,
      actionSuccess: null,
      createResource: vi.fn().mockResolvedValue({})
    }

    stores.venues = {
      venues: [
        { id: 7, name: 'Sede de prueba', isActive: true }
      ],
      loading: false,
      error: null,
      fetchVenues: vi.fn().mockResolvedValue([])
    }
  })

  it('crea un recurso asociado a la sede seleccionada', async () => {
    const wrapper = mount(ResourceCreatePanel)
    const form = wrapper.find('form')

    const selects = form.findAll('select')
    const inputs = form.findAll('input')

    await selects[0].setValue('7')
    await inputs[0].setValue('Cancha de prueba')
    await inputs[1].setValue('Cancha')
    await inputs[2].setValue('20')

    await form.trigger('submit')
    await flushPromises()

    expect(wrapper.emitted('created')).toHaveLength(1)
    expect(stores.resources.createResource).toHaveBeenCalledWith(
      expect.objectContaining({
        venueId: 7,
        name: 'Cancha de prueba',
        type: 'Cancha',
        reservationMode: 'RESERVABLE',
        capacity: 20,
        isActive: true
      })
    )
  })

  it('permite cancelar sin crear un recurso', async () => {
    const wrapper = mount(ResourceCreatePanel)
    await wrapper.find('.form-actions button[type=button]').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
    expect(stores.resources.createResource).not.toHaveBeenCalled()
  })

  it('rechaza una capacidad inválida', async () => {
    const wrapper = mount(ResourceCreatePanel)
    const form = wrapper.find('form')

    const selects = form.findAll('select')
    const inputs = form.findAll('input')

    await selects[0].setValue('7')
    await inputs[0].setValue('Cancha de prueba')
    await inputs[1].setValue('Cancha')
    await inputs[2].setValue('0')

    await form.trigger('submit')
    await flushPromises()

    expect(stores.resources.createResource).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(
      'La capacidad debe ser un entero positivo.'
    )
  })
})
