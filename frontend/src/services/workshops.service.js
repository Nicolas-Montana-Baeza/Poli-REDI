import api from './api'

export const workshopsService = {
  async getAll() {
    const response = await api.get('/workshops')

    return response.data
  },

  async getMine() {
    const response = await api.get('/workshop-enrollments/mine')

    return response.data
  },

  async enroll(workshopId) {
    const response = await api.post(
      `/workshops/${workshopId}/enroll`
    )

    return response.data
  }
}
