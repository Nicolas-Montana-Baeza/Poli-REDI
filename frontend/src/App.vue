<script setup>
import { ref, computed, onMounted } from 'vue'
import Calendar from './components/Calendar.vue'
import SchedulePicker from './components/SchedulePicker.vue'
import NextReservations from './components/NextReservations.vue'
import ResourcePicker from './components/ResourcePicker.vue'
import SportPicker from './components/SportPicker.vue'
import QuickActions from './components/QuickActions.vue'
import Sidebar from './components/Sidebar.vue'
import { api } from './services/api.js'

const selectedDate = ref(null)
const selectedResource = ref(null)
const selectedSport = ref(null)
const selectedSchedule = ref(null)

const fullyBookedDates = ref([])
const resources = ref([])
const sports = ref([])
const reservations = ref([])
const bookingConfig = ref({})

// Fetch initial data from API
onMounted(async () => {
  try {
    fullyBookedDates.value = await api.getFullyBookedDates()
    resources.value = await api.getResources()
    sports.value = await api.getSports()
    reservations.value = await api.getNextReservations()
    bookingConfig.value = await api.getBookingConfig()
  } catch (error) {
    console.error('Error loading data:', error)
  }
})

const handleDateSelected = (dateObj) => {
  selectedDate.value = dateObj
}

const handleResourceSelected = (resource) => {
  selectedResource.value = resource
  selectedSport.value = null
}

const handleSportSelected = (sport) => {
  selectedSport.value = sport
}

const handleScheduleSelected = (schedule) => {
  selectedSchedule.value = schedule
}

const reservationPayload = computed(() => {
  if (
    !selectedDate.value ||
    !selectedResource.value ||
    !selectedSport.value ||
    !selectedSchedule.value
  ) return null

  return {
    resource_id: selectedResource.value.id,
    sport: selectedSport.value,
    start_time: `${selectedDate.value.fullDate}T${selectedSchedule.value.hour}:00`,
    duration: selectedSchedule.value.duration,
    user_email: 'nicolas@ucentral.cl'
  }
})

const submitReservation = () => {
  console.log('Payload para API:', reservationPayload.value)
}

const handleReserve = () => {
  console.log('Abrir formulario de reserva')
}

const handleMyReservations = () => {
  console.log('Ver mis reservaciones')
}

const handleAvailability = () => {
  console.log('Ver disponibilidad')
}

const handleCancel = () => {
  console.log('Cancelar acción')
}

const handleHistory = () => {
  console.log('Ver historial')
}

const handleContact = () => {
  console.log('Contactar soporte')
}

const handleGoStats = () => {
  console.log('Ver estadísticas')
}

const handleSettings = () => {
  console.log('Configuración')
}

const handleHelp = () => {
  console.log('Ayuda')
}

const handleLogout = () => {
  console.log('Cerrar sesión')
}

</script>

<template>
  <Calendar
    :fullyBookedDates="fullyBookedDates"
    :selectedDate="selectedDate"
    @date-selected="handleDateSelected"
  />

  <ResourcePicker
    :selectedResource="selectedResource"
    @resource-selected="handleResourceSelected"
  />

  <SportPicker
    :selectedSport="selectedSport"
    :selectedResource="selectedResource"
    @sport-selected="handleSportSelected"
  />

  <SchedulePicker
    :selectedDate="selectedDate"
    @schedule-selected="handleScheduleSelected"
  />

  <button
    v-if="reservationPayload"
    class="reserve-btn"
    @click="submitReservation"
  >
    Reservar
  </button>

  <QuickActions @reserve="handleReserve" @my-reservations="handleMyReservations" @availability="handleAvailability" @cancel="handleCancel" />

  <NextReservations :reservations="reservations" />
  <Sidebar @history="handleHistory" @contact="handleContact" @go-stats="handleGoStats" @settings="handleSettings" @help="handleHelp" @logout="handleLogout" />
</template>

<style scoped>
.reserve-btn {
  margin: 20px 0;
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: 12px;
  background: #4caf50;
  color: white;
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
}

.reserve-btn:hover {
  opacity: 0.9;
}
</style>