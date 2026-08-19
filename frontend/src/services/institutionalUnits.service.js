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
  }
}
