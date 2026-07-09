import { defineStore } from 'pinia'

import { reservationsService } from '@/services/reservations.service'

const getFriendlyReservationError = (error, fallback) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend. Verifica que el servidor esté encendido.'
  }

  return error.response?.data?.error || fallback
}

export const useReservationsStore =
  defineStore('reservations', {
    state: () => ({
      reservations: [],

      availabilityReservations: [],

      myReservations: [],

      loading: false,

      loadingError: null,

      availabilityLoading: false,

      availabilityLoadingError: null,

      myLoading: false,

      myLoadingError: null,

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
        } catch (error) {
          this.reservations = []
          this.loadingError = getFriendlyReservationError(
            error,
            'No se pudieron cargar las reservas.'
          )
        } finally {
          this.loading = false
        }
      },

      async fetchAvailabilityReservations() {
        this.availabilityLoading = true
        this.availabilityLoadingError = null

        try {
          this.availabilityReservations =
            await reservationsService.getAvailability()
        } catch (error) {
          this.availabilityReservations = []
          this.availabilityLoadingError = getFriendlyReservationError(
            error,
            'No se pudo validar la disponibilidad actual.'
          )
        } finally {
          this.availabilityLoading = false
        }
      },

      async fetchMyReservations() {
        this.myLoading = true
        this.myLoadingError = null

        try {
          this.myReservations =
            await reservationsService.getMine()
        } catch (error) {
          this.myReservations = []
          this.myLoadingError = getFriendlyReservationError(
            error,
            'No se pudieron cargar tus reservas.'
          )
        } finally {
          this.myLoading = false
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

          this.availabilityReservations.push(
            createdReservation
          )

          this.myReservations.push(
            createdReservation
          )

          return createdReservation
        } catch (error) {
          const message =
            getFriendlyReservationError(
              error,
              'No se pudo crear la reserva.'
            )

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

          this.availabilityReservations =
            this.availabilityReservations.map((reservation) =>
              reservation.id === id
                ? cancelledReservation
                : reservation
            )

          this.myReservations =
            this.myReservations.map((reservation) =>
              reservation.id === id
                ? cancelledReservation
                : reservation
            )

          return cancelledReservation
        } catch (error) {
          const message =
            getFriendlyReservationError(
              error,
              'No se pudo cancelar la reserva.'
            )

          this.actionError = message

          throw new Error(message)
        } finally {
          this.loading = false
        }
      }
    }
  })
