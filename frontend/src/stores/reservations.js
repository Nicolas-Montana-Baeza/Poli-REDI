import { defineStore } from 'pinia'

import { reservationsService } from '../services/reservations.service'

export const useReservationsStore = defineStore('reservations', {
  state: () => ({
    reservations: [],
    loading: false,
    error: null
  }),

  actions: {
    async fetchReservations() {
      this.loading = true
      this.error = null

      try {
        this.reservations = await reservationsService.getAll()
      } catch (error) {
        this.error = 'No se pudieron cargar las reservas'

        console.error(
          'Error cargando reservas:',
          error
        )
      } finally {
        this.loading = false
      }
    },

    async createReservation(reservation) {
      this.loading = true
      this.error = null

      try {
        const createdReservation =
          await reservationsService.create(reservation)

        this.reservations.push(createdReservation)

        return createdReservation
      } catch (error) {
        const message =
          error.response?.data?.error ||
          'No se pudo crear la reserva'

        this.error = message

        console.error(
          'Error creando reserva:',
          error
        )

        throw new Error(message)
      } finally {
        this.loading = false
      }
    }
  }
})