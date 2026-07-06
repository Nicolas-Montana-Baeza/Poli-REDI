import { defineStore } from 'pinia'

import { activitiesService } from '@/services/activities.service'

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
    error: null
  }),

  actions: {
    async fetchActivities() {
      this.loading = true
      this.error = null

      try {
        this.activities = (await activitiesService.getAll())
          .filter((activity) => isValidActivityName(activity.name))
      } catch {
        this.activities = []
        this.error = 'No se pudieron cargar las actividades'
      } finally {
        this.loading = false
      }
    }
  }
})
