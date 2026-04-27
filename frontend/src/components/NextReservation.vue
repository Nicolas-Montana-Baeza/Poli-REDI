<template>
  <div class="next-reservation">
    <div v-if="reservation">
      <h2>Próxima Reserva</h2>
      <div class="reservation-card">
        <p v-if="reservation.guestName || reservation.name"><strong>Huésped:</strong> {{ reservation.guestName || reservation.name }}</p>
        <p><strong>Fecha:</strong> {{ formattedDate }}</p>
        <p><strong>Hora:</strong> {{ formattedTime }}</p>
        <p v-if="reservation.location"><strong>Ubicación:</strong> {{ reservation.location }}</p>
        <p v-if="reservation.notes"><strong>Notas:</strong> {{ reservation.notes }}</p>
      </div>
    </div>
    <div v-else class="empty-state">
      <p>Sin reservas próximas.</p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  reservation: {
    type: Object,
    default: null
  }
})

const formattedDate = computed(() => {
  if (!props.reservation?.day && !props.reservation?.date) return 'TBD'
  
  const dateStr = props.reservation?.day || props.reservation?.date
  const date = new Date(dateStr)
  
  if (Number.isNaN(date.getTime())) return 'TBD'
  
  return date.toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

const formattedTime = computed(() => {
  if (!props.reservation?.time) return 'TBD'
  
  // Handle time slot format like "08:00 - 09:00"
  if (typeof props.reservation.time === 'string') {
    return props.reservation.time
  }
  
  return props.reservation.time
})
</script>

<style scoped>
.next-reservation {
  border: 1px solid #dcdcdc;
  border-radius: 10px;
  padding: 16px;
  background: #fff;
  max-width: 420px;
}

.next-reservation h2 {
  margin: 0 0 12px;
  font-size: 18px;
}

.reservation-card {
  display: grid;
  gap: 8px;
}

.reservation-card p {
  margin: 0;
  font-size: 14px;
}

.empty-state {
  color: #666;
  font-size: 14px;
}
</style>
