import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const { api, getCurrentAccount, logout } = vi.hoisted(() => ({
  api: {
    get: vi.fn(),
    patch: vi.fn()
  },
  getCurrentAccount: vi.fn(),
  logout: vi.fn()
}))

vi.mock('@/services/api', () => ({ default: api }))
vi.mock('@/auth/authService', () => ({
  getCurrentAccount,
  logout
}))

import { useAuthStore } from '@/stores/auth'

const deferred = () => {
  let resolve
  const promise = new Promise((done) => {
    resolve = done
  })
  return { promise, resolve }
}

describe('auth store profile lifecycle', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    getCurrentAccount.mockResolvedValue({
      homeAccountId: 'account-a',
      username: 'student@example.test'
    })
  })

  it('deduplicates concurrent /me requests for the same account', async () => {
    const request = deferred()
    api.get.mockReturnValue(request.promise)
    const store = useAuthStore()

    const first = store.loadAuthUser()
    const second = store.loadAuthUser()
    request.resolve({ data: { id: 1, rut: '' } })

    await Promise.all([first, second])
    expect(api.get).toHaveBeenCalledTimes(1)
    expect(store.profileReady).toBe(true)
  })

  it('does not restore a stale profile after logout', async () => {
    const request = deferred()
    api.get.mockReturnValue(request.promise)
    const store = useAuthStore()

    const loading = store.loadAuthUser()
    await Promise.resolve()
    await store.logoutUser()
    request.resolve({ data: { id: 1, rut: '12345678-5' } })
    await loading

    expect(store.user).toBeNull()
    expect(store.profileReady).toBe(false)
  })

  it('refreshes /me after a write-once conflict', async () => {
    api.get
      .mockResolvedValueOnce({ data: { id: 1, rut: '' } })
      .mockResolvedValueOnce({ data: { id: 1, rut: '12345678-5' } })
    api.patch.mockRejectedValue({
      response: { status: 409, data: { error: 'El RUT ya fue registrado.' } }
    })
    const store = useAuthStore()
    await store.loadAuthUser()

    await expect(store.updateRut('11111111-1'))
      .rejects.toThrow('El RUT ya fue registrado.')
    expect(api.get).toHaveBeenCalledTimes(2)
    expect(store.user.rut).toBe('12345678-5')
    expect(store.profileReady).toBe(true)
  })
})
