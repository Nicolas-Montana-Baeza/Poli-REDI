import { defineStore }
from 'pinia'

export const useResourcesStore =
  defineStore('resources', {

    state: () => ({

      resources: [
        {
          id: 1,

          name: 'Cancha 1',

          type: 'Exterior',

          status: 'available'
        },

        {
          id: 2,

          name: 'Cancha 2',

          type: 'Exterior',

          status: 'busy'
        },

        {
          id: 3,

          name: 'Cancha 3',

          type: 'Exterior',

          status: 'available'
        },

        {
          id: 4,

          name: 'Gimnasio',

          type: 'Interior',

          status: 'maintenance'
        },

        {
          id: 5,

          name: 'Piscina',

          type: 'Interior',

          status: 'available'
        },
        {
          id: 6,
          name: 'Sala Multiuso',
          type: 'Interior',
          status: 'available'
        }
      ]
    }),

    getters: {

      availableResources:
        (state) => {

          return state.resources.filter(
            r =>
              r.status ===
              'available'
          )
        }
    },

    actions: {

      addResource(resource) {

        this.resources.push({
          id: Date.now(),

          ...resource
        })
      },

      updateStatus(
        id,
        status
      ) {

        const resource =
          this.resources.find(
            r => r.id === id
          )

        if (resource) {
          resource.status =
            status
        }
      }
    }
  })