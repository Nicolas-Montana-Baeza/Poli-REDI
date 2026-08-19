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

      // Las membresías se mantienen separadas por unidad para no mezclar
      // relaciones institucionales al navegar entre distintas estructuras.
      membershipsByUnit: {},

      loading: false,
      creating: false,
      loadingMemberships: false,
      assigningMembership: false,
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

      async loadMemberships(unitId) {
        this.loadingMemberships = true
        this.error = null

        try {
          const memberships =
            await institutionalUnitsService.getMemberships(
              unitId
            )

          this.membershipsByUnit[unitId] =
            Array.isArray(memberships)
              ? memberships
              : []

          return this.membershipsByUnit[unitId]
        } catch (error) {
          this.error = getFriendlyError(
            error,
            'No se pudieron cargar los miembros de la unidad.'
          )

          throw new Error(this.error)
        } finally {
          this.loadingMemberships = false
        }
      },

      async addMembership(
        unitId,
        payload
      ) {
        this.assigningMembership = true
        this.error = null

        try {
          const membership =
            await institutionalUnitsService.addMembership(
              unitId,
              payload
            )

          const current = [
            ...(this.membershipsByUnit[unitId] || [])
          ]

          // El backend puede crear o reactivar una relación existente.
          // Por eso reemplazamos por ID/userId en vez de agregar duplicados
          // visuales a la lista.
          const index = current.findIndex(
            (item) => (
              item.id === membership.id ||
              item.userId === membership.userId
            )
          )

          if (index >= 0) {
            current.splice(
              index,
              1,
              membership
            )
          } else {
            current.push(membership)
          }

          current.sort(
            (a, b) => (
              String(a.userFullName || a.userEmail)
                .localeCompare(
                  String(b.userFullName || b.userEmail),
                  'es'
                )
            )
          )

          this.membershipsByUnit[unitId] =
            current

          return membership
        } catch (error) {
          this.error = getFriendlyError(
            error,
            'No se pudo asignar el usuario a la unidad.'
          )

          throw new Error(this.error)
        } finally {
          this.assigningMembership = false
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
