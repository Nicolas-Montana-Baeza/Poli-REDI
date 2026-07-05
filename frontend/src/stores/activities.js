import { defineStore } from 'pinia'

import { activitiesService } from '@/services/activities.service'

export const useActivitiesStore = defineStore('activities', {
  state: () => ({
    activities: [],
    loading: false,
    error: null
  }),

  actions: {
    async fetchActivities() {
      this.loading = true
      this.error = null

      try {
        this.activities = await activitiesService.getAll()
      } catch {
        this.activities = []
        this.error = 'No se pudieron cargar las actividades'
      } finally {
        this.loading = false
      }
    },

    async createActivity(activity) {
      this.loading = true
      this.error = null

      try {
        const createdActivity = await activitiesService.create(activity)
        const exists = this.activities.some(
          (item) => item.id === createdActivity.id
        )

        if (!exists) {
          this.activities.push(createdActivity)
          this.activities.sort((a, b) => a.name.localeCompare(b.name))
        }

        return createdActivity
      } catch (error) {
        const message =
          error.response?.data?.error ||
          error.response?.data?.detail ||
          'No se pudo crear la actividad'

        this.error = message
        throw new Error(message)
      } finally {
        this.loading = false
      }
    }
  }
})
