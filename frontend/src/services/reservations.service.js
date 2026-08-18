import api from './api'

export const reservationsService = {
  async getAll() {
    const response = await api.get('/reservations')

    return response.data
  },

  async getMine() {
    const response = await api.get('/reservations/mine')

    return response.data
  },

  async getAvailability({ from, to, signal } = {}) {
    const response = await api.get('/availability/reservations', {
      params: { from, to },
      signal
    })

    return response.data
  },

  async getById(id) {
    const response = await api.get(`/reservations/${id}`)

    return response.data
  },

  async create(reservation) {
    const response = await api.post('/reservations', reservation)

    return response.data
  },

  async cancel(reservationId) {
    const response = await api.patch('/reservations/cancel', {
      reservationId
    })

    return response.data
  },

  // ------------------------------------------------------------
  // Reservas grupales MVP2.
  // ------------------------------------------------------------

  // Obtiene el estado actual del grupo asociado a un join code.
  //
  // Esta operación no incorpora al usuario automáticamente. Permite mostrar
  // primero información como participantes actuales, mínimo requerido,
  // capacidad y condición del grupo antes de confirmar la participación.
  async getGroupProgress(joinCode) {
    const response = await api.get(
      `/reservations/join/${encodeURIComponent(joinCode)}`
    )

    return response.data
  },

  // Incorpora al usuario autenticado a la reserva grupal.
  //
  // El frontend nunca envía un userId: la API obtiene la identidad desde
  // la sesión autenticada y evita que un cliente pueda inscribir a terceros.
  async joinGroup(joinCode) {
    const response = await api.post(
      `/reservations/join/${encodeURIComponent(joinCode)}`
    )

    return response.data
  },

  // Retira al usuario autenticado del grupo.
  //
  // El backend mantiene la regla de negocio que impide al owner abandonar
  // la reserva mediante este flujo y conserva CONFIRMED + AT_RISK cuando
  // un grupo previamente confirmado cae bajo el mínimo.
  async leaveGroup(joinCode) {
    const response = await api.delete(
      `/reservations/join/${encodeURIComponent(joinCode)}`
    )

    return response.data
  },

  // Obtiene el listado completo de participantes de una reserva grupal.
  //
  // Esta operación está protegida en backend y solo puede ser utilizada
  // por el owner de la reserva o por un administrador.
  async getParticipants(reservationId) {
    const response = await api.get(
      `/reservations/${reservationId}/participants`
    )

    return response.data
  }
}
