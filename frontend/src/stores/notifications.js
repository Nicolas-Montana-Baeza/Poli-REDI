import { defineStore } from 'pinia'

import { notificationsService } from '@/services/notifications.service'

export const useNotificationsStore = defineStore('notifications', {
  state: () => ({
    notifications: [],
    loading: false,
    error: null
  }),

  getters: {
    unreadCount: (state) => {
      return state.notifications.filter(
        notification => !notification.isRead
      ).length
    }
  },

  actions: {
    async fetchNotifications() {
      this.loading = true
      this.error = null

      try {
        this.notifications =
          await notificationsService.getAll()
      } catch {
        this.notifications = []
        this.error = 'No se pudieron cargar las notificaciones'
      } finally {
        this.loading = false
      }
    }
  }
})
