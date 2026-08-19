import { defineStore } from 'pinia'

import { workshopsService } from '@/services/workshops.service'

// ============================================================================
// ERRORES
// ============================================================================

const getFriendlyWorkshopError = (
  error,
  fallback
) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend. Verifica que el servidor esté encendido.'
  }

  return (
    error.response?.data?.error ||
    fallback
  )
}

// ============================================================================
// STORE
// ============================================================================

export const useWorkshopsStore =
  defineStore('workshops', {
    state: () => ({
      workshops: [],

      loading: false,
      loadingError: null,

      actionError: null,
      actionSuccess: null,

      // Solo una operación por tarjeta debe mostrarse como pendiente.
      //
      // El backend sigue siendo responsable de proteger la concurrencia real
      // y la capacidad del taller.
      mutatingId: null,
      mutationType: null
    }),

    actions: {
      // ----------------------------------------------------------------------
      // CONSULTA
      // ----------------------------------------------------------------------

      async fetchWorkshops() {
        this.loading = true
        this.loadingError = null

        try {
          this.workshops =
            await workshopsService.getAll()
        } catch (error) {
          this.workshops = []

          this.loadingError =
            getFriendlyWorkshopError(
              error,
              'No se pudieron cargar los talleres.'
            )
        } finally {
          this.loading = false
        }
      },

      // ----------------------------------------------------------------------
      // ACTUALIZACIÓN LOCAL
      // ----------------------------------------------------------------------
      //
      // Los endpoints de enroll/leave devuelven el workshop ya enriquecido.
      // Reemplazamos únicamente esa tarjeta sin volver a cargar todo el listado.

      replaceWorkshop(updatedWorkshop) {
        const index =
          this.workshops.findIndex(
            workshop =>
              workshop.id ===
              updatedWorkshop.id
          )

        if (index === -1) {
          this.workshops.push(
            updatedWorkshop
          )

          return
        }

        this.workshops.splice(
          index,
          1,
          updatedWorkshop
        )
      },

      // ----------------------------------------------------------------------
      // INSCRIPCIÓN
      // ----------------------------------------------------------------------

      async enroll(workshopId) {
        this.mutatingId = workshopId
        this.mutationType = 'ENROLL'

        this.actionError = null
        this.actionSuccess = null

        try {
          const updatedWorkshop =
            await workshopsService.enroll(
              workshopId
            )

          this.replaceWorkshop(
            updatedWorkshop
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
          this.mutatingId = null
          this.mutationType = null
        }
      },

      // ----------------------------------------------------------------------
      // RETIRO
      // ----------------------------------------------------------------------

      async leave(workshopId) {
        this.mutatingId = workshopId
        this.mutationType = 'LEAVE'

        this.actionError = null
        this.actionSuccess = null

        try {
          const updatedWorkshop =
            await workshopsService.leave(
              workshopId
            )

          this.replaceWorkshop(
            updatedWorkshop
          )

          this.actionSuccess =
            'Tu inscripción fue cancelada correctamente.'

          return updatedWorkshop
        } catch (error) {
          const message =
            getFriendlyWorkshopError(
              error,
              'No se pudo cancelar la inscripción.'
            )

          this.actionError = message

          throw new Error(message)
        } finally {
          this.mutatingId = null
          this.mutationType = null
        }
      },

      clearMessages() {
        this.actionError = null
        this.actionSuccess = null
      }
    }
  })
