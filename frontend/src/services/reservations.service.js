import api from './api'

export const reservationsService = {
  async getAll() {
    const response = await api.get('/reservations')

    return response.data
  },

  async getMine() {
    const response = await api.get('/reservations/mine')

    return response.data
  },

  async getAvailability() {
    const response = await api.get('/availability/reservations')

    return response.data
  },

  async create(reservation) {
    const response = await api.post('/reservations', reservation)

    return response.data
  },
  async getCurrentPolicy() {
    const response = await api.get('/reservation-policy/current')
    return response.data
  },
  async updateTarget(reservationId, targetParticipants) {
    const response = await api.patch(`/reservations/${reservationId}/target-participants`, { targetParticipants })
    return response.data
  },

  async cancel(reservationId) {
    const response = await api.patch('/reservations/cancel', {
      reservationId
    })

    return response.data
  }
}
