<script setup>
import { onMounted } from 'vue'
import AvailabilitySection from '../components/availability/AvailabilitySection.vue'
import api from '../services/api'

async function testMe() {
  try {
    const response = await api.get('/me')
    console.log('Usuario autenticado:', response.data)
  } catch (error) {
    console.error('Error consultando /me:', error.response?.data || error)
  }
}

onMounted(() => {
  testMe()
})
</script>

<template>
  <main class="availability-view">
    <AvailabilitySection />
  </main>
</template>

<style scoped>
.availability-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
</style>