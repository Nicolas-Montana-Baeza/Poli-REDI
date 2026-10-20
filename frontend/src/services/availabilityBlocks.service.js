import api from './api'

export const availabilityBlocksService = {
  async getAll({ from, to }) {
    const response = await api.get('/admin/availability-blocks', {
      params: { from, to }
    })
    return response.data
  },

  async create(payload) {
    const response = await api.post('/admin/availability-blocks', payload)
    return response.data
  },

  async deactivate(id) {
    const response = await api.patch(
      `/admin/availability-blocks/${id}`,
      { isActive: false }
    )
    return response.data
  }
}
