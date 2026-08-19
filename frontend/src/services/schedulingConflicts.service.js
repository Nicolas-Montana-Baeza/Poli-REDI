import api from './api'

// ============================================================================
// CONFLICTOS DE PROGRAMACIÓN
// ============================================================================
//
// La UI solamente expresa decisiones administrativas.
//
// La validación del plan resultante, locks, incompatibilidades, excepciones
// por ocurrencia y cierre del conflicto continúan siendo responsabilidad
// exclusiva del backend.

export const schedulingConflictsService = {
  async getAll(status = 'PENDING') {
    const response = await api.get(
      '/admin/scheduling-conflicts',
      {
        params: { status }
      }
    )

    return response.data
  },

  async getById(conflictId) {
    const response = await api.get(
      `/admin/scheduling-conflicts/${conflictId}`
    )

    return response.data
  },

  async resolveItem(
    conflictId,
    itemId,
    decision
  ) {
    const response = await api.patch(
      `/admin/scheduling-conflicts/${conflictId}/items/${itemId}`,
      decision
    )

    return response.data
  }
}
