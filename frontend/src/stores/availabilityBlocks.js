import { defineStore } from 'pinia'
import { availabilityBlocksService } from '../services/availabilityBlocks.service'

const apiError = (error, fallback) => error?.response?.data?.error || fallback

export const useAvailabilityBlocksStore = defineStore('availabilityBlocks', {
  state: () => ({
    blocks: [],
    loading: false,
    saving: false,
    updatingId: null,
    error: null,
    actionError: null,
    actionSuccess: null
  }),

  actions: {
    async fetchBlocks(range) {
      this.loading = true
      this.error = null
      try {
        const items = await availabilityBlocksService.getAll(range)
        this.blocks = Array.isArray(items) ? items : []
        return this.blocks
      } catch (error) {
        this.error = apiError(error, 'No se pudieron cargar los bloqueos.')
        throw new Error(this.error)
      } finally {
        this.loading = false
      }
    },

    async createBlock(payload) {
      this.saving = true
      this.actionError = null
      this.actionSuccess = null
      try {
        const created = await availabilityBlocksService.create(payload)
        this.blocks = [...this.blocks.filter(item => item.id !== created.id), created]
        this.actionSuccess = 'Bloqueo creado correctamente.'
        return created
      } catch (error) {
        this.actionError = apiError(error, 'No se pudo crear el bloqueo.')
        throw new Error(this.actionError)
      } finally {
        this.saving = false
      }
    },

    async deactivateBlock(id) {
      this.updatingId = id
      this.actionError = null
      this.actionSuccess = null
      try {
        const updated = await availabilityBlocksService.deactivate(id)
        this.blocks = this.blocks.map(item => item.id === id ? updated : item)
        this.actionSuccess = 'Bloqueo desactivado. El registro se conserva en el historial.'
        return updated
      } catch (error) {
        this.actionError = apiError(error, 'No se pudo desactivar el bloqueo.')
        throw new Error(this.actionError)
      } finally {
        this.updatingId = null
      }
    }
  }
})
