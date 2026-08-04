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

  async getAvailability(range = {}) {
    const params = range?.from && range?.to
      ? { from: range.from, to: range.to }
      : undefined
    const response = await api.get('/availability/reservations', { params })

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
  async getGroupProgress(code) {
    return (await api.get(`/group-reservations/${encodeURIComponent(code)}`)).data
  },
  async confirmGroup(code) {
    return (await api.put(`/group-reservations/${encodeURIComponent(code)}/confirmation`)).data
  },
  async withdrawGroup(code) {
    return (await api.delete(`/group-reservations/${encodeURIComponent(code)}/confirmation`)).data
  },
  async getJoinCode(reservationId) {
    return (await api.get(`/reservations/${reservationId}/join-code`)).data
  },
  async rotateJoinCode(reservationId) {
    return (await api.post(`/reservations/${reservationId}/join-code/rotate`)).data
  },

  async cancel(reservationId) {
    const response = await api.patch('/reservations/cancel', {
      reservationId
    })

    return response.data
  }
}
