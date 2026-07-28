// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ResourcePicker from './ResourcePicker.vue'

const resources = [
  { id: 1, name: 'Cancha 1, Centro Deportivo', type: 'Cancha' },
  { id: 2, name: 'Sala Spinning, Centro Deportivo', type: 'Sala' },
  { id: 3, name: 'Piscina, Centro Deportivo', type: 'Piscina' },
  { id: 4, name: 'Muro Escalada, Centro Deportivo', type: 'Muro Escalada' }
]

describe('ResourcePicker', () => {
  it('shows only resource names and preserves the selected resource', () => {
    const wrapper = mount(ResourcePicker, { props: { resources, selectedId: 1 } })

    expect(wrapper.get('label').text()).toBe('Instalación')
    expect(wrapper.get('select').element.value).toBe('1')
    expect(wrapper.findAll('option').map(option => option.text())).toEqual([
      'Selecciona una instalación',
      ...resources.map(resource => resource.name)
    ])
    expect(wrapper.text()).not.toContain('Disponible')
    expect(wrapper.find('input[type="search"]').exists()).toBe(false)
  })

  it('emits the complete selected resource', async () => {
    const wrapper = mount(ResourcePicker, { props: { resources } })

    await wrapper.get('select').setValue('3')

    expect(wrapper.emitted('select')[0][0]).toEqual(resources[2])
  })

  it('supports disabled, loading, empty and error states', async () => {
    const loading = mount(ResourcePicker, {
      props: { resources, loading: true }
    })
    expect(loading.get('select').attributes('disabled')).toBeDefined()
    expect(loading.get('option').text()).toBe('Cargando instalaciones...')

    const empty = mount(ResourcePicker)
    expect(empty.get('select').attributes('disabled')).toBeDefined()
    expect(empty.text()).toContain('No hay instalaciones disponibles.')

    const invalid = mount(ResourcePicker, {
      props: { resources, error: 'Selecciona una instalación.' }
    })
    expect(invalid.get('select').attributes('aria-invalid')).toBe('true')
    expect(invalid.get('[role="alert"]').text()).toBe('Selecciona una instalación.')
    expect(invalid.text()).toContain('Selecciona una instalación.')
  })
})
