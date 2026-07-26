// @vitest-environment jsdom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import ConfirmModal from './ConfirmModal.vue'

const mounted = []
const open = async props => {
  const trigger = document.createElement('button')
  document.body.appendChild(trigger)
  trigger.focus()
  const wrapper = mount(ConfirmModal, { attachTo: document.body, props: { show: true, destructive: true, ...props } })
  mounted.push(wrapper)
  await wrapper.vm.$nextTick()
  return { wrapper, trigger, dialog: document.querySelector('[role="dialog"]') }
}
afterEach(() => {
  mounted.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
  document.body.style.overflow = ''
})
describe('ConfirmModal', () => {
  it('expone semántica, bloquea scroll y enfoca la opción segura', async () => {
    const { dialog } = await open()
    expect(dialog.getAttribute('aria-modal')).toBe('true')
    expect(dialog.getAttribute('aria-labelledby')).toBeTruthy()
    expect(dialog.getAttribute('aria-describedby')).toBeTruthy()
    expect(document.body.style.overflow).toBe('hidden')
    expect(document.activeElement.textContent).toContain('Volver')
  })
  it('atrapa Tab, permite Escape y restaura foco', async () => {
    const { wrapper, trigger, dialog } = await open()
    const buttons = dialog.querySelectorAll('button')
    buttons[1].focus()
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(document.activeElement).toBe(buttons[0])
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('cancel')).toHaveLength(1)
    await wrapper.setProps({ show: false })
    expect(document.activeElement).toBe(trigger)
  })
  it('durante loading ignora Escape y backdrop', async () => {
    const { wrapper } = await open({ loading: true })
    document.querySelector('[role="dialog"]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    await document.querySelector('.modal-backdrop').click()
    expect(wrapper.emitted('cancel')).toBeUndefined()
  })
})
