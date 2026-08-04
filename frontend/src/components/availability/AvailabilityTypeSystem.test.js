// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import AvailabilityTypeLegend from './AvailabilityTypeLegend.vue'
import GeneralCalendarView from './GeneralCalendarView.vue'
import ReservationBlock from './ReservationBlock.vue'
import ResourceTimeline from './ResourceTimeline.vue'
import {
  getAvailabilityDisplayTitle,
  getAvailabilityType
} from '@/utils/availabilityType'

const futureReservation = {
  id: 1,
  resourceId: 1,
  startTime: '2099-07-28T19:30:00',
  durationMinutes: 60,
  status: 'CONFIRMED',
  title: 'Basquetbol'
}

describe('getAvailabilityType', () => {
  it.each([
    ['CLASS', 'class', 'Clase'],
    ['WORKSHOP', 'workshop', 'Taller'],
    ['TRAINING', 'training', 'Entrenamiento'],
    ['CHAMPIONSHIP', 'championship', 'Campeonato'],
    ['EVENT', 'event', 'Evento'],
    ['OTHER', 'institutional', 'Actividad institucional'],
    ['UNKNOWN', 'institutional', 'Actividad institucional']
  ])('maps scheduled type %s without title inference', (
    activityType,
    key,
    label
  ) => {
    expect(getAvailabilityType({
      availabilityKind: 'SCHEDULED_ACTIVITY',
      activityType
    })).toMatchObject({ key, label, family: expect.any(String) })
  })

  it('uses the explicit workshop flag before other structural fields', () => {
    expect(getAvailabilityType({
      isWorkshop: true,
      availabilityKind: 'GROUP_RESERVATION',
      activityType: 'CLASS'
    }).key).toBe('workshop')
  })

  it('recognizes an external group even when private metrics are absent', () => {
    expect(getAvailabilityType({
      availabilityKind: 'GROUP_RESERVATION',
      targetParticipants: null
    }).key).toBe('group')
  })

  it.each(['PENDING', 'CONFIRMED'])(
    'keeps the same group type when status is %s',
    (status) => {
      expect(getAvailabilityType({
        availabilityKind: 'GROUP_RESERVATION',
        status
      }).key).toBe('group')
    }
  )

  it('keeps compatibility with local group items without availabilityKind', () => {
    expect(getAvailabilityType({
      targetParticipants: 10
    }).key).toBe('group')
  })

  it('uses resource context for open use', () => {
    expect(getAvailabilityType(
      { availabilityKind: 'RESERVATION' },
      { reservationMode: 'OPEN_USE' }
    ).key).toBe('open-use')
  })

  it('does not infer type from title, status, legacy type or personal data', () => {
    const type = getAvailabilityType({
      title: 'Taller grupal de campeonato',
      status: 'PENDING',
      type: 'workshop',
      userFullName: 'Nombre privado',
      userEmail: 'privado@example.com',
      userRut: '11.111.111-1'
    })

    expect(type.key).toBe('reservation')
    expect(JSON.stringify(type)).not.toContain('privado')
  })
})

describe('getAvailabilityDisplayTitle', () => {
  it.each([
    [{ isWorkshop: true, title: 'Taller: Entrenamiento funcional' }, 'Entrenamiento funcional'],
    [{ isWorkshop: true, title: '  taller  -  Judo  ' }, 'Judo'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', activityType: 'WORKSHOP', title: 'TALLER: Esgrima' }, 'Esgrima'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', activityType: 'CLASS', title: 'Clase: Yoga' }, 'Yoga'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', activityType: 'TRAINING', title: 'Entrenamiento - Selección' }, 'Selección'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', activityType: 'CHAMPIONSHIP', title: 'Campeonato: Apertura' }, 'Apertura'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', activityType: 'EVENT', title: 'Evento - Bienvenida' }, 'Bienvenida']
  ])('removes only the canonical leading prefix from structured activities', (item, expected) => {
    const original = { ...item }

    expect(getAvailabilityDisplayTitle(item)).toBe(expected)
    expect(item).toEqual(original)
  })

  it('does not infer type from the title or remove internal text', () => {
    expect(getAvailabilityDisplayTitle({
      title: 'Taller: Reserva privada'
    })).toBe('Taller: Reserva privada')
    expect(getAvailabilityDisplayTitle({
      isWorkshop: true,
      title: 'Encuentro Taller: avanzado'
    })).toBe('Encuentro Taller: avanzado')
    expect(getAvailabilityDisplayTitle({
      availabilityKind: 'SCHEDULED_ACTIVITY',
      activityType: 'CLASS',
      title: 'Taller: Yoga'
    })).toBe('Taller: Yoga')
  })
})

describe('AvailabilityTypeChip', () => {
  it('is textual, non-focusable and can be hidden from duplicate announcements', () => {
    const wrapper = mount(AvailabilityTypeChip, {
      props: {
        item: { availabilityKind: 'GROUP_RESERVATION' },
        ariaHidden: true
      }
    })

    expect(wrapper.element.tagName).toBe('SPAN')
    expect(wrapper.text()).toBe('Reserva grupal')
    expect(wrapper.attributes('tabindex')).toBeUndefined()
    expect(wrapper.attributes('aria-hidden')).toBe('true')
  })
})

describe('ReservationBlock availability type', () => {
  it('separates type from status and exposes one safe accessible name', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          availabilityKind: 'GROUP_RESERVATION',
          targetParticipants: null,
          userFullName: 'Nombre privado',
          userEmail: 'privado@example.com',
          userRut: '11.111.111-1'
        },
        resource: {
          id: 1,
          reservationMode: 'RESERVABLE'
        }
      }
    })

    const button = wrapper.get('button')
    const label = button.attributes('aria-label')

    expect(wrapper.get('[data-availability-type="group"]').text())
      .toBe('Reserva grupal')
    expect(wrapper.get('[data-availability-type="group"]')
      .attributes('aria-hidden')).toBe('true')
    expect(wrapper.get('.block-heading').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="group"]').element)
    expect(button.classes()).toContain('confirmed')
    expect(label).toContain('Reserva grupal')
    expect(label).toContain('Basquetbol')
    expect(label).toContain('Confirmada')
    expect(label).toContain('19:30 - 20:30')
    expect(label).toContain('Abrir detalle')
    expect(label).not.toContain('Nombre privado')
    expect(label).not.toContain('privado@example.com')
    expect(label).not.toContain('11.111.111-1')
  })

  it('keeps the chip visible in compact blocks and hides only the time line', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          durationMinutes: 30
        }
      }
    })

    expect(wrapper.get('[data-availability-type="reservation"]').exists())
      .toBe(true)
    expect(wrapper.get('.block-heading').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="reservation"]').element)
    expect(wrapper.get('button').classes()).toContain('compact')
  })

  it.each([
    [{ isWorkshop: true }, 'local workshop'],
    [{ availabilityKind: 'SCHEDULED_ACTIVITY', isScheduledActivity: true, activityType: 'WORKSHOP' }, 'scheduled workshop']
  ])('uses one workshop signal and a clean title for a %s', (structuralFields) => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          ...structuralFields,
          title: 'Taller: Entrenamiento funcional'
        }
      }
    })

    const label = wrapper.get('button').attributes('aria-label')

    expect(wrapper.get('[data-availability-type="workshop"]').text())
      .toBe('Taller')
    expect(wrapper.get('.block-heading strong').text())
      .toBe('Entrenamiento funcional')
    expect(label.match(/Taller/g)).toHaveLength(1)
    expect(label).toContain('Entrenamiento funcional')
    expect(label).not.toContain('Taller: Entrenamiento funcional')
  })

  it('shows an accessible enrolled state with text and icon', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          isWorkshop: true,
          title: 'Taller: Entrenamiento funcional',
          workshop: {
            id: 3,
            isEnrolled: true
          }
        }
      }
    })

    const badge = wrapper.get('[data-workshop-enrollment="enrolled"]')

    expect(badge.text()).toBe('Inscrito')
    expect(badge.attributes('title')).toBe('Ya estás inscrito en este taller')
    expect(badge.find('svg').exists()).toBe(true)
    expect(wrapper.get('button').attributes('aria-label')).toContain('Inscrito')
    expect(wrapper.get('.block-heading').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="workshop"]').element)
    expect(wrapper.get('.block-meta').element.lastElementChild)
      .toBe(badge.element)
  })

  it('uses absence of the badge to represent a workshop without enrollment', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          isWorkshop: true,
          title: 'Taller: Entrenamiento funcional',
          workshop: {
            id: 3,
            isEnrolled: false
          }
        }
      }
    })

    expect(wrapper.find('[data-workshop-enrollment]').exists()).toBe(false)
    expect(wrapper.get('button').attributes('aria-label'))
      .not.toContain('Inscrito')
  })

  it('does not show workshop enrollment state for other availability types', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          availabilityKind: 'GROUP_RESERVATION',
          workshop: { id: 3, isEnrolled: true }
        }
      }
    })

    expect(wrapper.find('[data-workshop-enrollment]').exists()).toBe(false)
    expect(wrapper.get('button').attributes('aria-label'))
      .not.toContain('Inscrito')
  })

  it('keeps ordinary reservation titles intact', () => {
    const wrapper = mount(ReservationBlock, {
      props: {
        reservation: {
          ...futureReservation,
          title: 'Taller: Reserva privada'
        }
      }
    })

    expect(wrapper.get('.block-heading strong').text())
      .toBe('Taller: Reserva privada')
  })

  it('keeps compact content within the block and above the time ruler', () => {
    const source = readFileSync(
      join(process.cwd(), 'src/components/availability/ReservationBlock.vue'),
      'utf8'
    )

    expect(source).toContain('text-overflow: ellipsis')
    expect(source).toContain('white-space: nowrap')
    expect(source).toContain('flex: 0 0 auto')
    expect(source).toContain('max-width: none')
    expect(source).toContain('box-sizing: border-box')
    expect(source).toContain('z-index: 5')
  })
})

describe('availability type legend', () => {
  const resources = [
    { id: 1, reservationMode: 'RESERVABLE' },
    { id: 2, reservationMode: 'OPEN_USE' }
  ]

  const reservations = [
    {
      ...futureReservation,
      id: 1,
      availabilityKind: 'GROUP_RESERVATION',
      targetParticipants: null,
      status: 'PENDING'
    },
    {
      ...futureReservation,
      id: 2,
      availabilityKind: 'GROUP_RESERVATION'
    },
    {
      ...futureReservation,
      id: 3,
      isWorkshop: true
    },
    {
      ...futureReservation,
      id: 4,
      availabilityKind: 'SCHEDULED_ACTIVITY',
      activityType: 'UNKNOWN'
    },
    {
      ...futureReservation,
      id: 5,
      startTime: '2099-07-29T19:30:00'
    }
  ]

  it('deduplicates visible types, keeps stable order and includes open use', () => {
    const wrapper = mount(AvailabilityTypeLegend, {
      props: {
        resources,
        reservations,
        selectedDate: '2099-07-28'
      }
    })

    expect(wrapper.text()).toContain('Tipos de bloque')
    expect(wrapper.findAll('[data-availability-type="group"]'))
      .toHaveLength(1)
    expect(wrapper.findAll('[data-availability-type="open-use"]'))
      .toHaveLength(1)
    expect(wrapper.findAll('[data-availability-type="workshop"]'))
      .toHaveLength(1)
    expect(wrapper.findAll('[data-availability-type="institutional"]'))
      .toHaveLength(1)
    expect(wrapper.findAll('[data-availability-type]')
      .map(chip => chip.attributes('data-availability-type')))
      .toEqual([
        'group',
        'open-use',
        'workshop',
        'institutional'
      ])
  })

  it('uses horizontal overflow and a mobile layout without fake controls', () => {
    const source = readFileSync(
      join(
        process.cwd(),
        'src',
        'components',
        'availability',
        'AvailabilityTypeLegend.vue'
      ),
      'utf8'
    )

    expect(source).toContain('overflow-x: auto')
    expect(source).toContain('@media (max-width: 768px)')
    expect(source).not.toContain('<button')
  })
})

describe('availability views', () => {
  it('keeps agenda titles flexible and indicators fixed on narrow screens', () => {
    const source = readFileSync(
      join(
        process.cwd(),
        'src',
        'components',
        'availability',
        'GeneralCalendarView.vue'
      ),
      'utf8'
    )

    expect(source).toContain('.reservation-title-row h3')
    expect(source).toContain('flex: 1 1 auto')
    expect(source).toContain('.reservation-indicators')
    expect(source).toContain('flex: 0 0 auto')
    expect(source).toContain('@media (max-width: 480px)')
    expect(source).toContain('flex-direction: column-reverse')
  })

  it('uses the same type in the daily agenda and emits the selected item', async () => {
    const wrapper = mount(GeneralCalendarView, {
      props: {
        resources: [
          { id: 1, reservationMode: 'OPEN_USE' }
        ],
        reservations: [futureReservation],
        selectedDate: '2099-07-28'
      }
    })

    const item = wrapper.get('.agenda-item')
    const label = item.attributes('aria-label')

    expect(wrapper.get('[data-availability-type="open-use"]').text())
      .toBe('Uso libre')
    expect(wrapper.get('.reservation-indicators').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="open-use"]').element)
    expect(label).toContain('Uso libre')
    expect(label).toContain('Basquetbol')
    expect(label).toContain('Confirmada')
    expect(label).toContain('Abrir detalle')

    await item.trigger('click')
    expect(wrapper.emitted('reservation-selected')[0][0])
      .toEqual(futureReservation)
  })

  it('uses the chip and clean activity name in the daily agenda', () => {
    const wrapper = mount(GeneralCalendarView, {
      props: {
        resources: [{ id: 1, reservationMode: 'RESERVABLE' }],
        reservations: [{
          ...futureReservation,
          isWorkshop: true,
          title: 'Taller: Entrenamiento funcional',
          workshop: {
            id: 3,
            isEnrolled: false
          }
        }],
        selectedDate: '2099-07-28'
      }
    })

    const label = wrapper.get('.agenda-item').attributes('aria-label')

    expect(wrapper.get('[data-availability-type="workshop"]').text())
      .toBe('Taller')
    expect(wrapper.get('.reservation-title-row h3').text())
      .toBe('Entrenamiento funcional')
    expect(wrapper.find('[data-workshop-enrollment]').exists()).toBe(false)
    expect(wrapper.get('.reservation-indicators').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="workshop"]').element)
    expect(label.match(/Taller/g)).toHaveLength(1)
  })

  it('shows one open-use chip in the resource caption, not per attendance', () => {
    const wrapper = mount(ResourceTimeline, {
      props: {
        resource: {
          id: 2,
          name: 'Piscina',
          type: 'Piscina',
          status: 'available',
          reservationMode: 'OPEN_USE'
        },
        reservations: [
          {
            ...futureReservation,
            id: 20,
            resourceId: 2
          },
          {
            ...futureReservation,
            id: 21,
            resourceId: 2
          }
        ],
        selectedDate: '2099-07-28'
      }
    })

    expect(wrapper.findAll('[data-availability-type="open-use"]'))
      .toHaveLength(1)
    expect(wrapper.get('.open-use-mode').element.lastElementChild)
      .toBe(wrapper.get('[data-availability-type="open-use"]').element)
    expect(wrapper.findAllComponents(ReservationBlock)).toHaveLength(0)
    expect(wrapper.text()).toContain(
      'La intensidad indica la cantidad de reservas simultáneas.'
    )
  })
})
