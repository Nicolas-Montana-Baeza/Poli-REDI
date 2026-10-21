import { defineStore } from 'pinia'

import {
  institutionalActivitiesService
} from '../services/institutionalActivities.service'


const friendlyError = (
  error,
  fallback
) => {
  return (
    error?.response?.data?.error ||
    error?.message ||
    fallback
  )
}


export const useInstitutionalActivitiesStore =
  defineStore(
    'institutionalActivities',
    {
      state: () => ({
        activitiesByUnit: {},
        loading: false,
        creating: false,
        saving: false,
        cancellingId: null,
        error: null
      }),

      actions: {
        clearError() {
          this.error = null
        },

        async loadByUnit(unitId) {
          this.loading = true
          this.error = null

          try {
            const activities =
              await institutionalActivitiesService
                .getByUnit(unitId)

            this.activitiesByUnit[unitId] =
              Array.isArray(activities)
                ? activities
                : []

            return this.activitiesByUnit[unitId]
          } catch (error) {
            this.error = friendlyError(
              error,
              'No se pudo cargar la programación institucional.'
            )

            throw error
          } finally {
            this.loading = false
          }
        },

        async update(id, payload) {
          this.saving = true
          this.error = null
          try {
            const updated = await institutionalActivitiesService.update(id, payload)
            const items = this.activitiesByUnit[payload.unitId] || []
            this.activitiesByUnit[payload.unitId] = items.map(item =>
              item.id === id ? updated : item
            )
            return updated
          } catch (error) {
            this.error = friendlyError(error, 'No se pudo editar la actividad.')
            throw error
          } finally {
            this.saving = false
          }
        },

        async cancel(id, unitId) {
          this.cancellingId = id
          this.error = null
          try {
            const cancelled = await institutionalActivitiesService.cancel(id)
            const items = this.activitiesByUnit[unitId] || []
            this.activitiesByUnit[unitId] = items.map(item =>
              item.id === id ? cancelled : item
            )
            return cancelled
          } catch (error) {
            this.error = friendlyError(error, 'No se pudo cancelar la actividad.')
            throw error
          } finally {
            this.cancellingId = null
          }
        },

        async create(payload) {
          this.creating = true
          this.error = null

          try {
            const activity =
              await institutionalActivitiesService
                .create(payload)

            const current = [
              ...(this.activitiesByUnit[
                payload.unitId
              ] || [])
            ]

            current.unshift(activity)

            this.activitiesByUnit[
              payload.unitId
            ] = current

            return activity
          } catch (error) {
            this.error = friendlyError(
              error,
              'No se pudo crear la actividad institucional.'
            )

            throw error
          } finally {
            this.creating = false
          }
        }
      }
    }
  )
