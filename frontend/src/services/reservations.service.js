import api from './api'

export const reservationsService = {
  async getAll() {
    try {
      const response = await api.get('/reservations')
      return response.data
    } catch (error) {
      console.error('Error en reservations.service getAll:', error.response?.data || error)
      throw error
    }
  },

  async create(reservation) {
    try {
      const response = await api.post('/reservations', reservation)
      return response.data
    } catch (error) {
      console.error('Error en reservations.service create:', error.response?.data || error)
      throw error
    }
  },

  async cancel(reservationId) {
    try {
      const response = await api.patch('/reservations/cancel', {
        reservationId,
        requestedByUserId: 1
      })

      return response.data
    } catch (error) {
      console.error('Error en reservations.service cancel:', error.response?.data || error)
      throw error
    }
  }
}