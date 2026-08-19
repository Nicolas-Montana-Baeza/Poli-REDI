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
