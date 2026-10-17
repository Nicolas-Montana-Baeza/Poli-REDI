import api from './api'

export const reservationPolicyService = {
  async getCurrent() {
    const response = await api.get('/reservation-policy/current')
    return response.data
  },

  async getHistory() {
    const response = await api.get('/admin/reservation-policies')
    return response.data
  },

  async publish(policy, idempotencyKey) {
    const response = await api.post(
      '/admin/reservation-policies',
      policy,
      {
        headers: {
          'Idempotency-Key': idempotencyKey
        }
      }
    )

    return response.data
  }
}
