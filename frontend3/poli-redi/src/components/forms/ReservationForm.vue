<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  slot: Object, // { resource, hour }
  visible: Boolean
})

const emit = defineEmits(['close', 'submit'])

const form = ref({
  resource: '',
  date: '',
  hour: '',
  duration: 1,
  sport: '',
  participants: 10
})

watch(() => props.slot, (newSlot) => {
  if (newSlot) {
    form.value.resource = newSlot.resource.name
    form.value.hour = newSlot.hour
  }
})

const handleSubmit = () => {
  emit('submit', { ...form.value })
}

const handleClose = () => {
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="overlay">

    <div class="modal">

      <h2>Crear Reserva</h2>

      <label>Recurso</label>
      <input v-model="form.resource" disabled />

      <label>Hora</label>
      <input v-model="form.hour" disabled />

      <label>Fecha</label>
      <input type="date" v-model="form.date" />

      <label>Duración (horas)</label>
      <input type="number" min="1" v-model="form.duration" />

      <label>Deporte</label>
      <input v-model="form.sport" placeholder="Ej: fútbol" />

      <label>Participantes</label>
      <input type="number" min="1" v-model="form.participants" />

      <div class="actions">
        <button class="cancel" @click="handleClose">Cancelar</button>
        <button class="submit" @click="handleSubmit">Reservar</button>
      </div>

    </div>

  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0,0,0,0.4);

  display: flex;
  align-items: center;
  justify-content: center;
}

.modal {
  background: white;
  padding: 25px;
  border-radius: 12px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

h2 {
  margin-bottom: 10px;
}

input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 6px;
}

.actions {
  display: flex;
  justify-content: space-between;
  margin-top: 15px;
}

.cancel {
  background: #eee;
}

.submit {
  background: #f97316;
  color: white;
}
</style>