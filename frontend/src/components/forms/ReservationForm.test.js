// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ReservationForm from './ReservationForm.vue'

const groupResource = {
  id: 2,
  name: 'Cancha 2, Centro Deportivo',
  capacity: 22,
  reservationMode: 'RESERVABLE'
}

const mountForm = (overrides = {}) => mount(ReservationForm, {
  props: {
    visible: true,
    slot: { resource: groupResource, date: '2099-07-28', hour: '12:00' },
    resources: [groupResource],
    activities: [],
    policy: {
      minimumParticipants: 10,
      groupResourceIds: ['1', '2', '7']
    },
    ...overrides
  },
  global: {
    stubs: {
      Teleport: true,
      ResourcePicker: true,
      DateTimePicker: true
    }
  }
})

describe('ReservationForm group reservations', () => {
  it('shows an editable participant target with default 10', () => {
    const wrapper = mountForm()
    const input = wrapper.get('#target-participants')

    expect(input.isVisible()).toBe(true)
    expect(input.element.value).toBe('10')
    expect(input.attributes('min')).toBe('10')
    expect(input.attributes('max')).toBe('22')
  })

  it.each([
    [['1', '2', '7'], 2],
    [[1, 2, 7], '2']
  ])('classifies Cancha 2 as group with policy IDs %j and resource ID %s', async (groupResourceIds, resourceId) => {
    const resource = { ...groupResource, id: resourceId }
    const wrapper = mountForm({
      slot: { resource, date: '2099-07-28', hour: '18:45' },
      resources: [resource],
      policy: { minimumParticipants: 10, groupResourceIds }
    })

    expect(wrapper.get('#target-participants').element.value).toBe('10')
    await wrapper.get('.submit-btn').trigger('click')
    expect(wrapper.emitted('submit')[0][0].targetParticipants).toBe(10)
  })

  it('submits the selected target participant count', async () => {
    const wrapper = mountForm()
    await wrapper.get('#target-participants').setValue('17')
    await wrapper.get('.submit-btn').trigger('click')

    expect(wrapper.emitted('submit')).toHaveLength(1)
    expect(wrapper.emitted('submit')[0][0].targetParticipants).toBe(17)
  })

  it('rejects a target above the frozen resource capacity', async () => {
    const wrapper = mountForm()
    await wrapper.get('#target-participants').setValue('23')
    await wrapper.get('.submit-btn').trigger('click')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('Ingresa un objetivo entre 10 y 22.')
  })
})
