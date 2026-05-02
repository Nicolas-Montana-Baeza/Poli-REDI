<template>
  <div class="quick-actions">
    <h2>Acciones rápidas</h2>

    <div class="actions-grid">
      <button @click="$emit('reserve')">📅 Reservar</button>
      <button @click="$emit('my-reservations')">📖 Mis reservas</button>
      <button @click="$emit('availability')">🕒 Disponibilidad</button>
      <button @click="$emit('cancel')">❌ Cancelar</button>
    </div>
  </div>
</template>

<script setup>
// No logic yet
</script>

<style scoped>
.quick-actions {
  background: white;
  border-radius: 16px;
  padding: 20px;
  margin: 20px 0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

button {
  border: none;
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  font-size: 1rem;
  font-weight: bold;
  background: #f4f4f4;
}

button:hover {
  background: #e8e8e8;
}
</style>