// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import SkeletonLoader from './SkeletonLoader.vue'

describe('SkeletonLoader', () => {
  it.each([
    ['availability-timelines', '.timeline-card'],
    ['media-grid', '.media-card'],
    ['card-grid', '.content-card'],
    ['list', '.list-card'],
    ['detail', '.detail-panel'],
    ['metrics-table', '.metric-card'],
    ['compact-rows', '.compact-row']
  ])('renderiza el arquetipo %s como contenido decorativo', (variant, selector) => {
    const wrapper = mount(SkeletonLoader, { props: { variant, items: 4 } })

    expect(wrapper.get('.skeleton').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find(selector).exists()).toBe(true)
    expect(wrapper.findAll('button, a, input, select, textarea, [tabindex]')).toHaveLength(0)
  })

  it('mantiene aliases antiguos sin reintroducir controles falsos', () => {
    const wrapper = mount(SkeletonLoader, {
      props: { variant: 'reservations', items: 2 }
    })

    expect(wrapper.findAll('.list-card')).toHaveLength(2)
    expect(wrapper.find('.skeleton-list').exists()).toBe(true)
    expect(wrapper.find('.button').exists()).toBe(false)
  })

  it('limita cantidades y configura columnas/carrusel sin estilos de acción', async () => {
    const wrapper = mount(SkeletonLoader, {
      props: {
        variant: 'media-grid',
        items: 20,
        columns: 4,
        mobileCarousel: true
      }
    })

    expect(wrapper.findAll('.media-card')).toHaveLength(12)
    expect(wrapper.get('.skeleton').classes()).toContain('fixed-columns')
    expect(wrapper.get('.skeleton').classes()).toContain('mobile-carousel')
    expect(wrapper.get('.skeleton').attributes('style')).toContain('--skeleton-columns: 4')
  })

  it('declara breakpoints objetivo y desactiva movimiento reducido', () => {
    const source = readFileSync(
      resolve(process.cwd(), 'src/components/ui/SkeletonLoader.vue'),
      'utf8'
    )

    expect(source).toContain('@media (max-width: 1024px)')
    expect(source).toContain('@media (max-width: 900px)')
    expect(source).toContain('@media (max-width: 768px)')
    expect(source).toContain('@media (max-width: 600px)')
    expect(source).toContain('flex: 0 0 min(82vw, 280px)')
    expect(source).toContain('@media (prefers-reduced-motion: reduce)')
    expect(source).toMatch(/prefers-reduced-motion[\s\S]*animation:\s*none/)
  })
})
