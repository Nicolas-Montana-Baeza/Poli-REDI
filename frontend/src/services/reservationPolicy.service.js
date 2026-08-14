import api from './api'

export const reservationPolicyService = {
  async getCurrent() {
    const response = await api.get('/reservation-policy/current')
    return response.data
  }
}
