import api from './api'

export const resourcesService = {
  async getAll() {
    const response = await api.get('/resources')

    return response.data
  }
}
