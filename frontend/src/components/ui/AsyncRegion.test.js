// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'

import AsyncRegion from './AsyncRegion.vue'

describe('AsyncRegion', () => {
  it('expone busy y anuncia la carga sin exponer el skeleton', () => {
    const wrapper = mount(AsyncRegion, {
      props: {
        loading: true,
        loadingLabel: 'Cargando instalaciones',
        skeletonVariant: 'media-grid',
        skeletonItems: 2
      }
    })

    expect(wrapper.get('.async-region').attributes('aria-busy')).toBe('true')
    expect(wrapper.get('[role="status"]').text()).toBe('Cargando instalaciones')
    expect(wrapper.get('.skeleton').attributes('aria-hidden')).toBe('true')
    expect(wrapper.findAll('.media-card')).toHaveLength(2)
  })

  it('prioriza loading, luego error, vacío y finalmente contenido', async () => {
    const wrapper = mount(AsyncRegion, {
      props: { loading: true, error: 'Error seguro', empty: true },
      slots: {
        default: '<p class="content">Contenido</p>',
        empty: '<p class="empty">Vacío</p>'
      }
    })

    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(wrapper.find('.content').exists()).toBe(false)

    await wrapper.setProps({ loading: false })
    expect(wrapper.get('[role="alert"]').text()).toBe('Error seguro')
    expect(wrapper.find('.empty').exists()).toBe(false)

    await wrapper.setProps({ error: '' })
    expect(wrapper.get('.empty').text()).toBe('Vacío')

    await wrapper.setProps({ empty: false })
    expect(wrapper.get('.content').text()).toBe('Contenido')
    expect(wrapper.get('.async-region').attributes('aria-busy')).toBe('false')
  })

  it('permite personalizar estados sin duplicar el contrato accesible', () => {
    const wrapper = mount(AsyncRegion, {
      props: { error: 'falló' },
      slots: {
        error: ({ error }) => h('p', { class: 'custom-error' }, String(error))
      }
    })

    expect(wrapper.text()).toContain('falló')
  })
})
