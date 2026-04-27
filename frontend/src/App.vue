<script setup>
import Calendar from './components/Calendar.vue'
import SchedulePicker from './components/SchedulePicker.vue'
import NextReservations from './components/NextReservations.vue'
import ResourcePicker from './components/ResourcePicker.vue'
import SportPicker from './components/SportPicker.vue'
import QuickActions from './components/QuickActions.vue'
import Sidebar from './components/Sidebar.vue'
import { ref, computed } from 'vue'

const selectedDate = ref(null)
const selectedResource = ref(null)
const selectedSport = ref(null)
const selectedSchedule = ref(null)

const fullyBookedDates = [
  '2026-04-28',
  '2026-04-30',
]

const reservations = ref([
  {
    id: 1,
    place: 'Cancha 1',
    date: 'Jue 22 mayo, 17:00 - 18:30',
    location: 'Fútbol',
    status: 'Confirmada',
  },
  {
    id: 2,
    place: 'Piscina',
    date: 'Vie 23 mayo, 10:00 - 11:00',
    location: 'Natación',
    status: 'Confirmada',
  }
])

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