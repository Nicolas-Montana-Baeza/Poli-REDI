import { defineStore } from 'pinia'

import { resourcesService } from '../services/resources.service'

export const useResourcesStore = defineStore('resources', {
  state: () => ({
    resources: [],
    loading: false,
    error: null
  }),

  getters: {
    availableResources: (state) => {
      return state.resources.filter(
        resource => resource.status === 'available'
      )
    }
  },

  actions: {
    async fetchResources() {
      this.loading = true
      this.error = null

      try {
        this.resources = await resourcesService.getAll()
      } catch (error) {
        this.error = 'No se pudieron cargar los recursos'

        console.error(
          'Error cargando recursos:',
          error
        )
      } finally {
        this.loading = false
      }
    }
  }
})