import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ResourcesView from './ResourcesView.vue'

const stores = vi.hoisted(() => ({ resources: null, venues: null, auth: null }))

vi.mock('@/stores/resources', () => ({ useResourcesStore: () => stores.resources }))
vi.mock('@/stores/venues', () => ({ useVenuesStore: () => stores.venues }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => stores.auth }))

const mountView = () => mount(ResourcesView, {
  global: {
    stubs: {
      VenuesAdminPanel: true,
      ResourceCreatePanel: true,
      SkeletonLoader: true,
      RouterLink: true
    }
  }
})

describe('ResourcesView - navegación UX administrativa', () => {
  beforeEach(() => {
    stores.resources = {
      resources: [{
        id: 1, venueId: 1, name: 'Cancha 1', type: 'Cancha',
        reservationMode: 'RESERVABLE', capacity: 22,
        imageUrl: '', isActive: true
      }],
      loading: false, error: null,
      actionError: null, actionSuccess: null,
      fetchResources: vi.fn().mockResolvedValue([])
    }
    stores.venues = {
      venues: [{ id: 1, name: 'Sede Central', isActive: true }],
      loading: false, error: null
    }
    stores.auth = { isAdmin: true, loadAuthUser: vi.fn() }
  })

  it('mantiene el listado como vista inicial y separa las sedes', async () => {
    const wrapper = mountView()
    expect(wrapper.find('.filters').exists()).toBe(true)
    expect(wrapper.find('.venues-section').element.style.display).toBe('none')

    await wrapper.findAll('.admin-tabs button')[1].trigger('click')
    expect(wrapper.find('.venues-section').element.style.display).not.toBe('none')
    expect(wrapper.find('.filters').exists()).toBe(false)

    await wrapper.findAll('.admin-tabs button')[0].trigger('click')
    expect(wrapper.find('.filters').exists()).toBe(true)
  })

  it('muestra el formulario de alta solo cuando se solicita', async () => {
    const wrapper = mountView()
    expect(wrapper.find('.create-section').exists()).toBe(false)
    await wrapper.find('.page-header > button').trigger('click')
    expect(wrapper.find('.create-section').exists()).toBe(true)
    await wrapper.find('.page-header > button').trigger('click')
    expect(wrapper.find('.create-section').exists()).toBe(false)
  })

  it('mantiene la consulta sin acciones administrativas para no administradores', () => {
    stores.auth.isAdmin = false
    const wrapper = mountView()
    expect(wrapper.find('.admin-tabs').exists()).toBe(false)
    expect(wrapper.find('.page-header > button').exists()).toBe(false)
    expect(wrapper.findAll('.resource-card').length).toBe(1)
  })
})
