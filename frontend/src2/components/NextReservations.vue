<script setup>
defineProps({
  reservations: {
    type: Array,
    default: () => []
  }
})
</script>

<template>
  <div class="next-reservations">
    <h2>Mis próximas reservas</h2>

    <div
      v-for="reservation in reservations"
      :key="reservation.id"
      class="reservation-card"
    >
      <div class="reservation-info">
        <h3>{{ reservation.place }}</h3>
        <p>{{ reservation.date }}</p>
        <p>{{ reservation.location }}</p>
      </div>

      <span class="status">
        {{ reservation.status }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.next-reservations {
  margin-top: 20px;
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

h2 {
  margin-bottom: 16px;
}

.reservation-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid #eee;
}

.reservation-info h3 {
  margin: 0;
  font-size: 1rem;
}

.reservation-info p {
  margin: 3px 0;
  color: #666;
  font-size: 0.85rem;
}

.status {
  background: #dff7e4;
  color: #1b7f3b;
  padding: 5px 10px;
  border-radius: 10px;
  font-size: 0.8rem;
  font-weight: bold;
}
</style>