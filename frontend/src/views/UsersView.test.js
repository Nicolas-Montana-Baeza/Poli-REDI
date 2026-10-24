import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import UsersView from './UsersView.vue'

const mocks = vi.hoisted(() => ({ store: null, auth: null }))
vi.mock('@/stores/userStore', () => ({ useUsersStore: () => mocks.store }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.auth }))

let wrapper
beforeEach(() => {
  mocks.auth = reactive({ user: { id: 1, isAdmin: true } })
  mocks.store = reactive({
    users: [
      { id: 1, fullName: 'Admin', email: 'admin@example.org', isAdmin: true, isBlocked: false },
      { id: 2, fullName: 'Estudiante', email: 'estudiante@example.org', isAdmin: false, isBlocked: false },
      { id: 3, fullName: 'Usuario bloqueado', email: 'bloqueado@example.org', isAdmin: false, isBlocked: true }
    ],
    loading: false, error: null, actionError: null, actionSuccess: null,
    updatingId: null, historyLoadingId: null, historyByUser: {}, roleHistoryLoadingId: null, roleHistoryByUser: {},
    fetchUsers: vi.fn(async () => {}),
    clearMessages: vi.fn(),
    changeAccess: vi.fn(async (id, blocked, reason) => ({ user: { id, isBlocked: blocked }, audit: { reason } })),
    loadAccessHistory: vi.fn(async () => []),
    loadRoleHistory: vi.fn(async () => []),
    changeRole: vi.fn(async (id, isAdmin, reason) => ({ user: { id, isAdmin }, audit: { reason } }))
  })
})
afterEach(() => { wrapper?.unmount(); wrapper = null; vi.clearAllMocks() })
const start = async () => { wrapper = mount(UsersView); await flushPromises(); return wrapper }

describe('ADMIN-002 gestión de usuarios', () => {
  it('muestra usuarios y distingue bloqueados', async () => {
    await start()
    expect(wrapper.text()).toContain('Usuario bloqueado')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.findAll('.user-card')).toHaveLength(3)
  })
  it('no permite cambiar su propia cuenta', async () => {
    await start()
    const own = wrapper.findAll('.user-card')[0]
    expect(own.text()).toContain('Tu cuenta')
    expect(own.findAll('button')[1].attributes('disabled')).toBeDefined()
  })
  it('requiere motivo y confirmación antes de bloquear', async () => {
    await start()
    const student = wrapper.findAll('.user-card')[1]
    await student.findAll('button')[1].trigger('click')
    expect(mocks.store.changeAccess).not.toHaveBeenCalled()
    const form = wrapper.get('.decision-form')
    expect(form.find('button[type="submit"]').attributes('disabled')).toBeDefined()
    await form.get('textarea').setValue('Incumplimiento registrado')
    await form.trigger('submit')
    await flushPromises()
    expect(mocks.store.changeAccess).toHaveBeenCalledWith(2, true, 'Incumplimiento registrado')
  })
  it('permite solicitar desbloqueo', async () => {
    await start()
    const target = wrapper.findAll('.user-card')[2]
    await target.findAll('button')[1].trigger('click')
    await wrapper.get('.decision-form textarea').setValue('Revisión administrativa')
    await wrapper.get('.decision-form').trigger('submit')
    await flushPromises()
    expect(mocks.store.changeAccess).toHaveBeenCalledWith(3, false, 'Revisión administrativa')
  })
  it('consulta historial bajo demanda', async () => {
    await start()
    await wrapper.findAll('.user-card')[1].findAll('button')[0].trigger('click')
    await flushPromises()
    expect(mocks.store.loadAccessHistory).toHaveBeenCalledWith(2)
    expect(mocks.store.loadRoleHistory).toHaveBeenCalledWith(2)
  })
  it('permite solicitar promoción con motivo explícito', async () => {
    await start()
    await wrapper.findAll('.user-card')[1].findAll('button')[2].trigger('click')
    expect(mocks.store.changeRole).not.toHaveBeenCalled()
    const form = wrapper.get('.decision-form')
    expect(form.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await form.get('textarea').setValue('Nombramiento por coordinación deportiva')
    await form.trigger('submit')
    await flushPromises()
    expect(mocks.store.changeRole).toHaveBeenCalledWith(2, true, 'Nombramiento por coordinación deportiva')
  })
  it('permite retirar el rol de otro administrador', async () => {
    mocks.store.users.push({ id: 4, fullName: 'Segundo admin', email: 'segundo@example.org', isAdmin: true, isBlocked: false })
    await start()
    await wrapper.findAll('.user-card')[3].findAll('button')[2].trigger('click')
    await wrapper.get('.decision-form textarea').setValue('Finalización de responsabilidad')
    await wrapper.get('.decision-form').trigger('submit')
    await flushPromises()
    expect(mocks.store.changeRole).toHaveBeenCalledWith(4, false, 'Finalización de responsabilidad')
  })
  it('no permite modificar el propio rol ni promover a un bloqueado', async () => {
    await start()
    const cards = wrapper.findAll('.user-card')
    expect(cards[0].findAll('button')[2].attributes('disabled')).toBeDefined()
    expect(cards[2].findAll('button')[2].attributes('disabled')).toBeDefined()
  })
})
