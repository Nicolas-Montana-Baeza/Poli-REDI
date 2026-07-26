import { defineStore } from 'pinia'
import api from '../services/api'
import {
  getCurrentAccount,
  logout
} from '../auth/authService'

const getFriendlyApiError = (error, fallback) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend. Verifica que el servidor esté encendido.'
  }

  return error.response?.data?.error || fallback
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    account: null,
    user: null,
    loading: false,
    error: null,
    errorStatus: null,
    profileReady: false,
    requestGeneration: 0
  }),

  getters: {
    isAuthenticated: (state) => !!state.account,
    isAdmin: (state) => state.user?.isAdmin === true,
    isBlocked: (state) => state.user?.isBlocked === true
  },

  actions: {
    async loadAuthUser() {
      if (this._loadPromise) return this._loadPromise
      const generation = ++this.requestGeneration
      this.loading = true
      this.profileReady = false
      this.error = null
      this.errorStatus = null
      this._loadPromise = (async () => {
        try {
        this.account = await getCurrentAccount()

        if (!this.account) {
          if (generation !== this.requestGeneration) return null
          this.user = null
          return null
        }

        const response = await api.get('/me')
        if (generation !== this.requestGeneration) return null
        this.user = response.data
        this.profileReady = true

        return this.user
      } catch (error) {
        if (generation !== this.requestGeneration) return null
        this.error = getFriendlyApiError(
          error,
          'No se pudo cargar tu sesión.'
        )
        this.errorStatus = error.response?.status || null
        this.user = null
        return null
      } finally {
        if (generation === this.requestGeneration) this.loading = false
      }
      })()
      try {
        return await this._loadPromise
      } finally {
        if (generation === this.requestGeneration) this._loadPromise = null
      }
    },

    async logoutUser() {
      this.requestGeneration += 1
      this._loadPromise = null
      this.account = null
      this.user = null
      this.profileReady = false
      this.loading = false
      this.error = null
      this.errorStatus = null
      await logout()
    },

    async updateRut(rut) {
      const generation = this.requestGeneration
      this.error = null

      try {
        const response = await api.patch('/me/rut', {
          rut
        })

        if (generation !== this.requestGeneration) return null
        this.user = response.data
        this.profileReady = true

        return this.user
      } catch (error) {
        this.error = getFriendlyApiError(
          error,
          'No se pudo actualizar el RUT.'
        )
        if (error.response?.status === 409) {
          await this.loadAuthUser()
        }

        throw new Error(this.error)
      }
    }
  }
})
