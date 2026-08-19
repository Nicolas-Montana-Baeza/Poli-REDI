import api from './api'

// ============================================================================
// PROGRAMACIÓN INSTITUCIONAL - UNIDADES
// ============================================================================
//
// Esta capa conoce únicamente el contrato HTTP.
//
// Las reglas de autorización permanecen en backend para evitar confiar en el
// navegador como fuente de seguridad.

export const institutionalUnitsService = {
  async getAll() {
    const response = await api.get(
      '/institutional-units'
    )

    return response.data
  },

  async create(payload) {
    const response = await api.post(
      '/admin/institutional-units',
      payload
    )

    return response.data
  },

  // Consulta las relaciones institucionales de una unidad concreta.
  //
  // El backend valida que el actor pueda gestionar esa unidad; conocer su ID
  // no entrega acceso por sí solo.
  async getMemberships(unitId) {
    const response = await api.get(
      `/institutional-units/${unitId}/memberships`
    )

    return response.data
  },

  // La asignación MEMBER/MANAGER es una operación administrativa en MVP2.
  // userId proviene de la lista institucional de usuarios del backend.
  async addMembership(
    unitId,
    payload
  ) {
    const response = await api.post(
      `/admin/institutional-units/${unitId}/memberships`,
      payload
    )

    return response.data
  }
}
