import { defineStore } from 'pinia'

import { activitiesService } from '@/services/activities.service'

const fetchPromises = new WeakMap()

const isValidActivityName = (name) => {
  const normalized = String(name || '').trim().toLowerCase()
  const words = normalized
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .split(/\s+/)
  const blockedTerms = [
    'puta',
    'puto',
    'putas',
    'putos',
    'mierda',
    'weon',
    'weona',
    'weones',
    'weonas',
    'wea',
    'weas',
    'ctm',
    'maricon',
    'marica',
    'pico',
    'pene',
    'sexo',
    'fuck',
    'shit',
    'bitch'
  ]

  return (
    /^[\p{L}\s]{3,120}$/u.test(normalized) &&
    !blockedTerms.some((term) => words.includes(term))
  )
}

export const useActivitiesStore = defineStore('activities', {
  state: () => ({
    activities: [],
    loading: false,
    status: 'idle',
    hasLoaded: false,
    requestId: 0,
    error: null
  }),

  getters: {
    initialLoading: (state) => state.loading && !state.hasLoaded,
    refreshing: (state) => state.loading && state.hasLoaded
  },

  actions: {
    async fetchActivities(options = {}) {
      const force = options?.force === true
      const activePromise = fetchPromises.get(this)

      if (activePromise && !force) {
        return activePromise
      }

      const requestId = ++this.requestId
      this.loading = true
      this.status = 'loading'
      this.error = null

      const fetchPromise = (async () => {
        try {
          const activities = (await activitiesService.getAll())
            .filter((activity) => isValidActivityName(activity.name))

          if (requestId !== this.requestId) {
            return activities
          }

          this.activities = activities
          this.hasLoaded = true
          this.status = 'success'

          return activities
        } catch {
          if (requestId !== this.requestId) {
            return null
          }

          if (!this.hasLoaded) {
            this.activities = []
          }

          this.status = 'error'
          this.error = 'No se pudieron cargar las actividades'

          return null
        } finally {
          if (requestId === this.requestId) {
            this.loading = false
          }

          if (fetchPromises.get(this) === fetchPromise) {
            fetchPromises.delete(this)
          }
        }
      })()

      fetchPromises.set(this, fetchPromise)

      return fetchPromise
    }
  }
})
