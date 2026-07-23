// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ParticipantsProgress from './ParticipantsProgress.vue'

const progress = overrides => ({
  participantCount: 9,
  minimumParticipants: 10,
  targetParticipants: 12,
  capacity: 20,
  status: 'PENDING',
  confirmationDeadline: '2026-07-20T14:00:00-04:00',
  isMember: false,
  isOwner: false,
  ...overrides
})

describe('ParticipantsProgress', () => {
  beforeEach(() => vi.useFakeTimers())
  it('permite confirmar cuando está activa y dentro del plazo', async () => {
    vi.setSystemTime(new Date('2026-07-20T17:00:00Z'))
    const wrapper = mount(ParticipantsProgress, { props: { progress: progress() } })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('permite retirar a miembro no owner y nunca al owner', () => {
    vi.setSystemTime(new Date('2026-07-20T17:00:00Z'))
    expect(mount(ParticipantsProgress, { props: { progress: progress({ isMember: true }) } }).text()).toContain('Retirar participación')
    const owner = mount(ParticipantsProgress, { props: { progress: progress({ isMember: true, isOwner: true }) } })
    expect(owner.find('button').exists()).toBe(false)
    expect(owner.text()).toContain('El solicitante no puede retirarse.')
  })

  it('oculta acciones canceladas o vencidas con mensaje correcto', () => {
    vi.setSystemTime(new Date('2026-07-20T19:00:00Z'))
    const expired = mount(ParticipantsProgress, { props: { progress: progress() } })
    expect(expired.find('button').exists()).toBe(false)
    expect(expired.text()).toContain('El plazo de confirmación ya venció.')
    const cancelled = mount(ParticipantsProgress, { props: { progress: progress({ status: 'CANCELLED' }) } })
    expect(cancelled.find('button').exists()).toBe(false)
    expect(cancelled.text()).toContain('La solicitud fue cancelada')
  })
})
