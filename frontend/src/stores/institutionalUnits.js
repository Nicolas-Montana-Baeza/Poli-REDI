import { defineStore } from 'pinia'

import {
  institutionalUnitsService
} from '../services/institutionalUnits.service'

const getFriendlyError = (
  error,
  fallback
) => {
  if (!error.response) {
    return (
      'No se pudo conectar con el backend. ' +
      'Verifica que el servidor esté encendido.'
    )
  }

  return (
    error.response?.data?.error ||
    fallback
  )
}

// ============================================================================
// STORE DE UNIDADES INSTITUCIONALES
// ============================================================================
//
// La lista almacenada aquí es la fuente de verdad de la pantalla.
//
// Cuando se crea una unidad correctamente se incorpora directamente al estado.
// Esto evita realizar inmediatamente un segundo GET solo para actualizar la
// interfaz.

export const useInstitutionalUnitsStore = defineStore(
  'institutionalUnits',
  {
    state: () => ({
      units: [],
      loading: false,
      creating: false,
      error: null
    }),

    getters: {
      activeUnits: (state) => (
        state.units.filter(
          (unit) => unit.isActive === true
        )
      )
    },

    actions: {
      clearError() {
        this.error = null
      },

      async loadUnits() {
        this.loading = true
        this.error = null

        try {
          const units =
            await institutionalUnitsService.getAll()

          this.units = Array.isArray(units)
            ? units
            : []

          return this.units
        } catch (error) {
          this.error = getFriendlyError(
            error,
            'No se pudieron cargar las unidades institucionales.'
          )

          throw new Error(this.error)
        } finally {
          this.loading = false
        }
      },

      async createUnit(payload) {
        this.creating = true
        this.error = null

        try {
          const unit =
            await institutionalUnitsService.create(
              payload
            )

          this.units.push(unit)

          this.units.sort(
            (a, b) => (
              a.name.localeCompare(
                b.name,
                'es'
              )
            )
          )

          return unit
        } catch (error) {
          this.error = getFriendlyError(
            error,
            'No se pudo crear la unidad institucional.'
          )

          throw new Error(this.error)
        } finally {
          this.creating = false
        }
      }
    }
  }
)
