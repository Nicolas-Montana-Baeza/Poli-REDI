import { defineStore } from 'pinia'

import { resourcesService } from '../services/resources.service'

export const useResourcesStore = defineStore('resources', {
  state: () => ({
    resources: [],
    loading: false,
    error: null,
    actionError: null,
    actionSuccess: null,
    creatingResource: false,
    updatingResourceId: null,
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

    async createResource(payload) {
      this.creatingResource = true
      this.actionError = null
      this.actionSuccess = null

      try {
        const resource = await resourcesService.create(payload)

        this.resources = [...this.resources, resource]
          .sort((a, b) => a.id - b.id)

        this.actionSuccess = 'Recurso creado correctamente'
        return resource
      } catch (error) {
        this.actionError =
          error.response?.data?.error ||
          'No se pudo crear el recurso'

        throw new Error(this.actionError)
      } finally {
        this.creatingResource = false
      }
    },

    async updateOperational(resourceId, changes) {
      this.updatingResourceId = resourceId
      this.actionError = null
      this.actionSuccess = null

      try {
        const updatedResource =
          await resourcesService.updateOperational(
            resourceId,
            changes
          )

        this.resources = this.resources.map((resource) =>
          resource.id === updatedResource.id
            ? updatedResource
            : resource
        )

        this.actionSuccess =
          'Configuración actualizada correctamente'

        return updatedResource
      } catch (error) {
        this.actionError =
          error.response?.data?.error ||
          'No se pudo actualizar el recurso'

        throw new Error(this.actionError)
      } finally {
        this.updatingResourceId = null
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
