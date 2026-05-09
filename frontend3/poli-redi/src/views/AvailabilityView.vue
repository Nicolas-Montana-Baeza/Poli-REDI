<script setup>
import { ref } from 'vue'
import ScheduleGrid from '../components/availability/ScheduleGrid.vue'
import ReservationForm from '../components/forms/ReservationForm.vue'

const resources = [
  { id: 1, name: 'Cancha 1' },
  { id: 2, name: 'Cancha 2' },
  { id: 3, name: 'Cancha 3' },
  { id: 4, name: 'Gimnasio' },
  { id: 5, name: 'Piscina' }
]

const selectedSlot = ref(null)
const showForm = ref(false)

const handleSlot = (slot) => {
  selectedSlot.value = slot
  showForm.value = true
}

const handleSubmit = (data) => {
  console.log('Reserva creada:', data)

  // 🔥 aquí irá tu API después

  showForm.value = false
}

const handleClose = () => {
  showForm.value = false
}
</script>

<template>
  <div class="view">

    <h1>Disponibilidad</h1>

    <ScheduleGrid 
      :resources="resources"
      @slot-selected="handleSlot"
    />

    <ReservationForm
      :slot="selectedSlot"
      :visible="showForm"
      @submit="handleSubmit"
      @close="handleClose"
    />

  </div>
</template>

<style scoped>
.view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>