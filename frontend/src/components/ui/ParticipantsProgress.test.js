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
    expect(owner.text()).toContain('no puedes retirar tu propia participación')
  })

  it.each([
    ['CANCELLED', 'Cancelada', 'Esta reserva fue cancelada.'],
    ['EXPIRED', 'Vencida', 'El plazo para reunir participantes terminó y la reserva no fue confirmada.']
  ])('renderiza %s como estado terminal compacto', (status, badge, message) => {
    vi.setSystemTime(new Date('2026-07-20T19:00:00Z'))
    const wrapper = mount(ParticipantsProgress, { props: { progress: progress({ status }) } })
    expect(wrapper.text()).toContain('Estado de la invitación')
    expect(wrapper.text()).toContain(badge)
    expect(wrapper.text()).toContain(message)
    expect(wrapper.find('progress').exists()).toBe(false)
    expect(wrapper.find('.visual-progress').exists()).toBe(false)
    expect(wrapper.find('button').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('participantes confirmados')
    expect(wrapper.text()).not.toContain('La reserva se confirmará')
    expect(wrapper.text()).not.toContain('Plazo')
    expect(wrapper.attributes('aria-label')).toBe(`Estado de la invitación: ${badge}. ${message}`)
  })

  it('los estados activos conservan el copy aprobado, la barra y las acciones', () => {
    vi.setSystemTime(new Date('2026-07-20T17:00:00Z'))
    const wrapper = mount(ParticipantsProgress, { props: { progress: progress({ participantCount: 0, targetParticipants: 0, capacity: 0 }) } })
    expect(wrapper.text()).toContain('Participantes')
    expect(wrapper.text()).toContain('0 de 0 participantes confirmados')
    expect(wrapper.text()).toContain('Pendiente')
    expect(wrapper.attributes('aria-label')).toContain('Estado Pendiente')
    expect(wrapper.get('progress').attributes('max')).toBe('1')
    expect(wrapper.get('progress').attributes('value')).toBe('0')
    expect(wrapper.get('progress').attributes('aria-label')).toBe('0 participantes confirmados; mínimo 10; objetivo 0; capacidad 0')
    expect(wrapper.get('progress').classes()).toContain('sr-only')
    expect(wrapper.text()).toContain('La reserva se confirmará automáticamente al llegar a 10 participantes.')
    expect(wrapper.text()).toContain('Disponible hasta el')
    expect(wrapper.text()).toContain('Confirmar participación')
    expect(wrapper.get('.visual-fill').attributes('style')).toContain('width: 0%')
    expect(wrapper.findAll('.visual-marker')).toHaveLength(1)
    expect(wrapper.text()).toContain('Mínimo 10')
    expect(wrapper.text()).toContain('Capacidad 0')
    expect(wrapper.text()).not.toContain('Progreso de participantes')
    expect(wrapper.text()).not.toContain('Mínimo requerido')
    expect(wrapper.text()).not.toContain('La reserva se confirma al alcanzar el mínimo requerido.')
  })

  it('calcula barra por capacidad, cambia al alcanzar mínimo y posiciona marcadores', async () => {
    vi.setSystemTime(new Date('2026-07-20T17:00:00Z'))
    const wrapper = mount(ParticipantsProgress, { props: { progress: progress() } })
    expect(wrapper.get('.visual-progress').attributes('aria-hidden')).toBe('true')
    expect(wrapper.get('.visual-fill').attributes('style')).toContain('width: 45%')
    expect(wrapper.get('.visual-fill').classes()).not.toContain('reached')
    const markerStyles = wrapper.findAll('.visual-marker').map(item => item.attributes('style'))
    expect(wrapper.findAll('.visual-marker')[0].classes()).toContain('marker-minimum')
    expect(wrapper.findAll('.visual-marker')[1].classes()).toContain('marker-target')
    expect(markerStyles[0]).toContain('left: 50%')
    expect(markerStyles[1]).toContain('left: 60%')
    expect(wrapper.get('progress').attributes('max')).toBe('20')
    expect(wrapper.get('progress').attributes('value')).toBe('9')
    expect(wrapper.get('progress').attributes('aria-label')).toBe('9 participantes confirmados; mínimo 10; objetivo 12; capacidad 20')
    await wrapper.setProps({ progress: progress({ participantCount: 10 }) })
    expect(wrapper.get('.visual-fill').classes()).toContain('reached')
  })

  it('combina marcador cuando mínimo y objetivo coinciden', () => {
    vi.setSystemTime(new Date('2026-07-20T17:00:00Z'))
    const wrapper = mount(ParticipantsProgress, { props: { progress: progress({ participantCount: 1, minimumParticipants: 10, targetParticipants: 10 }) } })
    expect(wrapper.findAll('.visual-marker')).toHaveLength(1)
    expect(wrapper.get('.visual-marker').classes()).toContain('marker-combined')
    expect(wrapper.get('progress').attributes('aria-label')).toBe('1 participante confirmado; mínimo 10; objetivo 10; capacidad 20')
    expect(wrapper.text()).toContain('Mínimo 10')
    expect(wrapper.text()).not.toContain('Objetivo 10')
  })
})
