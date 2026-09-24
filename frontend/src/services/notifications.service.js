import api from './api'

export const notificationsService = {
  async getAll() {
    const response = await api.get('/notifications')

    return response.data
  },

  async markRead(notificationId) {
    const response = await api.patch(
      `/notifications/${notificationId}/read`
    )

    return response.data
  }
}
