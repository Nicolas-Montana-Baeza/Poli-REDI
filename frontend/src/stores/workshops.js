import { defineStore } from 'pinia'

import { workshopsService } from '@/services/workshops.service'

const getFriendlyWorkshopError = (error, fallback) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend. Verifica que el servidor esté encendido.'
  }

  return error.response?.data?.error || fallback
}

export const useWorkshopsStore =
  defineStore('workshops', {
    state: () => ({
      workshops: [],
      loading: false,
      loadingError: null,
      actionError: null,
      actionSuccess: null,
      enrollingId: null
    }),

    actions: {
      async fetchWorkshops() {
        this.loading = true
        this.loadingError = null

        try {
          this.workshops = await workshopsService.getAll()
        } catch (error) {
          this.workshops = []
          this.loadingError = getFriendlyWorkshopError(
            error,
            'No se pudieron cargar los talleres.'
          )
        } finally {
          this.loading = false
        }
      },

      async enroll(workshopId) {
        this.enrollingId = workshopId
        this.actionError = null
        this.actionSuccess = null

        try {
          const updatedWorkshop =
            await workshopsService.enroll(workshopId)

          this.workshops = this.workshops.map((workshop) =>
            workshop.id === updatedWorkshop.id
              ? updatedWorkshop
              : workshop
          )

          this.actionSuccess =
            'Inscripción registrada correctamente.'

          return updatedWorkshop
        } catch (error) {
          const message =
            getFriendlyWorkshopError(
              error,
              'No se pudo registrar la inscripción.'
            )

          this.actionError = message

          throw new Error(message)
        } finally {
          this.enrollingId = null
        }
      },

      clearMessages() {
        this.actionError = null
        this.actionSuccess = null
      }
    }
  })
