import { defineStore } from 'pinia'

import { resourcesService } from '../services/resources.service'

const fetchPromises = new WeakMap()

export const useResourcesStore = defineStore('resources', {
  state: () => ({
    resources: [],
    loading: false,
    status: 'idle',
    hasLoaded: false,
    requestId: 0,
    error: null,
    actionError: null,
    actionSuccess: null,
    updatingImageId: null
  }),

  getters: {
    availableResources: (state) => {
      return state.resources.filter(
        resource => resource.status === 'available'
      )
    },
    initialLoading: (state) => state.loading && !state.hasLoaded,
    refreshing: (state) => state.loading && state.hasLoaded
  },

  actions: {
    async fetchResources(options = {}) {
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
          const resources = await resourcesService.getAll()

          if (requestId !== this.requestId) {
            return resources
          }

          this.resources = resources
          this.hasLoaded = true
          this.status = 'success'

          return resources
        } catch {
          if (requestId !== this.requestId) {
            return null
          }

          if (!this.hasLoaded) {
            this.resources = []
          }

          this.status = 'error'
          this.error = 'No se pudieron cargar los recursos'

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
    },

    async updateImage(resourceId, imageUrl) {
      this.updatingImageId = resourceId
      this.actionError = null
      this.actionSuccess = null

      try {
        const updatedResource =
          await resourcesService.updateImage(resourceId, imageUrl)

        this.resources = this.resources.map((resource) =>
          resource.id === updatedResource.id
            ? updatedResource
            : resource
        )

        this.actionSuccess = 'Imagen actualizada correctamente'

        return updatedResource
      } catch (error) {
        this.actionError =
          error.response?.data?.error ||
          'No se pudo actualizar la imagen'

        throw new Error(this.actionError)
      } finally {
        this.updatingImageId = null
      }
    }
  }
})
