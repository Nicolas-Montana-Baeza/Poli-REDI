// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const state = vi.hoisted(() => ({
  auth: null,
  resources: null,
  reservations: null
}))

vi.mock('@/stores/auth', async () => {
  const { reactive } = await import('vue')
  state.auth = reactive({
    user: { fullName: 'Nicolás Pérez' },
    account: null,
    loadAuthUser: vi.fn()
  })
  return { useAuthStore: () => state.auth }
})

vi.mock('@/stores/resources', async () => {
  const { reactive } = await import('vue')
  state.resources = reactive({
    resources: [],
    loading: false,
    error: null,
    fetchResources: vi.fn()
  })
  return { useResourcesStore: () => state.resources }
})

vi.mock('@/stores/reservations', async () => {
  const { reactive } = await import('vue')
  state.reservations = reactive({
    myReservations: [],
    myLoading: false,
    myLoadingError: null,
    fetchMyReservations: vi.fn()
  })
  return { useReservationsStore: () => state.reservations }
})

import DashboardView from './DashboardView.vue'

const RouterLinkStub = {
  props: ['to'],
  template: '<a :data-to="typeof to === \'string\' ? to : JSON.stringify(to)"><slot /></a>'
}

const mountDashboard = () =>
  mount(DashboardView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        SkeletonLoader: { template: '<div class="skeleton-stub" />' }
      }
    }
  })

const resource = (id, name, type) => ({
  id,
  name,
  type,
  status: 'available',
  imageUrl: ''
})

const reservation = (id, title, startTime) => ({
  id,
  title,
  resourceName: 'Cancha 2, Centro Deportivo',
  status: 'PENDING',
  startTime,
  durationMinutes: 60
})

describe('DashboardView', () => {
  beforeEach(() => {
    state.auth.user = { fullName: 'Nicolás Pérez' }
    state.resources.resources.splice(
      0,
      state.resources.resources.length,
      resource(1, 'Piscina', 'Piscina'),
      resource(2, 'Sala Spinning', 'Sala'),
      resource(3, 'Muro Escalada', 'Muro')
    )
    state.resources.loading = false
    state.resources.error = null
    state.reservations.myReservations.splice(
      0,
      state.reservations.myReservations.length,
      reservation(8, 'Basquetbol', '2099-07-29T19:15:00-03:00'),
      reservation(7, 'Fútbol', '2099-07-28T18:00:00-03:00')
    )
    state.reservations.myLoading = false
    state.reservations.myLoadingError = null
    vi.clearAllMocks()
  })

  it('muestra las acciones del hero y una sola próxima reserva', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    expect(wrapper.text()).toContain('Hola, Nicolás')
    expect(wrapper.find('[data-to="/availability"]').text())
      .toContain('Reservar instalación')
    expect(wrapper.find('[data-to="/join"]').text())
      .toContain('Ingresar código')
    expect(wrapper.findAll('#next-reservation-title')).toHaveLength(1)
    expect(wrapper.text()).toContain('Fútbol')
    expect(wrapper.text()).not.toContain('Basquetbol')
    expect(wrapper.find('[data-to="/reservations/7"]').text())
      .toContain('Ver detalle')
    expect(wrapper.find('[data-to="/history"]').text())
      .toContain('Ver todas mis reservas')
    expect(wrapper.text()).not.toContain('Mis Reservas')
    expect(wrapper.text()).not.toContain('Organiza tu próxima actividad')
  })

  it('filtra instalaciones por búsqueda y por los tipos reales', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    const search = wrapper.get('input[placeholder="Buscar instalación"]')
    await search.setValue('pisc')
    expect(wrapper.text()).toContain('Piscina')
    expect(wrapper.text()).not.toContain('Sala Spinning')

    await search.setValue('')
    const select = wrapper.get('select[aria-label="Filtrar por tipo"]')
    expect(select.findAll('option').map((option) => option.text()))
      .toEqual(['Todas', 'Muro', 'Piscina', 'Sala'])
    await select.setValue('Sala')
    expect(wrapper.text()).toContain('Sala Spinning')
    expect(wrapper.text()).not.toContain('Muro Escalada')
  })

  it('mantiene estados de carga, error y ausencia de resultados', async () => {
    state.resources.loading = true
    state.reservations.myLoading = true
    const loadingWrapper = mountDashboard()
    await flushPromises()
    expect(loadingWrapper.findAll('.skeleton-stub')).toHaveLength(2)
    loadingWrapper.unmount()

    state.resources.loading = false
    state.reservations.myLoading = false
    state.resources.error = 'No se pudieron cargar los recursos'
    state.reservations.myLoadingError = 'No se pudieron cargar tus reservas.'
    const errorWrapper = mountDashboard()
    await flushPromises()
    expect(errorWrapper.text()).toContain('No se pudieron cargar los recursos')
    expect(errorWrapper.text()).toContain('No se pudieron cargar tus reservas.')
    errorWrapper.unmount()

    state.resources.error = null
    state.reservations.myLoadingError = null
    state.reservations.myReservations.splice(
      0,
      state.reservations.myReservations.length
    )
    const emptyWrapper = mountDashboard()
    await flushPromises()
    expect(emptyWrapper.text()).toContain('Aún no tienes reservas próximas')
    await emptyWrapper
      .get('input[placeholder="Buscar instalación"]')
      .setValue('inexistente')
    expect(emptyWrapper.text()).toContain('No encontramos instalaciones')
  })

  it('mantiene el contrato responsive de instalaciones sin una lista vertical móvil', () => {
    const carouselSource = readFileSync(
      resolve(process.cwd(), 'src/components/dashboard/FacilityCarousel.vue'),
      'utf8'
    )
    const cardSource = readFileSync(
      resolve(process.cwd(), 'src/components/dashboard/FacilityCard.vue'),
      'utf8'
    )

    expect(carouselSource).toContain('@media (max-width: 600px)')
    expect(carouselSource).toContain('overflow-x: auto')
    expect(carouselSource).toContain('scroll-snap-type: x mandatory')
    expect(carouselSource).toContain('flex: 0 0 min(82vw, 280px)')
    expect(carouselSource).toContain('repeat(2, minmax(0, 1fr))')
    expect(carouselSource).toContain('repeat(4, minmax(0, 1fr))')
    expect(cardSource).toContain('height: min(42vw, 150px)')
    expect(cardSource).toContain('aspect-ratio: 16 / 9')
  })
})
