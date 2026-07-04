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
        this.error = 'No se pudieron cargar las actividades'
      } finally {
        this.loading = false
      }
    }
  }
})
