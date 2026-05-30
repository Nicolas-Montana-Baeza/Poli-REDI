import { defineStore } from 'pinia'
import api from '../services/api'
import { getCurrentAccount, logout } from '../auth/authService'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    account: null,
    user: null,
    loading: false,
    error: null
  }),

  getters: {
    isAuthenticated: (state) => !!state.account,
    isAdmin: (state) => state.user?.isAdmin === true,
    isBlocked: (state) => state.user?.isBlocked === true
  },

  actions: {
    async loadAuthUser() {
      this.loading = true
      this.error = null

      try {
        this.account = await getCurrentAccount()

        if (!this.account) {
          this.user = null
          return null
        }

        const response = await api.get('/me')
        this.user = response.data

        return this.user
      } catch (error) {
        this.error = error.response?.data?.error || error.message
        this.user = null
        return null
      } finally {
        this.loading = false
      }
    },

    async logoutUser() {
      this.account = null
      this.user = null
      await logout()
    }
  }
})