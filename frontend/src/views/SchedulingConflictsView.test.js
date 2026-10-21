import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import SchedulingConflictsView from './SchedulingConflictsView.vue'

const mocks = vi.hoisted(() => ({ store: null }))

vi.mock('@/stores/schedulingConflicts', () => ({
  useSchedulingConflictsStore: () => mocks.store
}))

const sampleConflict = () => ({
  id: 71,
  resourceId: 1,
  resourceName: 'Cancha 1',
  status: 'PENDING',
  createdAt: '2026-10-21T12:00:00Z',
  items: [{
    id: 700,
    entityType: 'RESERVATION',
    title: 'Reserva particular',
    resolution: 'PENDING',
    occurrenceStart: '2026-10-21T14:00:00Z',
    occurrenceEnd: '2026-10-21T15:00:00Z'
  }]
})

let wrapper

beforeEach(() => {
  mocks.store = reactive({
    conflicts: [sampleConflict()],
    selectedConflict: null,
    statusFilter: 'PENDING',
    loading: false,
    loadingDetail: false,
    resolvingItemId: null,
    error: null,
    actionError: null,
    actionSuccess: null,

    fetchConflicts: vi.fn().mockResolvedValue(),

    fetchConflict: vi.fn(async () => {
      mocks.store.selectedConflict = sampleConflict()
      return mocks.store.selectedConflict
    }),

    resolveItem: vi.fn().mockResolvedValue(),

    clearMessages: vi.fn(() => {
      mocks.store.actionError = null
      mocks.store.actionSuccess = null
    })
  })
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ''
})

const openConflict = async () => {
  wrapper = mount(SchedulingConflictsView, {
    attachTo: document.body
  })

  await flushPromises()

  await wrapper
    .get('.conflict-card .primary-button')
    .trigger('click')

  await flushPromises()
}

const prepareCancellation = async () => {
  await wrapper
    .get('.item-card textarea')
    .setValue('Resolución administrativa justificada')

  await wrapper
    .get('.decision-button.cancel')
    .trigger('click')

  await nextTick()
}

describe('ADMIN-005: resoluciones administrativas', () => {
  it('consulta los conflictos pendientes', async () => {
    wrapper = mount(SchedulingConflictsView)
    await flushPromises()

    expect(mocks.store.fetchConflicts)
      .toHaveBeenCalledWith('PENDING')

    expect(wrapper.text())
      .toContain('Cancha 1')

    expect(mocks.store.resolveItem)
      .not.toHaveBeenCalled()
  })

  it('rechaza una decisión sin justificación', async () => {
    await openConflict()

    await wrapper
      .get('.decision-button.cancel')
      .trigger('click')

    expect(wrapper.text())
      .toContain('Debes ingresar una nota administrativa')

    expect(document.querySelector('.confirmation-modal'))
      .toBeNull()

    expect(mocks.store.resolveItem)
      .not.toHaveBeenCalled()
  })

  it('permite abandonar la confirmación sin cancelar', async () => {
    await openConflict()
    await prepareCancellation()

    expect(document.querySelector('.confirmation-modal'))
      .not.toBeNull()

    expect(mocks.store.resolveItem)
      .not.toHaveBeenCalled()

    document
      .querySelector('.confirmation-actions .secondary-button')
      .click()

    await nextTick()

    expect(document.querySelector('.confirmation-modal'))
      .toBeNull()

    expect(mocks.store.resolveItem)
      .not.toHaveBeenCalled()
  })

  it('envía CANCEL solo después de confirmar expresamente', async () => {
    await openConflict()
    await prepareCancellation()

    expect(mocks.store.resolveItem)
      .not.toHaveBeenCalled()

    document
      .querySelector('.confirmation-actions .confirm-button')
      .click()

    await flushPromises()

    expect(mocks.store.resolveItem).toHaveBeenCalledWith(
      71,
      700,
      {
        resolution: 'CANCEL',
        resolutionNote: 'Resolución administrativa justificada'
      }
    )

    expect(document.querySelector('.confirmation-modal'))
      .toBeNull()
  })

  it('conserva la confirmación abierta si la API falla', async () => {
    mocks.store.resolveItem = vi.fn(async () => {
      mocks.store.actionError = 'No se pudo resolver el conflicto.'
      throw new Error('API rechazada')
    })

    await openConflict()
    await prepareCancellation()

    document
      .querySelector('.confirmation-actions .confirm-button')
      .click()

    await flushPromises()

    expect(document.querySelector('.confirmation-modal'))
      .not.toBeNull()

    expect(wrapper.text())
      .toContain('No se pudo resolver el conflicto.')
  })
})
