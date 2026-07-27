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

const getAccountKey = (account) =>
  account?.homeAccountId ||
  account?.localAccountId ||
  account?.username ||
  null

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
      const account = await getCurrentAccount()
      const accountKey = getAccountKey(account)

      if (
        accountKey &&
        this.profileReady &&
        this._loadedAccountKey === accountKey
      ) {
        return this.user
      }
      if (
        this._loadPromise &&
        this._loadPromiseAccountKey === accountKey
      ) {
        return this._loadPromise
      }

      const generation = ++this.requestGeneration
      this.account = account
      this.loading = true
      this.profileReady = false
      this.error = null
      this.errorStatus = null
      this._loadPromiseAccountKey = accountKey
      const loadPromise = (async () => {
        try {
          if (!account) {
            if (generation !== this.requestGeneration) return null
            this.user = null
            this._loadedAccountKey = null
            return null
          }

          const response = await api.get('/me')
          if (generation !== this.requestGeneration) return null
          this.user = response.data
          this._loadedAccountKey = accountKey
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
          this._loadedAccountKey = null
          return null
        } finally {
          if (generation === this.requestGeneration) {
            this.loading = false
          }
        }
      })()
      this._loadPromise = loadPromise

      try {
        return await loadPromise
      } finally {
        if (generation === this.requestGeneration) {
          this._loadPromise = null
          this._loadPromiseAccountKey = null
        }
      }
    },

    async logoutUser() {
      this.requestGeneration += 1
      this._loadPromise = null
      this._loadPromiseAccountKey = null
      this._loadedAccountKey = null
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
        const response = await api.patch('/me/rut', { rut })
        if (generation !== this.requestGeneration) return null
        this.user = response.data
        this.profileReady = true
        return this.user
      } catch (error) {
        if (generation !== this.requestGeneration) return null
        const message = getFriendlyApiError(
          error,
          'No se pudo actualizar el RUT.'
        )
        this.error = message
        if (error.response?.status === 409) {
          this.profileReady = false
          this._loadedAccountKey = null
          await this.loadAuthUser()
        }
        throw new Error(message)
      }
    }
  }
})
