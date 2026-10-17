import { defineStore } from 'pinia'
import { venuesService } from '../services/venues.service'

export const useVenuesStore = defineStore('venues', {
  state: () => ({
    venues: [],
    loading: false,
    saving: false,
    error: null,
    actionError: null,
    actionSuccess: null
  }),

  actions: {
    async fetchVenues() {
      this.loading = true
      this.error = null

      try {
        this.venues = await venuesService.getAll()
        return this.venues
      } catch (error) {
        this.error =
          error.response?.data?.error ||
          'No se pudieron cargar las sedes'
        throw new Error(this.error)
      } finally {
        this.loading = false
      }
    },

    async createVenue(payload) {
      this.saving = true
      this.actionError = null
      this.actionSuccess = null

      try {
        const venue = await venuesService.create(payload)

        this.venues = [...this.venues, venue]
          .sort((a, b) => a.id - b.id)

        this.actionSuccess = 'Sede creada correctamente'
        return venue
      } catch (error) {
        this.actionError =
          error.response?.data?.error ||
          'No se pudo crear la sede'
        throw new Error(this.actionError)
      } finally {
        this.saving = false
      }
    },

    async updateVenue(venueId, changes) {
      this.saving = true
      this.actionError = null
      this.actionSuccess = null

      try {
        const venue = await venuesService.update(venueId, changes)

        this.venues = this.venues.map((item) =>
          item.id === venue.id ? venue : item
        )

        this.actionSuccess = 'Sede actualizada correctamente'
        return venue
      } catch (error) {
        this.actionError =
          error.response?.data?.error ||
          'No se pudo actualizar la sede'
        throw new Error(this.actionError)
      } finally {
        this.saving = false
      }
    }
  }
})
