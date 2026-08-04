import { defineStore } from 'pinia'

import { reservationsService } from '@/services/reservations.service'

const queryPromises = new WeakMap()

const getQueryPromises = (store) => {
  let promises = queryPromises.get(store)

  if (!promises) {
    promises = new Map()
    queryPromises.set(store, promises)
  }

  return promises
}

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

      status: 'idle',

      hasLoaded: false,

      requestId: 0,

      availabilityLoading: false,

      availabilityLoadingError: null,

      availabilityStatus: 'idle',

      availabilityHasLoaded: false,

      availabilityRequestId: 0,

      myLoading: false,

      myLoadingError: null,

      myStatus: 'idle',

      myHasLoaded: false,

      myRequestId: 0,

      creating: false,

      cancellingId: null,

      actionError: null,

      actionSuccess: null
    }),

    getters: {
      initialLoading: (state) => state.loading && !state.hasLoaded,
      refreshing: (state) => state.loading && state.hasLoaded,
      availabilityInitialLoading: (state) =>
        state.availabilityLoading && !state.availabilityHasLoaded,
      availabilityRefreshing: (state) =>
        state.availabilityLoading && state.availabilityHasLoaded,
      myInitialLoading: (state) => state.myLoading && !state.myHasLoaded,
      myRefreshing: (state) => state.myLoading && state.myHasLoaded
    },

    actions: {
      async fetchReservations(options = {}) {
        const force = options?.force === true
        const promises = getQueryPromises(this)
        const activePromise = promises.get('all')

        if (activePromise && !force) {
          return activePromise
        }

        const requestId = ++this.requestId
        this.loading = true
        this.status = 'loading'
        this.loadingError = null

        const fetchPromise = (async () => {
          try {
            const reservations = await reservationsService.getAll()

            if (requestId !== this.requestId) {
              return reservations
            }

            this.reservations = reservations
            this.hasLoaded = true
            this.status = 'success'

            return reservations
          } catch (error) {
            if (requestId !== this.requestId) {
              return null
            }

            if (!this.hasLoaded) {
              this.reservations = []
            }

            this.status = 'error'
            this.loadingError = getFriendlyReservationError(
              error,
              'No se pudieron cargar las reservas.'
            )

            return null
          } finally {
            if (requestId === this.requestId) {
              this.loading = false
            }

            if (promises.get('all') === fetchPromise) {
              promises.delete('all')
            }
          }
        })()

        promises.set('all', fetchPromise)

        return fetchPromise
      },

      async fetchAvailabilityReservations(options = {}) {
        const force = options?.force === true
        const promises = getQueryPromises(this)
        const activePromise = promises.get('availability')

        if (activePromise && !force) {
          return activePromise
        }

        const requestId = ++this.availabilityRequestId
        this.availabilityLoading = true
        this.availabilityStatus = 'loading'
        this.availabilityLoadingError = null

        const fetchPromise = (async () => {
          try {
            const reservations = await reservationsService.getAvailability()

            if (requestId !== this.availabilityRequestId) {
              return reservations
            }

            this.availabilityReservations = reservations
            this.availabilityHasLoaded = true
            this.availabilityStatus = 'success'

            return reservations
          } catch (error) {
            if (requestId !== this.availabilityRequestId) {
              return null
            }

            if (!this.availabilityHasLoaded) {
              this.availabilityReservations = []
            }

            this.availabilityStatus = 'error'
            this.availabilityLoadingError = getFriendlyReservationError(
              error,
              'No se pudo validar la disponibilidad actual.'
            )

            return null
          } finally {
            if (requestId === this.availabilityRequestId) {
              this.availabilityLoading = false
            }

            if (promises.get('availability') === fetchPromise) {
              promises.delete('availability')
            }
          }
        })()

        promises.set('availability', fetchPromise)

        return fetchPromise
      },

      async fetchMyReservations(options = {}) {
        const force = options?.force === true
        const promises = getQueryPromises(this)
        const activePromise = promises.get('mine')

        if (activePromise && !force) {
          return activePromise
        }

        const requestId = ++this.myRequestId
        this.myLoading = true
        this.myStatus = 'loading'
        this.myLoadingError = null

        const fetchPromise = (async () => {
          try {
            const reservations = await reservationsService.getMine()

            if (requestId !== this.myRequestId) {
              return reservations
            }

            this.myReservations = reservations
            this.myHasLoaded = true
            this.myStatus = 'success'

            return reservations
          } catch (error) {
            if (requestId !== this.myRequestId) {
              return null
            }

            if (!this.myHasLoaded) {
              this.myReservations = []
            }

            this.myStatus = 'error'
            this.myLoadingError = getFriendlyReservationError(
              error,
              'No se pudieron cargar tus reservas.'
            )

            return null
          } finally {
            if (requestId === this.myRequestId) {
              this.myLoading = false
            }

            if (promises.get('mine') === fetchPromise) {
              promises.delete('mine')
            }
          }
        })()

        promises.set('mine', fetchPromise)

        return fetchPromise
      },

      async createReservation(reservation) {
        this.creating = true
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
          this.creating = false
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
        this.cancellingId = id
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
          if (this.cancellingId === id) {
            this.cancellingId = null
          }
        }
      }
    }
  })
