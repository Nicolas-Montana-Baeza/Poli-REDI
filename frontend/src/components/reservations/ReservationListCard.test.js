// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ReservationListCard from './ReservationListCard.vue'

const reservation = {
  id: 7,
  title: 'Basquetbol',
  resourceName: 'Cancha 2',
  startTime: '2099-07-28T19:15:00',
  durationMinutes: 60,
  status: 'PENDING'
}

describe('ReservationListCard selectable mode', () => {
  it('uses the complete card as the accessible detail action', async () => {
    const wrapper = mount(ReservationListCard, {
      props: { reservation, selectable: true }
    })
    const card = wrapper.get('button.reservation-list-card')

    expect(card.attributes('aria-label')).toBe('Ver detalle de Basquetbol')
    expect(wrapper.text()).not.toContain('Detalle')
    expect(wrapper.find('.detail-link').exists()).toBe(false)
    await card.trigger('click')
    expect(wrapper.emitted('open-detail')[0][0]).toEqual(reservation)
  })

  it('is activated by Enter and Space through native button behavior', () => {
    const wrapper = mount(ReservationListCard, {
      props: { reservation, selectable: true }
    })
    const card = wrapper.get('button.reservation-list-card')

    expect(card.attributes('type')).toBe('button')
    expect(card.attributes('tabindex')).toBeUndefined()
  })

  it('preserves the existing routed-card mode for other contexts', () => {
    const wrapper = mount(ReservationListCard, {
      props: { reservation, detailTo: '/reservations/7' },
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a class="detail-link"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.get('.detail-link').text()).toBe('Detalle')
    expect(wrapper.find('button.reservation-list-card').exists()).toBe(false)
  })
})
