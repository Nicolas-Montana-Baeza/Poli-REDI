import api from './api'

export const resourcesService = {
  async getAll() {
    const response = await api.get('/resources')

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
