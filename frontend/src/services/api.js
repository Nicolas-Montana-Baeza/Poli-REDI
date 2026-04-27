// API Service - currently with hard-coded data, ready to switch to real API calls

export const api = {
  // Get booking configuration (hours, rules, etc)
  async getBookingConfig() {
    return {
      startHour: 8,
      endHour: 20,
      selectableDaysRange: 30,
    }
  },

  // Get fully booked dates
  async getFullyBookedDates() {
    return [
      '2026-04-28',
      '2026-04-30',
    ]
  },

  // Get booked time slots for a specific date
  async getBookedSlots(date) {
    // date format: '2026-04-29'
    return []
  },

  // Get available time slots for a specific date
  async getAvailableSlots(date) {
    // date format: '2026-04-29'
    // If fully booked, return empty array
    const fullyBooked = await this.getFullyBookedDates()
    if (fullyBooked.includes(date)) {
      return []
    }
    
    const config = await this.getBookingConfig()
    const slots = []
    for (let hour = config.startHour; hour < config.endHour; hour++) {
      slots.push(`${hour}-${hour + 1}`)
    }
    return slots
  },

  // Get next reservations for the user
  async getNextReservations() {
    return [
      {
        id: 1,
        place: 'Cancha 1',
        date: 'Jue 22 mayo',
        time: '17:00 - 18:30',
        sport: 'Fútbol',
        status: 'Confirmada',
      },
      {
        id: 2,
        place: 'Piscina',
        date: 'Vie 23 mayo',
        time: '10:00 - 11:00',
        sport: 'Natación',
        status: 'Confirmada',
      }
    ]
  },

  // Get available resources/courts
  async getResources() {
    return [
      {
        id: 1,
        name: 'Cancha 1',
        type: 'Canchas de Fútbol',
        availability: 'Disponible',
      },
      {
        id: 2,
        name: 'Cancha 2',
        type: 'Canchas de Fútbol',
        availability: 'Disponible',
      },
      {
        id: 3,
        name: 'Piscina',
        type: 'Piscina',
        availability: 'Disponible',
      }
    ]
  },

  // Get available sports
  async getSports() {
    return [
      { id: 1, name: 'Fútbol', icon: '⚽' },
      { id: 2, name: 'Tenis', icon: '🎾' },
      { id: 3, name: 'Natación', icon: '🏊' },
      { id: 4, name: 'Baloncesto', icon: '🏀' },
    ]
  },

  // Create a booking
  async createBooking(data) {
    // { date, time, resource, sport, userId }
    return {
      success: true,
      id: Math.random(),
      message: 'Reserva creada exitosamente',
    }
  },

  // Cancel a booking
  async cancelBooking(bookingId) {
    return {
      success: true,
      message: 'Reserva cancelada',
    }
  },

  // Get user profile
  async getUserProfile() {
    return {
      id: 1,
      name: 'Usuario',
      email: 'usuario@example.com',
      credits: 100,
    }
  },
}
