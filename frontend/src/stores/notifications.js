import { defineStore } from 'pinia'

import { getCurrentAccount } from '@/auth/authService'
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
      const account = await getCurrentAccount()

      if (!account) {
        this.clearNotifications()
        return
      }

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
    },

    clearNotifications() {
      this.notifications = []
      this.error = null
      this.loading = false
    }
  }
})
