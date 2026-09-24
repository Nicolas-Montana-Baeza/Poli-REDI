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

    async markRead(notificationId) {
      try {
        const updated =
          await notificationsService.markRead(notificationId)

        const index = this.notifications.findIndex(
          notification => notification.id === notificationId
        )

        if (index !== -1) {
          this.notifications[index] = updated
        }

        return updated
      } catch {
        this.error = 'No se pudo actualizar la notificación'
        throw new Error(this.error)
      }
    },

    clearNotifications() {
      this.notifications = []
      this.error = null
      this.loading = false
    }
  }
})
