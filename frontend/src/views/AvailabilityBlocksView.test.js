import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import AvailabilityBlocksView from './AvailabilityBlocksView.vue'
import {
  santiagoLocalToISOString,
  shiftDateKey
} from '../utils/availabilityBlockTime'

const stores = vi.hoisted(() => ({ resources: null, blocks: null }))

vi.mock('@/stores/resources', () => ({
  useResourcesStore: () => stores.resources
}))
vi.mock('@/stores/availabilityBlocks', () => ({
  useAvailabilityBlocksStore: () => stores.blocks
}))

const sampleBlock = () => ({
  id: 7,
  resourceId: 1,
  resourceName: 'Cancha 1',
  blockType: 'MAINTENANCE',
  reason: 'Mantención preventiva',
  startTime: '2026-10-22T13:00:00Z',
  endTime: '2026-10-22T14:00:00Z',
  isActive: true
})

const mountView = async () => {
  const wrapper = mount(AvailabilityBlocksView)
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  stores.resources = {
    resources: [{ id: 1, name: 'Cancha 1', isActive: true }],
    loading: false,
    error: null,
    fetchResources: vi.fn().mockResolvedValue()
  }
  stores.blocks = reactive({
    blocks: [sampleBlock()],
    loading: false, saving: false, updatingId: null,
    error: null, actionError: null, actionSuccess: null,
    fetchBlocks: vi.fn().mockResolvedValue(),
    createBlock: vi.fn().mockResolvedValue(sampleBlock()),
    deactivateBlock: vi.fn().mockImplementation(async (id) => {
      stores.blocks.blocks = stores.blocks.blocks.map(block =>
        block.id === id ? { ...block, isActive: false } : block
      )
    })
  })
})

describe('AvailabilityBlocksView — administración', () => {
  it('consulta inicialmente el rango y muestra bloqueos activos', async () => {
    const wrapper = await mountView()
    expect(stores.resources.fetchResources).toHaveBeenCalledOnce()
    expect(stores.blocks.fetchBlocks).toHaveBeenCalledOnce()
    expect(wrapper.findAll('.block-row')).toHaveLength(1)
    expect(wrapper.text()).toContain('Mantención preventiva')
    expect(wrapper.find('.create-panel').exists()).toBe(false)
  })

  it('crea un bloqueo enviando fechas RFC3339 desde la hora de Santiago', async () => {
    const wrapper = await mountView()
    await wrapper.find('.page-header button').trigger('click')
    expect(wrapper.find('.create-panel').exists()).toBe(true)
    await wrapper.find('.block-form select').setValue('1')
    await wrapper.findAll('.block-form input')[0].setValue('2026-10-22T10:00')
    await wrapper.findAll('.block-form input')[1].setValue('2026-10-22T11:00')
    await wrapper.find('.block-form').trigger('submit')
    await flushPromises()
    expect(stores.blocks.createBlock).toHaveBeenCalledWith({
      resourceId: 1,
      blockType: 'MAINTENANCE',
      reason: '',
      startTime: '2026-10-22T13:00:00.000Z',
      endTime: '2026-10-22T14:00:00.000Z'
    })
    expect(wrapper.find('.create-panel').exists()).toBe(false)
  })

  it('mantiene el formulario abierto si la API rechaza un solapamiento', async () => {
    stores.blocks.createBlock = vi.fn().mockRejectedValue(
      new Error('la instalación ya está bloqueada')
    )
    const wrapper = await mountView()
    await wrapper.find('.page-header button').trigger('click')
    await wrapper.find('.block-form select').setValue('1')
    await wrapper.findAll('.block-form input')[0].setValue('2026-10-22T10:00')
    await wrapper.findAll('.block-form input')[1].setValue('2026-10-22T11:00')
    await wrapper.find('.block-form').trigger('submit')
    await flushPromises()
    expect(wrapper.find('.create-panel').exists()).toBe(true)
    expect(wrapper.find('.form-error').text()).toContain('bloqueada')
  })

  it('exige confirmación antes de desactivar y actualiza el estado', async () => {
    const wrapper = await mountView()
    await wrapper.find('.deactivate-trigger').trigger('click')
    expect(stores.blocks.deactivateBlock).not.toHaveBeenCalled()
    await wrapper.find('.confirm-deactivate').trigger('click')
    await flushPromises()
    expect(stores.blocks.deactivateBlock).toHaveBeenCalledWith(7)
    expect(wrapper.find('.block-row').exists()).toBe(false)
    await wrapper.findAll('.filters select')[1].setValue('INACTIVE')
    expect(wrapper.find('.block-row').text()).toContain('Inactivo')
  })

  it('rechaza rangos superiores a 31 días sin hacer solicitudes', async () => {
    const wrapper = await mountView()
    await wrapper.findAll('.filters input')[0].setValue('2026-10-01')
    await wrapper.findAll('.filters input')[1].setValue('2026-11-15')
    await wrapper.find('.filters').trigger('submit')
    await flushPromises()
    expect(stores.blocks.fetchBlocks).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('Selecciona un rango válido de 1 a 31 días.')
  })
})

describe('Horarios de bloqueo en Santiago', () => {
  it('aplica desplazamientos UTC de verano e invierno', () => {
    expect(santiagoLocalToISOString('2026-10-21T10:00'))
      .toBe('2026-10-21T13:00:00.000Z')
    expect(santiagoLocalToISOString('2026-06-21T10:00'))
      .toBe('2026-06-21T14:00:00.000Z')
  })

  it('rechaza horas inexistentes o ambiguas durante el cambio DST', () => {
    expect(() => santiagoLocalToISOString('2026-09-06T00:30'))
      .toThrow(/no existe/)
    expect(() => santiagoLocalToISOString('2026-04-04T23:30'))
      .toThrow(/ambigua/)
  })

  it('calcula el rango de fechas sin depender del huso del equipo', () => {
    expect(shiftDateKey('2026-10-20', 15)).toBe('2026-11-04')
  })
})
