import api from './api'

// ============================================================================
// WORKSHOPS INSTITUCIONALES
// ============================================================================
//
// WORKSHOP es un subtipo de institutional_activity.
//
// El frontend no mantiene un segundo modelo de calendario o cupos:
// programación, recurso, unidad e inscripciones provienen directamente del
// módulo institucional del backend.

export const workshopsService = {
  async getAll() {
    const response = await api.get('/workshops')

    return response.data
  },

  async getById(workshopId) {
    const response = await api.get(
      `/workshops/${workshopId}`
    )

    return response.data
  },

  async enroll(workshopId) {
    const response = await api.post(
      `/workshops/${workshopId}/enroll`
    )

    return response.data
  },

  async leave(workshopId) {
    const response = await api.delete(
      `/workshops/${workshopId}/enroll`
    )

    return response.data
  }
}
