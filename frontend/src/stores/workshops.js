import { defineStore } from 'pinia'

import { workshopsService } from '@/services/workshops.service'

const queryPromises = new WeakMap()

const getQueryPromises = (store) => {
  let promises = queryPromises.get(store)

  if (!promises) {
    promises = new Map()
    queryPromises.set(store, promises)
  }

  return promises
}

const getFriendlyWorkshopError = (error, fallback) => {
  if (!error.response) {
    return 'No se pudo conectar con el backend. Verifica que el servidor esté encendido.'
  }

  return error.response?.data?.error || fallback
}

export const useWorkshopsStore =
  defineStore('workshops', {
    state: () => ({
      workshops: [],
      myEnrollments: [],
      loading: false,
      status: 'idle',
      hasLoaded: false,
      requestId: 0,
      historyLoading: false,
      historyStatus: 'idle',
      historyHasLoaded: false,
      historyRequestId: 0,
      loadingError: null,
      historyLoadingError: null,
      actionError: null,
      actionSuccess: null,
      enrollingId: null,
      withdrawingId: null
    }),

    getters: {
      initialLoading: (state) => state.loading && !state.hasLoaded,
      refreshing: (state) => state.loading && state.hasLoaded,
      historyInitialLoading: (state) =>
        state.historyLoading && !state.historyHasLoaded,
      historyRefreshing: (state) =>
        state.historyLoading && state.historyHasLoaded
    },

    actions: {
      async fetchWorkshops(options = {}) {
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
            const workshops = await workshopsService.getAll()

            if (requestId !== this.requestId) {
              return workshops
            }

            this.workshops = workshops
            this.hasLoaded = true
            this.status = 'success'

            return workshops
          } catch (error) {
            if (requestId !== this.requestId) {
              return null
            }

            if (!this.hasLoaded) {
              this.workshops = []
            }

            this.status = 'error'
            this.loadingError = getFriendlyWorkshopError(
              error,
              'No se pudieron cargar los talleres.'
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

      async fetchMyEnrollments(options = {}) {
        const force = options?.force === true
        const promises = getQueryPromises(this)
        const activePromise = promises.get('history')

        if (activePromise && !force) {
          return activePromise
        }

        const requestId = ++this.historyRequestId
        this.historyLoading = true
        this.historyStatus = 'loading'
        this.historyLoadingError = null

        const fetchPromise = (async () => {
          try {
            const enrollments = await workshopsService.getMine()

            if (requestId !== this.historyRequestId) {
              return enrollments
            }

            this.myEnrollments = enrollments
            this.historyHasLoaded = true
            this.historyStatus = 'success'

            return enrollments
          } catch (error) {
            if (requestId !== this.historyRequestId) {
              return null
            }

            if (!this.historyHasLoaded) {
              this.myEnrollments = []
            }

            this.historyStatus = 'error'
            this.historyLoadingError = getFriendlyWorkshopError(
              error,
              'No se pudo cargar tu historial de talleres.'
            )

            return null
          } finally {
            if (requestId === this.historyRequestId) {
              this.historyLoading = false
            }

            if (promises.get('history') === fetchPromise) {
              promises.delete('history')
            }
          }
        })()

        promises.set('history', fetchPromise)

        return fetchPromise
      },

      async enroll(workshopId) {
        if (this.enrollingId !== null) {
          return null
        }

        this.enrollingId = workshopId
        this.actionError = null
        this.actionSuccess = null

        try {
          const updatedWorkshop =
            await workshopsService.enroll(workshopId)

          this.workshops = this.workshops.map((workshop) =>
            Number(workshop.id) === Number(updatedWorkshop.id)
              ? { ...workshop, ...updatedWorkshop }
              : workshop
          )

          this.actionSuccess =
            'Inscripción registrada correctamente.'

          if (this.historyHasLoaded) {
            await this.fetchMyEnrollments({ force: true })
          }

          return updatedWorkshop
        } catch (error) {
          const message =
            getFriendlyWorkshopError(
              error,
              'No se pudo registrar la inscripción.'
            )

          const payload = error.response?.data
          this.actionError = payload?.code === 'WORKSHOP_SCHEDULE_CONFLICT'
            ? {
                code: payload.code,
                message,
                conflict: payload.conflict
              }
            : { message }

          throw new Error(message)
        } finally {
          this.enrollingId = null
        }
      },

      async withdraw(workshopId) {
        if (this.withdrawingId !== null) {
          return null
        }

        this.withdrawingId = workshopId
        this.actionError = null
        this.actionSuccess = null

        try {
          const enrollmentChange =
            await workshopsService.withdraw(workshopId)

          const changedWorkshopId = Number(
            enrollmentChange.workshopId ??
            enrollmentChange.id ??
            workshopId
          )
          const currentWorkshop = this.workshops.find(
            (workshop) => Number(workshop.id) === changedWorkshopId
          )

          this.workshops = this.workshops.map((workshop) =>
            Number(workshop.id) === changedWorkshopId
              ? {
                  ...workshop,
                  isEnrolled: enrollmentChange.isEnrolled,
                  enrolledCount: enrollmentChange.enrolledCount
                }
              : workshop
          )

          if (enrollmentChange.changed === true) {
            this.myEnrollments = this.myEnrollments.map((enrollment) =>
              Number(enrollment.workshopId) === changedWorkshopId &&
              enrollment.status === 'CONFIRMED'
                ? { ...enrollment, status: 'CANCELLED' }
                : enrollment
            )
          }

          this.actionSuccess = enrollmentChange.changed === false
            ? 'Ya no estás inscrito en este taller.'
            : currentWorkshop?.title
              ? `Te desinscribiste de ${currentWorkshop.title}. Tu cupo quedó disponible.`
              : 'Te desinscribiste del taller. Tu cupo quedó disponible.'

          return enrollmentChange
        } catch (error) {
          const payload = error.response?.data
          const message = payload?.code === 'WORKSHOP_ENROLLMENT_CLOSED'
            ? 'Este taller ya no admite cambios en la inscripción.'
            : getFriendlyWorkshopError(
                error,
                'No pudimos cancelar tu inscripción. Intenta nuevamente.'
              )

          this.actionError = { message }
          throw new Error(message)
        } finally {
          this.withdrawingId = null
        }
      },

      clearMessages() {
        this.actionError = null
        this.actionSuccess = null
      }
    }
  })
