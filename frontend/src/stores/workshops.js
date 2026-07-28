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
      myEnrollments: [],
      loading: false,
      historyLoading: false,
      loadingError: null,
      historyLoadingError: null,
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

      async fetchMyEnrollments() {
        this.historyLoading = true
        this.historyLoadingError = null

        try {
          this.myEnrollments = await workshopsService.getMine()
        } catch (error) {
          this.myEnrollments = []
          this.historyLoadingError = getFriendlyWorkshopError(
            error,
            'No se pudo cargar tu historial de talleres.'
          )
        } finally {
          this.historyLoading = false
        }
      },

      async enroll(workshopId) {
        if (this.enrollingId !== null) {
          return null
        }

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

          const payload = error.response?.data
          this.actionError = payload?.code === 'WORKSHOP_SCHEDULE_CONFLICT'
            ? {
                code: payload.code,
                message,
                conflict: payload.conflict
              }
            : { message }

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
