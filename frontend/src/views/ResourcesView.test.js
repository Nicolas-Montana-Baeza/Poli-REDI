import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ResourcesView from './ResourcesView.vue'

const stores = vi.hoisted(() => ({
  resources: null,
  venues: null,
  auth: null
}))

vi.mock('@/stores/resources', () => ({
  useResourcesStore: () => stores.resources
}))

vi.mock('@/stores/venues', () => ({
  useVenuesStore: () => stores.venues
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => stores.auth
}))

const resource = (id, venueId, name, type) => ({
  id,
  venueId,
  name,
  type,
  reservationMode: 'RESERVABLE',
  capacity: 20,
  imageUrl: '',
  isActive: true,
  status: 'available'
})

const mountView = async () => {
  const wrapper = mount(ResourcesView, {
    global: {
      stubs: {
        VenuesAdminPanel: true,
        ResourceCreatePanel: true,
        SkeletonLoader: true,
        RouterLink: true
      }
    }
  })

  await flushPromises()
  return wrapper
}

const visibleResources = (wrapper) =>
  wrapper.findAll('article.resource-card')
    .map(card => card.text())
    .join(' | ')

describe('ResourcesView - filtros administrativos', () => {
  beforeEach(() => {
    stores.resources = {
      resources: [
        resource(1, 7, 'Cancha Norte', 'Cancha'),
        resource(2, 8, 'Cancha Sur', 'Cancha'),
        resource(3, 7, 'Sala Norte', 'Sala')
      ],
      loading: false,
      error: null,
      actionError: null,
      actionSuccess: null,
      fetchResources: vi.fn().mockResolvedValue([])
    }

    stores.venues = {
      venues: [
        { id: 7, name: 'Sede Norte', isActive: true },
        { id: 8, name: 'Sede Sur', isActive: true }
      ],
      loading: false,
      error: null
    }

    stores.auth = {
      isAdmin: true,
      loadAuthUser: vi.fn()
    }
  })

  it('filtra los recursos por sede', async () => {
    const wrapper = await mountView()

    const venueSelect = wrapper.find('.filters select')
    await venueSelect.setValue('7')

    const visible = visibleResources(wrapper)

    expect(visible).toContain('Cancha Norte')
    expect(visible).toContain('Sala Norte')
    expect(visible).not.toContain('Cancha Sur')
  })

  it('combina filtros de sede y tipo', async () => {
    const wrapper = await mountView()

    const selects = wrapper.findAll('.filters select')

    await selects[0].setValue('7')
    await selects[1].setValue('Sala')

    const visible = visibleResources(wrapper)

    expect(visible).toContain('Sala Norte')
    expect(visible).not.toContain('Cancha Norte')
    expect(visible).not.toContain('Cancha Sur')
  })

  it('restaura todas las sedes', async () => {
    const wrapper = await mountView()

    const venueSelect = wrapper.find('.filters select')

    await venueSelect.setValue('7')
    await venueSelect.setValue('ALL')

    const visible = visibleResources(wrapper)

    expect(visible).toContain('Cancha Norte')
    expect(visible).toContain('Cancha Sur')
    expect(visible).toContain('Sala Norte')
  })
})
