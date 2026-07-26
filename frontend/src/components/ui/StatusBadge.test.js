// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import StatusBadge from './StatusBadge.vue'

describe('StatusBadge', () => {
  it.each([
    ['PENDING', 'Pendiente'], ['CONFIRMED', 'Confirmada'], ['CANCELLED', 'Cancelada'],
    ['EXPIRED', 'Vencida'], ['REJECTED', 'Rechazada'], ['ACTIVE', 'Activo'], ['INACTIVE', 'Inactivo']
  ])('localiza %s', (status, text) => {
    expect(mount(StatusBadge, { props: { status } }).text()).toBe(text)
  })
  it('usa fallback neutral sin confundirlo con cancelado', () => {
    const unknown = mount(StatusBadge, { props: { status: 'CUSTOM' } })
    expect(unknown.text()).toBe('CUSTOM')
    expect(unknown.classes()).toContain('status-neutral')
    expect(unknown.classes()).not.toContain('status-cancelled')
    expect(mount(StatusBadge, { props: { status: '', label: 'Manual' } }).text()).toBe('Manual')
    expect(mount(StatusBadge).text()).toBe('Estado desconocido')
  })
})
