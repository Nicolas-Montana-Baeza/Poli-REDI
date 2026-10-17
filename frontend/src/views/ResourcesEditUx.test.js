import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ResourcesView from './ResourcesView.vue'

const stores = vi.hoisted(() => ({ resources: null, venues: null, auth: null }))
vi.mock('@/stores/resources', () => ({ useResourcesStore: () => stores.resources }))
vi.mock('@/stores/venues', () => ({ useVenuesStore: () => stores.venues }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => stores.auth }))

const makeResource = () => ({
  id: 3, venueId: 7, name: 'Cancha Norte', type: 'Cancha',
  reservationMode: 'RESERVABLE', capacity: 20, imageUrl: '/images/old.png', isActive: true
})

const mountView = () => mount(ResourcesView, {
  global: {
    stubs: {
      VenuesAdminPanel: true, ResourceCreatePanel: true,
      SkeletonLoader: true, RouterLink: true
    }
  }
})

describe('ResourcesView - edición unificada UX', () => {
  beforeEach(() => {
    stores.resources = {
      resources: [makeResource()], loading: false,
      error: null, actionError: null, actionSuccess: null,
      updatingResourceId: null, updatingImageId: null,
      fetchResources: vi.fn().mockResolvedValue([]),
      updateOperational: vi.fn().mockImplementation(async (id, changes) => {
        stores.resources.resources[0] = { ...stores.resources.resources[0], ...changes }
      }),
      updateImage: vi.fn().mockImplementation(async (id, imageUrl) => {
        stores.resources.resources[0] = { ...stores.resources.resources[0], imageUrl }
      })
    }
    stores.venues = {
      venues: [{ id: 7, name: 'Sede Norte', isActive: true }],
      loading: false, error: null,
      fetchVenues: vi.fn().mockResolvedValue([])
    }
    stores.auth = { isAdmin: true, loadAuthUser: vi.fn() }
  })

  it('abre un solo editor para datos e imagen', async () => {
    const wrapper = mountView()
    expect(wrapper.find('.resource-editor').exists()).toBe(false)
    await wrapper.find('.edit-resource').trigger('click')
    expect(wrapper.findAll('.resource-editor').length).toBe(1)
    expect(wrapper.find('.resource-editor input[maxlength="500"]').exists()).toBe(true)
    expect(wrapper.find('.resource-editor').text()).toContain('Guardar cambios')
  })

  it('guarda datos e imagen por separado, desde un solo formulario', async () => {
    const wrapper = mountView()
    await wrapper.find('.edit-resource').trigger('click')
    await wrapper.find('.resource-editor input[maxlength="120"]').setValue('Cancha Norte Nueva')
    await wrapper.find('.resource-editor input[maxlength="500"]').setValue('/images/nueva.png')
    await wrapper.find('.resource-editor').trigger('submit')
    await flushPromises()

    expect(stores.resources.updateOperational).toHaveBeenCalledWith(3, expect.objectContaining({
      name: 'Cancha Norte Nueva', isActive: true, capacity: 20
    }))
    expect(stores.resources.updateImage).toHaveBeenCalledWith(3, '/images/nueva.png')
    expect(wrapper.find('.resource-editor').exists()).toBe(false)
    expect(stores.venues.fetchVenues).toHaveBeenCalled()
  })

  it('conserva el editor si falla la imagen y al reintentar no repite datos guardados', async () => {
    const wrapper = mountView()
    stores.resources.updateImage.mockRejectedValueOnce(new Error('error de imagen'))
    await wrapper.find('.edit-resource').trigger('click')
    await wrapper.find('.resource-editor input[maxlength="120"]').setValue('Cancha Nueva')
    await wrapper.find('.resource-editor input[maxlength="500"]').setValue('/images/nueva.png')
    await wrapper.find('.resource-editor').trigger('submit')
    await flushPromises()

    expect(wrapper.find('.resource-editor').exists()).toBe(true)
    expect(stores.resources.actionError).toContain('datos generales se guardaron')
    expect(stores.resources.updateOperational).toHaveBeenCalledTimes(1)

    await wrapper.find('.resource-editor').trigger('submit')
    await flushPromises()
    expect(stores.resources.updateOperational).toHaveBeenCalledTimes(1)
    expect(stores.resources.updateImage).toHaveBeenCalledTimes(2)
    expect(wrapper.find('.resource-editor').exists()).toBe(false)
  })

  it('no realiza solicitudes si no hay cambios', async () => {
    const wrapper = mountView()
    await wrapper.find('.edit-resource').trigger('click')
    await wrapper.find('.resource-editor').trigger('submit')
    await flushPromises()
    expect(stores.resources.updateOperational).not.toHaveBeenCalled()
    expect(stores.resources.updateImage).not.toHaveBeenCalled()
  })
})
