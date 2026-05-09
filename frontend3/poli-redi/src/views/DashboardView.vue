<script setup>
import FacilityCarousel from '../components/dashboard/FacilityCarousel.vue'

const facilities = [
  { name: 'Cancha 1', type: 'Exterior', status: 'available' },
  { name: 'Cancha 2', type: 'Exterior', status: 'busy' },
  { name: 'Cancha 3', type: 'Exterior', status: 'available' },
  { name: 'Gimnasio', type: 'Interior', status: 'maintenance' },
  { name: 'Piscina', type: 'Interior', status: 'available' }
]

const resources = [
  { id: 1, name: 'Cancha 1'},
  { id: 2, name: 'Cancha 2'},
  { id: 3, name: 'Cancha 3'},
  { id: 4, name: 'Gimnasio'},
  { id: 5, name: 'Piscina'}
]

// ✔ función correcta
const handleSelect = (facility) => {
  console.log('Seleccionado:', facility)
}

// ✔ separada correctamente
const handleSlot = (slot) => {
  console.log('Slot seleccionado:', slot)
}
</script>

<template>
  <div class="view">

    <FacilityCarousel 
      :facilities="facilities"
      @select="handleSelect"
    />

  </div>
</template>

<style scoped>
.view {
  padding: 20px;
}
</style>