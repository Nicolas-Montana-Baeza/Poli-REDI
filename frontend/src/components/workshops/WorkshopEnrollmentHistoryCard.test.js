// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import WorkshopEnrollmentHistoryCard from './WorkshopEnrollmentHistoryCard.vue'

const enrollment = {
  id: 8,
  workshopId: 3,
  title: 'Escalada',
  description: 'Nivel inicial',
  location: 'Gimnasio',
  instructorName: 'Ana Pérez',
  dayText: 'Martes',
  scheduleText: '18:00 - 19:00',
  status: 'CONFIRMED',
  isActive: true,
  enrolledAt: '2026-07-20T15:00:00Z'
}

describe('WorkshopEnrollmentHistoryCard', () => {
  it('muestra el taller y la fecha de inscripción con etiqueta semántica', () => {
    const wrapper = mount(WorkshopEnrollmentHistoryCard, {
      props: { enrollment }
    })

    expect(wrapper.text()).toContain('Taller')
    expect(wrapper.text()).toContain('Inscrito')
    expect(wrapper.text()).toContain('Escalada')
    expect(wrapper.text()).toContain('Gimnasio')
    expect(wrapper.text()).toContain('Martes · 18:00 - 19:00')
    expect(wrapper.text()).toContain('Ana Pérez')
    expect(wrapper.text()).toContain('Inscripción:')
  })

  it('prioriza la cancelación y distingue un taller no vigente', async () => {
    const wrapper = mount(WorkshopEnrollmentHistoryCard, {
      props: {
        enrollment: { ...enrollment, status: 'CANCELLED', isActive: false }
      }
    })
    expect(wrapper.text()).toContain('Inscripción cancelada')
    expect(wrapper.text()).not.toContain('Taller no vigente')

    await wrapper.setProps({
      enrollment: { ...enrollment, status: 'CONFIRMED', isActive: false }
    })
    expect(wrapper.text()).toContain('Taller no vigente')
  })
})
