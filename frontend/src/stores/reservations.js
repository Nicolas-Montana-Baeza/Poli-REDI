import { defineStore } from 'pinia'

import { reservationsService } from '@/services/reservations.service'

export const useReservationsStore =
  defineStore('reservations', {
    state: () => ({
      reservations: [],

      loading: false,

      loadingError: null,

      actionError: null,

      actionSuccess: null
    }),

    actions: {
      async fetchReservations() {
        this.loading = true
        this.loadingError = null

        try {
          this.reservations =
            await reservationsService.getAll()
        } catch {
          this.reservations = []
          this.loadingError =
            'No se pudieron cargar las reservas'
        } finally {
          this.loading = false
        }
      },

      async createReservation(reservation) {
        this.loading = true
        this.actionError = null
        this.actionSuccess = null

        try {
          const createdReservation =
            await reservationsService.create(
              reservation
            )

          this.reservations.push(
            createdReservation
          )

          return createdReservation
        } catch (error) {
          const message =
            error.response?.data?.error ||
            'No se pudo crear la reserva'

          this.actionError = message

          throw new Error(message)
        } finally {
          this.loading = false
        }
      },

      clearActionError() {
        this.actionError = null
      },

      clearActionSuccess() {
        this.actionSuccess = null
      },

      setActionError(message) {
        this.actionError = message
        this.actionSuccess = null
      },

      setActionSuccess(message) {
        this.actionError = null
        this.actionSuccess = message
      },

      clearLoadingError() {
        this.loadingError = null
      },

      async cancelReservation(id) {
        this.loading = true
        this.actionError = null
        this.actionSuccess = null

        try {
          const cancelledReservation =
            await reservationsService.cancel(id)

          this.reservations =
            this.reservations.map((reservation) =>
              reservation.id === id
                ? cancelledReservation
                : reservation
            )

          return cancelledReservation
        } catch (error) {
          const message =
            error.response?.data?.error ||
            'No se pudo cancelar la reserva'

          this.actionError = message

          throw new Error(message)
        } finally {
          this.loading = false
        }
      }
    }
  })
