import { defineStore }
from 'pinia'

export const useReservationsStore =
  defineStore('reservations', {

    state: () => ({

      reservations: [
        {
          id: 1,

          resourceId: 1,

          hour: '18:00',

          title: 'Entrenamiento',

          type: 'normal'
        },

        {
          id: 2,

          resourceId: 2,

          hour: '20:00',

          title: 'Campeonato',

          type: 'priority'
        },

        {
          id: 3,

          resourceId: 4,

          hour: '17:00',

          title: 'Mantención',

          type: 'normal'
        }
      ]
    }),

    getters: {

      getByResource:
        (state) =>
        (resourceId) => {

          return state.reservations.filter(
            r =>
              r.resourceId ===
              resourceId
          )
        }
    },

    actions: {

      addReservation(
        reservation
      ) {

        this.reservations.push({
          id: Date.now(),

          ...reservation
        })
      },

      removeReservation(id) {

        this.reservations =
          this.reservations.filter(
            r => r.id !== id
          )
      }
    }
  })