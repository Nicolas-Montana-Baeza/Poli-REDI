import api from './api'

export const venuesService = {
  async getAll() {
    const response = await api.get('/admin/venues')
    return response.data
  },

  async create(payload) {
    const response = await api.post('/admin/venues', payload)
    return response.data
  },

  async update(venueId, changes) {
    const response = await api.patch(
      `/admin/venues/${venueId}`,
      changes
    )
    return response.data
  }
}
