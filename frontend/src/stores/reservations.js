import { defineStore } from 'pinia'

import { reservationsService } from '@/services/reservations.service'

let availabilityRequestId = 0
let availabilityController = null

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

      availabilityRangeKey: null,

      reservationDetail: null,

      myReservations: [],

      loading: false,

      loadingError: null,

      availabilityLoading: false,

      availabilityLoadingError: null,

      detailLoading: false,

      detailLoadingError: null,

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

      async fetchAvailabilityReservations({ from, to } = {}) {
        const requestId = ++availabilityRequestId
        availabilityController?.abort()
        availabilityController = new AbortController()
        this.availabilityLoading = true
        this.availabilityLoadingError = null
        this.availabilityRangeKey = null

        try {
          const reservations = await reservationsService.getAvailability({
            from,
            to,
            signal: availabilityController.signal
          })

          if (requestId !== availabilityRequestId) {
            return false
          }

          this.availabilityReservations = reservations
          this.availabilityRangeKey = from || null
          return true
        } catch (error) {
          if (requestId !== availabilityRequestId || error.code === 'ERR_CANCELED') {
            return false
          }

          this.availabilityReservations = []
          this.availabilityRangeKey = null
          this.availabilityRangeKey = from 
	    ? String(from).slice(0, 10) 
	    : null
          this.availabilityLoadingError = getFriendlyReservationError(
            error,
            'No se pudo validar la disponibilidad actual.'
          )
          return false
        } finally {
          if (requestId === availabilityRequestId) {
            this.availabilityLoading = false
          }
        }
      },

      async fetchReservationDetail(id) {
        this.detailLoading = true
        this.detailLoadingError = null
        this.reservationDetail = null

        try {
          this.reservationDetail = await reservationsService.getById(id)
          return this.reservationDetail
        } catch (error) {
          this.detailLoadingError = getFriendlyReservationError(
            error,
            'No se pudo cargar el detalle de la reserva.'
          )
          return null
        } finally {
          this.detailLoading = false
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

          if (this.reservationDetail?.id === id) {
            this.reservationDetail = cancelledReservation
          }

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
