import api from './api'

export const usersService = {
  async getAll() {
    const response = await api.get('/users')
    return response.data
  },
  async changeAccess(userId, payload) {
    const response = await api.patch(`/admin/users/${userId}/access`, payload)
    return response.data
  },
  async getAccessHistory(userId) {
    const response = await api.get(`/admin/users/${userId}/access-history`)
    return response.data
  },
  async changeRole(userId, payload) {
    const response = await api.patch(`/admin/users/${userId}/role`, payload)
    return response.data
  },
  async getRoleHistory(userId) {
    const response = await api.get(`/admin/users/${userId}/role-history`)
    return response.data
  }
}
