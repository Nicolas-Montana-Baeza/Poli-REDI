import { defineStore } from 'pinia'

import {
  schedulingConflictsService
} from '@/services/schedulingConflicts.service'

const friendlyError = (
  error,
  fallback
) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend.'
  }

  return (
    error.response?.data?.error ||
    fallback
  )
}

export const useSchedulingConflictsStore =
  defineStore('schedulingConflicts', {
    state: () => ({
      conflicts: [],
      selectedConflict: null,

      statusFilter: 'PENDING',

      loading: false,
      loadingDetail: false,

      resolvingItemId: null,

      error: null,
      actionError: null,
      actionSuccess: null
    }),

    actions: {
      async fetchConflicts(
        status = this.statusFilter
      ) {
        this.loading = true
        this.error = null
        this.statusFilter = status

        try {
          this.conflicts =
            await schedulingConflictsService
              .getAll(status)
        } catch (error) {
          this.conflicts = []

          this.error = friendlyError(
            error,
            'No se pudieron cargar los conflictos.'
          )
        } finally {
          this.loading = false
        }
      },

      async fetchConflict(conflictId) {
        this.loadingDetail = true
        this.actionError = null

        try {
          this.selectedConflict =
            await schedulingConflictsService
              .getById(conflictId)

          return this.selectedConflict
        } catch (error) {
          const message = friendlyError(
            error,
            'No se pudo cargar el conflicto.'
          )

          this.actionError = message

          throw new Error(message)
        } finally {
          this.loadingDetail = false
        }
      },

      async resolveItem(
        conflictId,
        itemId,
        decision
      ) {
        this.resolvingItemId = itemId
        this.actionError = null
        this.actionSuccess = null

        try {
          const updatedConflict =
            await schedulingConflictsService
              .resolveItem(
                conflictId,
                itemId,
                decision
              )

          this.selectedConflict =
            updatedConflict

          // El PATCH devuelve el snapshot actualizado del conflicto.
          // También sincronizamos la tarjeta del listado para evitar
          // una recarga completa innecesaria.
          const index =
            this.conflicts.findIndex(
              conflict =>
                conflict.id ===
                updatedConflict.id
            )

          if (index !== -1) {
            this.conflicts.splice(
              index,
              1,
              updatedConflict
            )
          }

          this.actionSuccess =
            updatedConflict.status === 'RESOLVED'
              ? 'Conflicto resuelto completamente.'
              : 'Decisión administrativa registrada.'

          return updatedConflict
        } catch (error) {
          const message = friendlyError(
            error,
            'No se pudo registrar la decisión.'
          )

          this.actionError = message

          throw new Error(message)
        } finally {
          this.resolvingItemId = null
        }
      },

      clearMessages() {
        this.actionError = null
        this.actionSuccess = null
      }
    }
  })
