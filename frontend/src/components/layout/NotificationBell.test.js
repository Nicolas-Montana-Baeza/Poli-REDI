import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import NotificationBell from './NotificationBell.vue'

const mocks = vi.hoisted(() => ({
  auth: null,
  notifications: null
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mocks.auth
}))

vi.mock('@/stores/notifications', () => ({
  useNotificationsStore: () => mocks.notifications
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() })
}))

describe('NotificationBell - actualización de notificaciones', () => {
  beforeEach(() => {
    mocks.auth = reactive({
      account: { username: 'usuario@poliredi.local' }
    })

    mocks.notifications = reactive({
      notifications: [],
      unreadCount: 0,
      loading: false,
      error: null,
      fetchNotifications: vi.fn().mockResolvedValue(undefined),
      clearNotifications: vi.fn(),
      markRead: vi.fn()
    })
  })

  it('consulta las notificaciones al abrir, sin exigir F5', async () => {
    const wrapper = mount(NotificationBell)

    await flushPromises()
    expect(mocks.notifications.fetchNotifications).toHaveBeenCalledTimes(1)

    const button = wrapper.get('button[aria-label="Abrir notificaciones"]')

    await button.trigger('click')
    await flushPromises()

    expect(mocks.notifications.fetchNotifications).toHaveBeenCalledTimes(2)
    expect(button.attributes('aria-expanded')).toBe('true')

    await button.trigger('click')
    await flushPromises()

    expect(mocks.notifications.fetchNotifications).toHaveBeenCalledTimes(2)

    wrapper.unmount()
  })

  it('no consulta notificaciones sin una sesión autenticada', async () => {
    mocks.auth.account = null

    const wrapper = mount(NotificationBell)
    await flushPromises()

    await wrapper.get('button[aria-label="Abrir notificaciones"]').trigger('click')
    await flushPromises()

    expect(mocks.notifications.fetchNotifications).not.toHaveBeenCalled()
    wrapper.unmount()
  })
})
