import api from './api'

export const resourcesService = {
  async getAll() {
    try {
      const response = await api.get('/resources')
      return response.data
    } catch (error) {
      console.error(
        'Error en resources.service:',
        JSON.stringify(error.response?.data, null, 2)
      )

      throw error
    }
  }
}