import { defineStore } from 'pinia'
import api from '../services/api'
import {
  clearDevRutResetFlag,
  getCurrentAccount,
  logout
} from '../auth/authService'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    account: null,
    user: null,
    loading: false,
    error: null,
    errorStatus: null
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
      this.errorStatus = null

      try {
        this.account = await getCurrentAccount()

        if (!this.account) {
          this.user = null
          return null
        }

        const response = await api.get('/me')
        this.user = response.data
        clearDevRutResetFlag()

        return this.user
      } catch (error) {
        this.error = error.response?.data?.error || error.message
        this.errorStatus = error.response?.status || null
        this.user = null
        return null
      } finally {
        this.loading = false
      }
    },

    async logoutUser() {
      this.account = null
      this.user = null
      this.error = null
      this.errorStatus = null
      await logout()
    },

    async updateRut(rut) {
      this.error = null

      try {
        const response = await api.patch('/me/rut', {
          rut
        })

        this.user = response.data

        return this.user
      } catch (error) {
        this.error =
          error.response?.data?.error ||
          'No se pudo actualizar el RUT'

        throw new Error(this.error)
      }
    }
  }
})
