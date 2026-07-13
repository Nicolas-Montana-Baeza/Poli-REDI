import { defineStore } from 'pinia'

import { resourcesService } from '../services/resources.service'

export const useResourcesStore = defineStore('resources', {
  state: () => ({
    resources: [],
    loading: false,
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
    }
  },

  actions: {
    async fetchResources() {
      this.loading = true
      this.error = null

      try {
        this.resources = await resourcesService.getAll()
      } catch {
        this.resources = []
        this.error = 'No se pudieron cargar los recursos'
      } finally {
        this.loading = false
      }
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
