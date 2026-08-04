// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import ReservationDetailModal from './ReservationDetailModal.vue'

const getJoinCode = vi.fn()
vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    getJoinCode: (...args) => getJoinCode(...args),
    rotateJoinCode: vi.fn()
  }
}))

const base = {
  id: 1,
  title: 'Basquetbol',
  resourceName: 'Cancha 2, Centro Deportivo',
  startTime: '2099-07-28T17:00:00',
  durationMinutes: 60,
  status: 'PENDING'
}
const mountModal = (reservationProps, componentProps = {}) => mount(ReservationDetailModal, {
  props: {
    visible: true,
    reservation: { ...base, ...reservationProps },
    ...componentProps
  },
  global: { stubs: { Teleport: true } }
})

describe('ReservationDetailModal', () => {
  beforeEach(() => {
    getJoinCode.mockReset()
  })

  it('shows a workshop summary without irrelevant progress', () => {
    const wrapper = mountModal({ isWorkshop: true, title: 'Taller: Entrenamiento funcional' })

    expect(wrapper.text().match(/Entrenamiento funcional/g)).toHaveLength(1)
    expect(wrapper.findAll('[data-availability-type="workshop"]')).toHaveLength(1)
    expect(wrapper.get('[data-availability-type="workshop"]').text()).toBe('Taller')
    expect(wrapper.text()).not.toContain('Actividad institucional')
    expect(wrapper.text()).not.toContain('Taller programado')
    expect(wrapper.text()).not.toContain('Taller: Entrenamiento funcional')
    expect(wrapper.text()).toContain('Este horario está reservado para el taller y no admite reservas particulares.')
    expect(wrapper.text()).toContain('Entendido')
    expect(wrapper.text()).not.toContain('Progreso')
    expect(wrapper.find('table').exists()).toBe(false)
    expect(wrapper.findAll('.fact')).toHaveLength(3)
    expect(wrapper.findAll('.fact svg')).toHaveLength(3)
    expect(wrapper.findAll('.fact .icon')).toHaveLength(3)
    expect(wrapper.get('.institutional-resource strong').text()).toBe('Cancha 2, Centro Deportivo')
    expect(wrapper.find('.institutional-resource .note-icon').exists()).toBe(true)
    expect(wrapper.find('.info-callout .callout-icon').exists()).toBe(true)
    expect(wrapper.find('.detail-body').exists()).toBe(true)
    expect(wrapper.find('.detail-actions').exists()).toBe(true)
  })

  it('ofrece desinscripcion solo con capacidad explicita y mantiene el detalle abierto', async () => {
    const wrapper = mountModal({
      isWorkshop: true,
      title: 'Taller: Entrenamiento funcional'
    }, {
      canWithdrawWorkshop: true,
      workshopActionMessage: 'Te desinscribiste del taller. Tu cupo quedó disponible.'
    })

    expect(wrapper.text()).toContain('Desinscribirme')
    expect(wrapper.get('.workshop-action-feedback').attributes('role'))
      .toBe('status')
    await wrapper.get('.detail-actions .danger').trigger('click')
    expect(wrapper.emitted('withdraw-workshop')).toHaveLength(1)
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    await wrapper.setProps({ workshopWithdrawing: true })
    expect(wrapper.text()).toContain('Desinscribiendo…')
    expect(wrapper.get('.detail-actions .danger').attributes('disabled'))
      .toBeDefined()
  })

  it('ofrece inscripcion al taller y bloquea el doble envio mientras procesa', async () => {
    const wrapper = mountModal({
      isWorkshop: true,
      title: 'Taller: Entrenamiento funcional'
    }, {
      canEnrollWorkshop: true
    })

    const button = wrapper.get('.detail-actions .primary')
    expect(button.text()).toBe('Inscribirme')
    await button.trigger('click')
    expect(wrapper.emitted('enroll-workshop')).toHaveLength(1)
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    await wrapper.setProps({ workshopEnrolling: true })
    expect(wrapper.text()).toContain('Inscribiendo…')
    expect(wrapper.get('.detail-actions .primary').attributes('disabled'))
      .toBeDefined()
    await wrapper.get('.detail-actions .primary').trigger('click')
    expect(wrapper.emitted('enroll-workshop')).toHaveLength(1)
  })

  it.each([
    ['sin capacidad explicita', {}],
    ['historial', { canEnrollWorkshop: true, readOnly: true }],
    ['actividad no taller', { canEnrollWorkshop: true, reservation: { ...base, isScheduledActivity: true, activityType: 'CLASS' } }]
  ])('oculta inscripcion en %s', (_label, componentProps) => {
    const reservation = componentProps.reservation || {
      ...base,
      isWorkshop: true,
      title: 'Taller: Entrenamiento funcional'
    }
    const { reservation: _ignored, ...props } = componentProps
    const wrapper = mountModal(reservation, props)

    expect(wrapper.text()).not.toContain('Inscribirme')
    expect(wrapper.emitted('enroll-workshop')).toBeUndefined()
  })

  it.each([
    ['sin inscripcion', {}],
    ['historial', { canWithdrawWorkshop: true, readOnly: true }],
    ['actividad no taller', { canWithdrawWorkshop: true, reservation: { ...base, isScheduledActivity: true, activityType: 'CLASS' } }]
  ])('oculta desinscripcion en %s', (_label, componentProps) => {
    const reservation = componentProps.reservation || {
      ...base,
      isWorkshop: true,
      title: 'Taller: Entrenamiento funcional'
    }
    const { reservation: _ignored, ...props } = componentProps
    const wrapper = mountModal(reservation, props)

    expect(wrapper.text()).not.toContain('Desinscribirme')
    expect(wrapper.emitted('withdraw-workshop')).toBeUndefined()
  })

  it('mounts the real location, date, time and duration Lucide SVGs', () => {
    const wrapper = mountModal({ isWorkshop: true, title: 'Taller: Entrenamiento funcional' })
    const selectors = [
      '.institutional-resource svg.lucide-map-pin',
      '.facts svg.lucide-calendar-days',
      '.facts svg.lucide-clock',
      '.facts svg.lucide-timer'
    ]
    const icons = selectors.map(selector => wrapper.get(selector))

    expect(icons).toHaveLength(4)
    expect(new Set(icons.map(icon => icon.element.innerHTML)).size).toBe(4)
    icons.forEach(icon => {
      expect(icon.attributes('aria-hidden')).toBe('true')
      expect(icon.attributes('width')).toBe('20')
      expect(icon.attributes('height')).toBe('20')
      expect(icon.findAll('path, rect, circle, line, polyline').length).toBeGreaterThan(0)
    })
  })

  it.each([
    ['CLASS', 'Clase: Yoga', 'Clase', 'Yoga'],
    ['TRAINING', 'Entrenamiento - Selección', 'Entrenamiento', 'Selección'],
    ['EVENT', 'Evento: Bienvenida', 'Evento', 'Bienvenida']
  ])('uses one type chip and a clean name for scheduled %s', (activityType, activityTitle, chip, name) => {
    const wrapper = mountModal({
      availabilityKind: 'SCHEDULED_ACTIVITY',
      isScheduledActivity: true,
      activityType,
      title: activityTitle
    })

    expect(wrapper.get('[data-availability-type]').text()).toBe(chip)
    expect(wrapper.text().match(new RegExp(name, 'g'))).toHaveLength(1)
    expect(wrapper.text()).not.toContain(activityTitle)
    expect(wrapper.find('.summary-heading').exists()).toBe(false)
  })

  it('shows group progress, bounded stepper and permission information', async () => {
    const wrapper = mountModal({
      participantCount: 4,
      minimumParticipants: 10,
      targetParticipants: 14,
      capacity: 22,
      canEditTarget: true,
      confirmationDeadline: '2099-07-28T18:15:00'
    }, { canEditTarget: true })
    await wrapper.setProps({ canCancel: false })

    expect(wrapper.text()).toContain('4 de 14 confirmados')
    expect(wrapper.text()).toContain('Mínimo 10')
    expect(wrapper.text()).toContain('Capacidad 22')
    expect(wrapper.text()).toContain('Solo quien creó la reserva o un administrador puede cancelarla.')
    expect(wrapper.find('.stepper').exists()).toBe(true)
    expect(wrapper.findAll('.stepper button')).toHaveLength(2)
    expect(wrapper.findAll('.stepper .stepper-icon')).toHaveLength(2)
    expect(wrapper.find('.permission-note .callout-icon').exists()).toBe(true)

    await wrapper.get('[aria-label="Aumentar objetivo"]').trigger('click')
    await wrapper.get('.save-target').trigger('click')
    expect(wrapper.emitted('update-target')[0][0]).toBe(15)
  })

  it('hides target editing and never emits without an explicit capability', async () => {
    const wrapper = mountModal({
      participantCount: 4,
      minimumParticipants: 10,
      targetParticipants: 14,
      capacity: 22,
      canEditTarget: true,
      confirmationDeadline: '2099-07-28T18:15:00'
    })

    expect(wrapper.text()).toContain('4 de 14 confirmados')
    expect(wrapper.text()).toContain('Mínimo 10')
    expect(wrapper.text()).toContain('Capacidad 22')
    expect(wrapper.text()).not.toContain('Objetivo de participantes')
    expect(wrapper.find('.target-editor').exists()).toBe(false)
    expect(wrapper.emitted('update-target')).toBeUndefined()
  })

  it.each([
    ['read-only', { readOnly: true }],
    ['invited participant', { participationMode: true }]
  ])('hides target editing in %s mode even if capability is passed', (_label, modeProps) => {
    const wrapper = mountModal({
      participantCount: 4,
      minimumParticipants: 10,
      targetParticipants: 14,
      capacity: 22,
      confirmationDeadline: '2099-07-28T18:15:00'
    }, { canEditTarget: true, ...modeProps })

    expect(wrapper.find('.target-editor').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Guardar cambios')
    expect(wrapper.emitted('update-target')).toBeUndefined()
  })

  it('uses bounded Lucide icons and contains no manually-authored SVG markup', () => {
    const source = readFileSync(join(process.cwd(), 'src/components/availability/ReservationDetailModal.vue'), 'utf8')

    expect(source).not.toMatch(/<svg[\s>]/i)
    expect(source).toContain('width: 20px !important')
    expect(source).toContain('height: 20px !important')
    expect(source).toContain('max-width: 20px')
    expect(source).toContain('max-height: 20px')
    expect(source).toContain('flex: 0 0 20px')
    expect(source).toContain('display: block')
    expect(source).toContain('flex-shrink: 0')
    expect(source).toContain('.fact > svg.fact-icon')
    expect(source).not.toMatch(/\.fact \.icon,\s*\.info-callout\s*,/)
  })

  it('consults the invitation code only on demand for an authorized owner', async () => {
    getJoinCode.mockResolvedValue({ joinCode: 'ABC-123' })
    const wrapper = mountModal({
      targetParticipants: 10,
      participantCount: 1,
      minimumParticipants: 10,
      capacity: 22
    })
    await wrapper.setProps({ canManageJoinCode: true })

    expect(getJoinCode).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Consultar código')
    await wrapper.get('.join-code-heading button').trigger('click')

    expect(getJoinCode).toHaveBeenCalledWith(1)
    await vi.waitFor(() => expect(wrapper.text()).toContain('ABC-123'))
    expect(wrapper.text()).toContain('Copiar código')
    expect(wrapper.text()).toContain('Compartir invitación')
  })

  it('does not expose invitation controls when authorization or status disallows them', async () => {
    const wrapper = mountModal({ targetParticipants: 10 })

    expect(wrapper.find('.join-code').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Consultar código')
    expect(getJoinCode).not.toHaveBeenCalled()
  })

  it('shows a recoverable error when invitation lookup fails', async () => {
    getJoinCode.mockRejectedValue(new Error('network'))
    const wrapper = mountModal({ targetParticipants: 10 })
    await wrapper.setProps({ canManageJoinCode: true })
    await wrapper.get('.join-code-heading button').trigger('click')

    await vi.waitFor(() => {
      expect(wrapper.get('[role="alert"]').text()).toContain('No se pudo recuperar el código.')
    })
  })

  it('reuses the detail safely for an invited participant', async () => {
    const wrapper = mountModal({
      participantCount: 1,
      minimumParticipants: 10,
      targetParticipants: 10,
      capacity: 22,
      confirmationDeadline: '2099-07-28T18:15:00',
      isMember: false,
      isOwner: false
    })
    await wrapper.setProps({ participationMode: true, canCancel: false })

    expect(wrapper.text()).toContain('Detalle de la invitación')
    expect(wrapper.text()).toContain('1 de 10 confirmados')
    expect(wrapper.text()).toContain('Confirmar participación')
    expect(wrapper.find('.summary-card').exists()).toBe(false)
    expect(wrapper.find('.join-code').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Instalación')
    expect(wrapper.text()).not.toContain('Cancelar reserva')
    expect(wrapper.text()).not.toContain('Objetivo de participantes')

    await wrapper.get('.participation-actions button').trigger('click')
    expect(wrapper.emitted('confirm-participation')).toHaveLength(1)
  })

  it('closes with Escape and exposes a modal dialog', async () => {
    const wrapper = mountModal({})
    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('tabindex')).toBe('-1')
    await dialog.trigger('keydown', { key: 'Escape' })
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('offers withdrawal only to a non-owner member in participation mode', async () => {
    const wrapper = mountModal({
      participantCount: 2,
      minimumParticipants: 10,
      targetParticipants: 10,
      capacity: 22,
      confirmationDeadline: '2099-07-28T18:15:00',
      isMember: true,
      isOwner: false
    })
    await wrapper.setProps({ participationMode: true })

    expect(wrapper.text()).toContain('Retirar participación')
    expect(wrapper.text()).not.toContain('Confirmar participación')
    await wrapper.get('.participation-actions button').trigger('click')
    expect(wrapper.emitted('withdraw-participation')).toHaveLength(1)
  })
})
