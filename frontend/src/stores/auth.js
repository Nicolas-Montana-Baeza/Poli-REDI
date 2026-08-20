import { defineStore } from 'pinia'
import api from '../services/api'
import {
  clearDevRutResetFlag,
  getCurrentAccount,
  logout
} from '../auth/authService'

let authInitializationPromise = null

const LOGOUT_PENDING_KEY =
  'poli_redi_logout_pending'

const hasPendingLogout = () => {
  if (typeof sessionStorage === 'undefined') {
    return false
  }

  return (
    sessionStorage.getItem(
      LOGOUT_PENDING_KEY
    ) === 'true'
  )
}

const markLogoutPending = () => {
  if (typeof sessionStorage === 'undefined') {
    return
  }

  sessionStorage.setItem(
    LOGOUT_PENDING_KEY,
    'true'
  )
}

const clearLogoutPending = () => {
  if (typeof sessionStorage === 'undefined') {
    return
  }

  sessionStorage.removeItem(
    LOGOUT_PENDING_KEY
  )
}

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

    // false = todavía no sabemos si existe una sesión.
    // true = la sesión ya fue resuelta.
    initialized: false,

    // Mantiene visible la transición global mientras
    // se destruye la sesión local / Microsoft.
    loggingOut: hasPendingLogout(),

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
    async initializeSession() {
      if (this.initialized) {
        return this.user
      }

      if (!authInitializationPromise) {
        authInitializationPromise =
          this.loadAuthUser()
            .finally(() => {
              authInitializationPromise = null
            })
      }

      return authInitializationPromise
    },

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
        this.error = getFriendlyApiError(
          error,
          'No se pudo cargar tu sesión.'
        )
        this.errorStatus = error.response?.status || null
        this.user = null
        return null
      } finally {
        this.loading = false
        this.initialized = true
      }
    },

    async logoutUser() {
      if (this.loggingOut) {
        return
      }

      markLogoutPending()
      this.loggingOut = true

      this.account = null
      this.user = null
      this.initialized = true
      this.error = null
      this.errorStatus = null

      try {
        await logout()
      } catch (error) {
        clearLogoutPending()
        this.loggingOut = false

        this.error =
          error?.message ||
          'No se pudo cerrar la sesión.'

        throw error
      }
    },

    finishLogout() {
      clearLogoutPending()

      this.loggingOut = false

      // Después de cerrar sesión ya conocemos
      // definitivamente el estado de autenticación:
      // no existe una sesión activa.
      this.account = null
      this.user = null
      this.initialized = true

      this.loading = false
      this.error = null
      this.errorStatus = null
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
        this.error = getFriendlyApiError(
          error,
          'No se pudo actualizar el RUT.'
        )

        throw new Error(this.error)
      }
    }
  }
})
