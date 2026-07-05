import api from './api'

export const activitiesService = {
  async getAll() {
    const response = await api.get('/activities')

    return response.data
  },

  async create(activity) {
    const response = await api.post('/activities', activity)

    return response.data
  }
}
