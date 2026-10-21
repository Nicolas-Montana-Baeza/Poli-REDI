import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DOMWrapper, flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import InstitutionalUnitsView from './InstitutionalUnitsView.vue'

const mocks = vi.hoisted(() => ({
  store: null,
  getUsers: vi.fn()
}))

vi.mock('../stores/institutionalUnits', () => ({
  useInstitutionalUnitsStore: () => mocks.store
}))

vi.mock('../services/userService', () => ({
  usersService: {
    getAll: mocks.getUsers
  }
}))

const unit = () => ({
  id: 4,
  name: 'Escuela de Informática',
  code: 'INF',
  unitType: 'ACADEMIC_PROGRAM',
  isActive: true
})

const existingMembership = () => ({
  id: 30,
  userId: 20,
  userFullName: 'Ana Prueba',
  userEmail: 'ana@example.test',
  role: 'MEMBER',
  isActive: true
})

let wrapper

beforeEach(() => {
  mocks.getUsers.mockReset()

  mocks.getUsers.mockResolvedValue([
    { id: 20, fullName: 'Ana Prueba', email: 'ana@example.test', isBlocked: false },
    { id: 21, fullName: 'Beatriz Prueba', email: 'beatriz@example.test', isBlocked: false },
    { id: 22, fullName: 'Usuario Bloqueado', email: 'bloqueado@example.test', isBlocked: true }
  ])

  mocks.store = reactive({
    units: [
      unit(),
      {
        id: 5,
        name: 'Unidad inactiva',
        code: 'OLD',
        unitType: 'OTHER',
        isActive: false
      }
    ],
    get activeUnits() {
      return this.units.filter(
        unit => unit.isActive === true
      )
    },
    membershipsByUnit: {},
    loading: false,
    creating: false,
    loadingMemberships: false,
    assigningMembership: false,
    error: null,

    clearError: vi.fn(() => {
      mocks.store.error = null
    }),

    loadUnits: vi.fn().mockResolvedValue(),

    loadMemberships: vi.fn(async id => {
      mocks.store.membershipsByUnit[id] = [
        existingMembership()
      ]
    }),

    createUnit: vi.fn(async payload => ({
      id: 40,
      ...payload,
      isActive: true
    })),

    addMembership: vi.fn(async (_id, payload) => ({
      id: 90,
      userId: payload.userId,
      userFullName: 'Beatriz Prueba',
      role: payload.role,
      isActive: true
    }))
  })
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ''
})

const mountView = async () => {
  wrapper = mount(InstitutionalUnitsView, {
    attachTo: document.body
  })
  await flushPromises()
}

const modal = selector => {
  const element = document.querySelector(selector)
  expect(element).not.toBeNull()
  return new DOMWrapper(element)
}

const openCreate = async () => {
  await wrapper.get('.page-header .primary-button').trigger('click')
  return modal('.modal-card:not(.memberships-modal)')
}

const openMembers = async () => {
  await wrapper.get('.unit-card .secondary-button').trigger('click')
  await flushPromises()
  return modal('.memberships-modal')
}

describe('ADMIN-005: unidades institucionales', () => {
  it('carga unidades y excluye las inactivas', async () => {
    await mountView()

    expect(mocks.store.loadUnits).toHaveBeenCalledOnce()
    expect(wrapper.findAll('.unit-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('Escuela de Informática')
    expect(wrapper.text()).not.toContain('Unidad inactiva')
  })

  it('filtra unidades por código institucional', async () => {
    mocks.store.units.push({
      id: 6,
      name: 'Unidad Deportiva',
      code: 'DEP',
      unitType: 'SPORTS_UNIT',
      isActive: true
    })

    await mountView()
    await wrapper.get('.search-field input').setValue('INF')

    expect(wrapper.findAll('.unit-card')).toHaveLength(1)
    expect(wrapper.text()).toContain('Escuela de Informática')
    expect(wrapper.text()).not.toContain('Unidad Deportiva')
  })

  it('crea una unidad normalizando el código', async () => {
    await mountView()

    const form = await openCreate()
    const inputs = form.findAll('input')

    await inputs[0].setValue('Nueva Escuela')
    await inputs[1].setValue('nueva')

    await form.get('form.unit-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.createUnit).toHaveBeenCalledWith({
      name: 'Nueva Escuela',
      code: 'NUEVA',
      unitType: 'ACADEMIC_PROGRAM'
    })

    expect(document.querySelector('#create-unit-title')).toBeNull()
    expect(wrapper.text()).toContain('creada correctamente')
  })

  it('conserva el formulario cuando falla la creación', async () => {
    mocks.store.createUnit = vi.fn(async () => {
      mocks.store.error = 'El código institucional ya existe.'
      throw new Error('Conflicto')
    })

    await mountView()
    const form = await openCreate()

    await form.findAll('input')[0].setValue('Escuela duplicada')
    await form.findAll('input')[1].setValue('INF')

    await form.get('form.unit-form').trigger('submit')
    await flushPromises()

    expect(document.querySelector('#create-unit-title')).not.toBeNull()
    expect(modal('.modal-card').text()).toContain(
      'El código institucional ya existe.'
    )
    expect(wrapper.find('.success-message').exists()).toBe(false)
  })

  it('consulta miembros y excluye usuarios bloqueados o asignados', async () => {
    await mountView()

    const form = await openMembers()

    expect(mocks.store.loadMemberships).toHaveBeenCalledWith(4)
    expect(mocks.getUsers).toHaveBeenCalledOnce()

    const options = form.findAll('.unit-form select')[0]
      .findAll('option')
      .map(option => option.text())
      .join(' ')

    expect(options).toContain('Beatriz Prueba')
    expect(options).not.toContain('Ana Prueba')
    expect(options).not.toContain('Usuario Bloqueado')
  })

  it('asigna MANAGER a un usuario disponible', async () => {
    await mountView()

    const form = await openMembers()
    const selects = form.findAll('.unit-form select')

    await selects[0].setValue('21')
    await selects[1].setValue('MANAGER')

    await form.get('form.unit-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.addMembership).toHaveBeenCalledWith(
      4,
      { userId: 21, role: 'MANAGER' }
    )

    expect(wrapper.text()).toContain('fue asignado como MANAGER')
  })

  it('no asigna miembros sin seleccionar usuario', async () => {
    await mountView()

    const form = await openMembers()
    await form.get('form.unit-form').trigger('submit')
    await flushPromises()

    expect(mocks.store.addMembership).not.toHaveBeenCalled()
    expect(form.text()).toContain('Debes seleccionar un usuario')
  })
})
