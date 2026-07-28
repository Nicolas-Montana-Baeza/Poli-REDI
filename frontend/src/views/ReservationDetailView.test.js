// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({
  route: null,
  reservations: null,
  getJoinCode: vi.fn(),
  rotateJoinCode: vi.fn(),
  clipboard: vi.fn()
}))

vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  state.route = reactive({ params: { id: '1' }, query: {} })
  return { useRoute: () => state.route, useRouter: () => ({ push: vi.fn() }) }
})
vi.mock('@/stores/auth', async () => {
  const { reactive } = await import('vue')
  const store = reactive({ user: { id: 7, isAdmin: false }, loadAuthUser: vi.fn(async () => store.user) })
  return { useAuthStore: () => store }
})
vi.mock('@/stores/reservations', async () => {
  const { reactive } = await import('vue')
  state.reservations = reactive({
    myReservations: [],
    reservations: [],
    myLoading: false,
    loading: false,
    myLoadingError: '',
    loadingError: '',
    actionError: '',
    actionSuccess: '',
    fetchMyReservations: vi.fn(),
    fetchReservations: vi.fn(),
    cancelReservation: vi.fn(),
    setActionSuccess: vi.fn()
  })
  return { useReservationsStore: () => state.reservations }
})
vi.mock('@/services/reservations.service', () => ({
  reservationsService: {
    getJoinCode: state.getJoinCode,
    rotateJoinCode: state.rotateJoinCode,
    updateTarget: vi.fn()
  }
}))

import ConfirmModal from '@/components/ui/ConfirmModal.vue'
import ReservationDetailView from './ReservationDetailView.vue'

const reservation = id => ({
  id, userId: 7, targetParticipants: 12, participantCount: 1,
  minimumParticipants: 10, capacity: 20, status: 'PENDING',
  confirmationDeadline: '2099-01-01T00:00:00Z', startTime: '2098-12-01T10:00:00-03:00',
  durationMinutes: 60, resourceName: `Cancha ${id}`, title: `Reserva ${id}`
})
const button = (wrapper, text) => wrapper.findAll('button').find(item => item.text() === text)
const deferred = () => {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

describe('ReservationDetailView: código de invitación', () => {
  beforeEach(() => {
    state.route.params.id = '1'
  })
  afterEach(() => {
    document.body.innerHTML = ''
    Object.defineProperty(navigator, 'share', { configurable: true, value: undefined })
    vi.clearAllMocks()
  })

  it('muestra la gestión del código al propietario aunque los identificadores tengan distinto tipo', async () => {
    const ownGroup = {
      ...reservation(1),
      userId: '7',
      targetParticipants: undefined,
      capacity: 20
    }
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, ownGroup)
    state.getJoinCode.mockResolvedValueOnce({ joinCode: 'codigo-visible' })
    const wrapper = mount(ReservationDetailView)
    await flushPromises()

    expect(button(wrapper, 'Código de invitación')).toBeTruthy()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('codigo-visible')
    wrapper.unmount()
  })

  it('no ofrece el código a terceros ni en reservas canceladas', async () => {
    state.reservations.myReservations.splice(
      0,
      state.reservations.myReservations.length,
      { ...reservation(1), userId: 99 }
    )
    const thirdPartyWrapper = mount(ReservationDetailView)
    await flushPromises()
    expect(button(thirdPartyWrapper, 'Código de invitación')).toBeUndefined()
    thirdPartyWrapper.unmount()

    state.reservations.myReservations.splice(
      0,
      state.reservations.myReservations.length,
      { ...reservation(1), status: 'CANCELLED' }
    )
    const cancelledWrapper = mount(ReservationDetailView)
    await flushPromises()
    expect(button(cancelledWrapper, 'Código de invitación')).toBeUndefined()
    cancelledWrapper.unmount()
    expect(state.getJoinCode).not.toHaveBeenCalled()
  })

  it('permite generar código para una reserva legacy y marca la confirmación como destructiva', async () => {
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1))
    state.getJoinCode.mockRejectedValueOnce({ response: { status: 404 } })
    state.rotateJoinCode.mockResolvedValueOnce({ joinCode: 'nuevo código' })
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Esta reserva todavía no tiene un código de invitación.')
    await button(wrapper, 'Generar código').trigger('click')
    const modal = wrapper.findComponent(ConfirmModal)
    expect(modal.props('destructive')).toBe(true)
    expect(document.activeElement.textContent).toContain('Conservar código actual')
    await document.querySelector('.modal-card .btn-danger').click()
    await flushPromises()
    expect(state.rotateJoinCode).toHaveBeenCalledWith(1)
    expect(wrapper.text()).toContain('nuevo código')
    expect(wrapper.text()).toContain('El código anterior dejó de funcionar.')
    wrapper.unmount()
  })

  it('copia el código sin persistirlo y limpia estado al cambiar de ruta', async () => {
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText: state.clipboard.mockResolvedValue() } })
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1), reservation(2))
    state.getJoinCode.mockResolvedValueOnce({ joinCode: 'a/b c' })
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    await button(wrapper, 'Copiar código').trigger('click')
    await flushPromises()
    expect(state.clipboard).toHaveBeenCalledWith('a/b c')
    expect(wrapper.text()).toContain('Código copiado al portapapeles.')
    state.route.params.id = '2'
    await flushPromises()
    expect(wrapper.text()).not.toContain('a/b c')
    expect(wrapper.text()).not.toContain('Código copiado al portapapeles.')
    expect(button(wrapper, 'Código de invitación').attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })

  it('permite ajustar el objetivo con controles accesibles sin exponer el RUT', async () => {
    state.reservations.myReservations.splice(
      0,
      state.reservations.myReservations.length,
      { ...reservation(1), canEditTarget: true, userRut: '12.345.678-5' }
    )
    const wrapper = mount(ReservationDetailView)
    await flushPromises()

    const target = wrapper.get('#owner-target')
    expect(target.element.value).toBe('12')
    await wrapper.get('[aria-label="Aumentar objetivo"]').trigger('click')
    expect(target.element.value).toBe('13')
    await wrapper.get('[aria-label="Disminuir objetivo"]').trigger('click')
    expect(target.element.value).toBe('12')
    expect(wrapper.text()).not.toContain('12.345.678-5')
    expect(wrapper.text()).toContain('Responsable')
    expect(wrapper.text()).toContain('Participantes')
    expect(wrapper.text()).toContain('1 de 12 participantes confirmados')
    expect(wrapper.text()).toContain('La reserva se confirmará automáticamente al llegar a 10 participantes.')
    expect(wrapper.text()).toContain('Mínimo 10')
    expect(wrapper.text()).toContain('Capacidad 20')
    expect(wrapper.text()).toContain('Objetivo de participantes')
    expect(wrapper.text()).toContain('Puedes cambiarlo hasta una hora antes')
    expect(button(wrapper, 'Guardar cambios')).toBeTruthy()
    expect(button(wrapper, 'Cancelar reserva')).toBeTruthy()
    expect(wrapper.text()).not.toContain('Progreso de participantes')
    expect(wrapper.text()).not.toContain('Guardar objetivo')
    expect(wrapper.text()).not.toContain('Define cuántas personas esperas reunir.')
    wrapper.unmount()
  })

  it('comparte con Web Share y conserva el copiado de enlace como alternativa', async () => {
    const share = vi.fn().mockResolvedValue()
    Object.defineProperty(navigator, 'share', { configurable: true, value: share })
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText: state.clipboard.mockResolvedValue() } })
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1))
    state.getJoinCode.mockResolvedValueOnce({ joinCode: 'codigo seguro' })
    const wrapper = mount(ReservationDetailView)
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()

    await button(wrapper, 'Compartir invitación').trigger('click')
    expect(share).toHaveBeenCalledWith({
      title: 'Invitación a reserva grupal',
      url: `${window.location.origin}/join/codigo%20seguro`
    })
    expect(wrapper.text()).toContain('Invitación compartida.')

    expect(wrapper.text()).toContain('Invita a tu grupo')
    expect(button(wrapper, 'Copiar código')).toBeTruthy()
    wrapper.unmount()
  })

  it('trata una respuesta vacía como ausencia esperada y conserva errores generales', async () => {
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1))
    state.getJoinCode.mockResolvedValueOnce({ joinCode: '' })
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Esta reserva todavía no tiene un código de invitación.')
    expect(wrapper.text()).not.toContain('No se pudo recuperar el código.')
    await button(wrapper, 'Código de invitación').trigger('click')
    state.getJoinCode.mockRejectedValueOnce({ response: { status: 500 } })
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('No se pudo recuperar el código.')
    wrapper.unmount()
  })

  it('descarta un getJoinCode tardío al cambiar a otra reserva', async () => {
    const pendingGet = deferred()
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1), reservation(2))
    state.getJoinCode.mockReturnValueOnce(pendingGet.promise)
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    state.route.params.id = '2'
    await flushPromises()
    pendingGet.resolve({ joinCode: 'secreto antiguo' })
    await flushPromises()
    expect(wrapper.text()).not.toContain('secreto antiguo')
    expect(wrapper.text()).not.toContain('No se pudo recuperar el código.')
    expect(button(wrapper, 'Código de invitación').attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })

  it('descarta una rotación tardía después de desmontar', async () => {
    const pendingRotate = deferred()
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1))
    state.getJoinCode.mockResolvedValueOnce({ joinCode: 'código vigente' })
    state.rotateJoinCode.mockReturnValueOnce(pendingRotate.promise)
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    await button(wrapper, 'Generar código nuevo').trigger('click')
    document.querySelector('.modal-card .btn-danger').click()
    await flushPromises()
    wrapper.unmount()
    pendingRotate.resolve({ joinCode: 'secreto rotado tardío' })
    await flushPromises()
    expect(document.body.textContent).not.toContain('secreto rotado tardío')
    expect(document.body.textContent).not.toContain('El código anterior dejó de funcionar.')
  })

  it('una rotación tardía no reabre el panel de una nueva reserva', async () => {
    const pendingRotate = deferred()
    state.reservations.myReservations.splice(0, state.reservations.myReservations.length, reservation(1), reservation(2))
    state.getJoinCode.mockResolvedValueOnce({ joinCode: 'código vigente' })
    state.rotateJoinCode.mockReturnValueOnce(pendingRotate.promise)
    const wrapper = mount(ReservationDetailView, { attachTo: document.body })
    await flushPromises()
    await button(wrapper, 'Código de invitación').trigger('click')
    await flushPromises()
    await button(wrapper, 'Generar código nuevo').trigger('click')
    document.querySelector('.modal-card .btn-danger').click()
    await flushPromises()
    state.route.params.id = '2'
    await flushPromises()
    pendingRotate.resolve({ joinCode: 'secreto de reserva uno' })
    await flushPromises()
    expect(wrapper.text()).not.toContain('secreto de reserva uno')
    expect(wrapper.text()).not.toContain('El código anterior dejó de funcionar.')
    expect(button(wrapper, 'Código de invitación').attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })
})
