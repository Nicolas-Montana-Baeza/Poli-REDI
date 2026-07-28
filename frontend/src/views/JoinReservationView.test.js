// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({
  route: null,
  get: vi.fn(),
  confirm: vi.fn(),
  withdraw: vi.fn()
}))
vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  state.route = reactive({ params: {} })
  return { useRoute: () => state.route }
})
vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    getGroupProgress: state.get,
    confirmGroup: state.confirm,
    withdrawGroup: state.withdraw
  }
}))

import source from './JoinReservationView.vue?raw'
import JoinReservationView from './JoinReservationView.vue'

const mountJoin = (options = {}) => mount(JoinReservationView, {
  ...options,
  global: {
    ...options.global,
    stubs: {
      Teleport: true,
      ...options.global?.stubs
    }
  }
})

const progress = overrides => ({
  participantCount: 1, minimumParticipants: 10, targetParticipants: 12,
  capacity: 20, status: 'PENDING', confirmationDeadline: '2099-01-01T00:00:00Z',
  isMember: false, isOwner: false, ...overrides
})
const submit = wrapper => wrapper.get('form').trigger('submit')
const deferred = () => {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

describe('JoinReservationView', () => {
  beforeEach(() => {
    state.route.params = {}
    vi.clearAllMocks()
    localStorage.clear()
    sessionStorage.clear()
  })

  it('renderiza sin main anidado y cumple el contrato responsive de 320 px', () => {
    const wrapper = mountJoin({ attachTo: document.body })
    expect(wrapper.find('main').exists()).toBe(false)
    expect(wrapper.get('article.app-card').exists()).toBe(true)
    expect(wrapper.get('[role="status"]').text()).toContain('Ingresa un código')
    expect(source).toContain('max-width:820px')
    expect(source).toContain('@media(max-width:600px)')
    expect(source).toContain('grid-template-columns:minmax(0,1fr)')
    expect(source).toContain('min-width:0')
    expect(source).toContain('text-overflow:ellipsis')
    wrapper.unmount()
  })

  it('trata el token como opaco, recorta solo al enviar y no persiste ni cambia la URL manual', async () => {
    const token = `  AbC/${'x'.repeat(120)}  `
    state.get.mockResolvedValue(progress())
    const wrapper = mountJoin({ attachTo: document.body })
    const input = wrapper.get('#join-code')
    await input.setValue(token)
    expect(input.element.value).toBe(token)
    const consultButton = wrapper.findAll('button').find(item => item.text() === 'Consultar reserva')
    consultButton.element.focus()
    await consultButton.trigger('click')
    await flushPromises()
    expect(state.get).toHaveBeenCalledWith(token.trim())
    expect(state.route.params.code).toBeUndefined()
    expect(localStorage.length).toBe(0)
    expect(sessionStorage.length).toBe(0)
    expect(input.attributes('autocomplete')).toBe('off')
    expect(input.attributes('autocapitalize')).toBe('off')
    expect(input.attributes('spellcheck')).toBe('false')
    expect(input.attributes('aria-describedby')).toBe('join-code-help')
    expect(document.activeElement).toBe(consultButton.element)
    expect(wrapper.find('h2').attributes('tabindex')).toBeUndefined()
    expect(wrapper.get('.sr-only').text()).toBe('Reserva encontrada. 1 participante confirmado; mínimo 10; objetivo 12; capacidad 20.')
    wrapper.unmount()
  })

  it('consulta automáticamente un código explícito en la URL', async () => {
    state.route.params = { code: 'UrlToken-123' }
    state.get.mockResolvedValue(progress())
    const wrapper = mountJoin()
    await flushPromises()
    expect(state.get).toHaveBeenCalledWith('UrlToken-123')
    expect(wrapper.get('#join-code').element.value).toBe('UrlToken-123')
    wrapper.unmount()
  })

  it.each(['resolve', 'reject'])('A→B descarta el %s tardío de A y conserva B', async lateAction => {
    const requestA = deferred()
    const requestB = deferred()
    state.route.params = { code: 'token-A-123' }
    state.get.mockReturnValueOnce(requestA.promise).mockReturnValueOnce(requestB.promise)
    const wrapper = mountJoin()
    await flushPromises()
    state.route.params.code = 'token-B-456'
    await flushPromises()
    expect(state.get).toHaveBeenNthCalledWith(2, 'token-B-456')
    requestB.resolve(progress({ participantCount: 5 }))
    await flushPromises()
    if (lateAction === 'resolve') requestA.resolve(progress({ participantCount: 99 }))
    else requestA.reject({ response: { status: 404 } })
    await flushPromises()
    expect(wrapper.get('#join-code').element.value).toBe('token-B-456')
    expect(wrapper.text()).toContain('5 de 12 confirmados')
    expect(wrapper.text()).not.toContain('99 de 12 confirmados')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('form').attributes('aria-busy')).toBe('false')
    wrapper.unmount()
  })

  it('valida vacío y longitud mínima enfocando el campo', async () => {
    const wrapper = mountJoin({ attachTo: document.body })
    await submit(wrapper)
    expect(wrapper.text()).toContain('Ingresa un código de invitación.')
    expect(document.activeElement).toBe(wrapper.get('#join-code').element)
    expect(wrapper.get('#join-code').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#join-code').attributes('aria-describedby')).toContain('join-code-error')
    await wrapper.get('#join-code').setValue('corto')
    await submit(wrapper)
    expect(wrapper.text()).toContain('al menos 8 caracteres')
    wrapper.unmount()
  })

  it('anuncia loading, error de red y errores HTTP seguros con foco', async () => {
    let resolveRequest
    state.get.mockReturnValueOnce(new Promise(resolve => { resolveRequest = resolve }))
    const wrapper = mountJoin({ attachTo: document.body })
    await wrapper.get('#join-code').setValue('token-valido')
    await submit(wrapper)
    expect(wrapper.get('form').attributes('aria-busy')).toBe('true')
    expect(wrapper.get('[role="status"]').text()).toContain('Consultando')
    resolveRequest(progress())
    await flushPromises()

    const cases = [
      [{}, 'No pudimos conectar con el servidor'],
      [{ response: { status: 404 } }, 'no existe'],
      [{ response: { status: 403 } }, 'cuenta activa'],
      [{ response: { status: 409, data: { code: 'CAPACITY_REACHED', error: '<detalle SQL>' } } }, 'estado actual'],
      [{ response: { status: 409, data: { code: 'DESCONOCIDO', error: '<detalle SQL>' } } }, 'estado actual'],
      [{ response: { status: 410 } }, 'plazo de confirmación']
    ]
    for (const [error, message] of cases) {
      state.get.mockRejectedValueOnce(error)
      await submit(wrapper)
      await flushPromises()
      expect(wrapper.get('[role="alert"]').text()).toContain(message)
      expect(wrapper.text()).not.toContain('<detalle SQL>')
      expect(wrapper.findComponent({ name: 'ParticipantsProgress' }).exists()).toBe(false)
      expect(document.activeElement).toBe(wrapper.get('[role="alert"]').element)
    }
    wrapper.unmount()
  })

  it('anuncia confirmación y retiro y conserva las acciones de la respuesta', async () => {
    state.get.mockResolvedValue(progress())
    state.confirm.mockResolvedValue(progress({ isMember: true, participantCount: 2 }))
    state.withdraw.mockResolvedValue(progress({ isMember: false, participantCount: 1 }))
    const wrapper = mountJoin()
    await wrapper.get('#join-code').setValue('token-valido')
    await submit(wrapper)
    await flushPromises()
    await wrapper.findAll('button').find(item => item.text() === 'Confirmar participación').trigger('click')
    await flushPromises()
    expect(state.confirm).toHaveBeenCalledWith('token-valido')
    expect(wrapper.text()).toContain('Participación confirmada.')
    await wrapper.findAll('button').find(item => item.text() === 'Retirar participación').trigger('click')
    await flushPromises()
    expect(state.withdraw).toHaveBeenCalledWith('token-valido')
    expect(wrapper.text()).toContain('Participación retirada.')
    wrapper.unmount()
  })

  it('una confirmación tardía de A no reemplaza el progreso de B', async () => {
    const pendingConfirm = deferred()
    state.route.params = { code: 'token-A-123' }
    state.get
      .mockResolvedValueOnce(progress({ participantCount: 1 }))
      .mockResolvedValueOnce(progress({ participantCount: 5 }))
    state.confirm.mockReturnValueOnce(pendingConfirm.promise)
    const wrapper = mountJoin()
    await flushPromises()
    await wrapper.findAll('button').find(item => item.text() === 'Confirmar participación').trigger('click')
    await flushPromises()
    state.route.params.code = 'token-B-456'
    await flushPromises()
    pendingConfirm.resolve(progress({ participantCount: 99, isMember: true }))
    await flushPromises()
    expect(wrapper.text()).toContain('5 de 12 confirmados')
    expect(wrapper.text()).not.toContain('99 de 12 confirmados')
    expect(wrapper.text()).not.toContain('Participación confirmada.')
    expect(wrapper.get('form').attributes('aria-busy')).toBe('false')
    wrapper.unmount()
  })

  it('un retiro tardío no muta estado después de desmontar', async () => {
    const pendingWithdraw = deferred()
    state.route.params = { code: 'token-A-123' }
    state.get.mockResolvedValueOnce(progress({ isMember: true, participantCount: 2 }))
    state.withdraw.mockReturnValueOnce(pendingWithdraw.promise)
    const wrapper = mountJoin({ attachTo: document.body })
    await flushPromises()
    await wrapper.findAll('button').find(item => item.text() === 'Retirar participación').trigger('click')
    await flushPromises()
    wrapper.unmount()
    pendingWithdraw.resolve(progress({ isMember: false, participantCount: 1 }))
    await flushPromises()
    expect(document.body.textContent).not.toContain('Participación retirada.')
    expect(document.body.textContent).not.toContain('1 de 12 participantes confirmados')
  })

  it('ignora respuestas al desmontar y no conserva el código localmente', async () => {
    let resolveRequest
    state.get.mockReturnValue(new Promise(resolve => { resolveRequest = resolve }))
    const wrapper = mountJoin({ attachTo: document.body })
    await wrapper.get('#join-code').setValue('token-secreto')
    await submit(wrapper)
    wrapper.unmount()
    resolveRequest(progress())
    await flushPromises()
    expect(document.body.textContent).not.toContain('token-secreto')
    expect(source).not.toMatch(/localStorage|sessionStorage|router\.replace/)
  })
})
