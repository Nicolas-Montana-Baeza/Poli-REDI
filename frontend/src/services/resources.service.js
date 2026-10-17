import api from './api'

export const resourcesService = {
  async getAll() {
    const response = await api.get('/resources')

    return response.data
  },

  async create(payload) {
    const response = await api.post('/admin/resources', payload)
    return response.data
  },

  async updateOperational(resourceId, changes) {
    const response = await api.patch(
      `/resources/${resourceId}`,
      changes
    )

    return response.data
  },

  async updateImage(resourceId, imageUrl) {
    const response = await api.patch(
      `/resources/${resourceId}/image`,
      { imageUrl }
    )

    return response.data
  }
}
