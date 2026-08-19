import api from './api'

export const institutionalActivitiesService = {
  // La consulta se realiza por unidad porque la autorización institucional
  // depende de la relación MANAGER/admin sobre esa unidad concreta.
  async getByUnit(unitId) {
    const response = await api.get(
      `/institutional-units/${unitId}/activities`
    )

    return response.data
  },

  // El backend conserva la autoridad sobre horarios, recursos, permisos,
  // bloqueos y generación de scheduling conflicts.
  async create(payload) {
    const response = await api.post(
      '/institutional-activities',
      payload
    )

    return response.data
  }
}
