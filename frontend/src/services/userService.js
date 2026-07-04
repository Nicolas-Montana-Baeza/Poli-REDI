import api from './api'

export const usersService = {
  async getAll() {
    const response = await api.get('/users')

    return response.data
  }
}
