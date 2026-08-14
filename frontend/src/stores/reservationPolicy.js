import { defineStore } from 'pinia'

import { reservationPolicyService } from '@/services/reservationPolicy.service'
import { normalizeReservationPolicy } from '@/utils/reservationRules'

export const useReservationPolicyStore = defineStore('reservationPolicy', {
  state: () => ({
    policy: null,
    loading: false,
    loaded: false,
    error: null
  }),

  actions: {
    async fetchCurrentPolicy() {
      this.loading = true
      this.error = null

      try {
        const policy = normalizeReservationPolicy(
          await reservationPolicyService.getCurrent()
        )

        if (!policy) {
          throw new Error('La política vigente no es válida.')
        }

        this.policy = policy
        this.loaded = true
        return policy
      } catch (error) {
        this.policy = null
        this.loaded = true
        this.error = error.response?.data?.error ||
          error.message ||
          'No se pudo cargar la política vigente.'
        return null
      } finally {
        this.loading = false
      }
    }
  }
})
