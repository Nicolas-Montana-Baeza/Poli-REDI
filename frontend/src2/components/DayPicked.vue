<template>
  <div class="schedule-picker">
    <header>
      <h2>Horario del día</h2>
      <p class="day-label">{{ formattedDay }}</p>
    </header>

    <div class="controls">
      <label>
        Hora inicial
        <select v-model.number="startHour">
          <option
            v-for="hour in startOptions"
            :key="hour"
            :value="hour"
          >
            {{ formatHour(hour) }}
          </option>
        </select>
      </label>

      <label>
        Hora final
        <select v-model.number="endHour">
          <option
            v-for="hour in endOptions"
            :key="hour"
            :value="hour"
          >
            {{ formatHour(hour) }}
          </option>
        </select>
      </label>
    </div>

    <div class="schedule">
      <div
        v-for="slot in slots"
        :key="slot.key"
        class="slot"
        :class="{ 
          reserved: confirmedSlots.includes(slot.key),
          selected: currentSelectedSlot === slot.key,
          disabled: confirmedSlots.includes(slot.key)
        }"
        @click="toggleReservation(slot.key)"
      >
        {{ slot.label }}
      </div>
    </div>

    <p class="hint">Haz clic en una franja horaria para marcarla como reservada.</p>

    <!-- Confirmation Modal -->
    <div v-if="showConfirmation" class="modal-overlay" @click="closeConfirmation">
      <div class="modal-content" @click.stop>
        <h3 class="modal-title">¿Confirmar horario seleccionado?</h3>
        <div class="selected-slots">
          <p class="slot-item">
            {{ currentSelectedSlot }}
          </p>
        </div>
        <div class="modal-buttons">
          <button class="btn-confirm-modal" @click="confirmSelection">Confirmar</button>
          <button class="btn-another-modal" @click="searchAnother">Buscar otra</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const emit = defineEmits(['reservation-confirmed'])

const props = defineProps({
  day: {
    type: String,
    default: ''
  }
})

const startHour = ref(8)
const endHour = ref(17)
const confirmedSlots = ref([])
const currentSelectedSlot = ref(null)
const showConfirmation = ref(false)

const formattedDay = computed(() => {
  if (!props.day) {
    return 'Día no seleccionado'
  }

  const parsed = new Date(props.day)
  if (Number.isNaN(parsed.getTime())) {
    return props.day
  }

  return parsed.toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

const startOptions = computed(() => {
  return Array.from({ length: 24 }, (_, index) => index).filter(
    hour => hour < endHour.value
  )
})

const endOptions = computed(() => {
  return Array.from({ length: 25 }, (_, index) => index).filter(
    hour => hour > startHour.value
  )
})

const slots = computed(() => {
  const result = []
  for (let hour = startHour.value; hour < endHour.value; hour += 1) {
    const key = `${hour}-${hour + 1}`
    result.push({
      key,
      label: `${formatHour(hour)} - ${formatHour(hour + 1)}`
    })
  }
  return result
})

watch(startHour, value => {
  if (value >= endHour.value) {
    endHour.value = Math.min(24, value + 1)
  }
})

watch(endHour, value => {
  if (value <= startHour.value) {
    startHour.value = Math.max(0, value - 1)
  }
})

function formatHour(hour) {
  const padded = String(hour).padStart(2, '0')
  return `${padded}:00`
}

function toggleReservation(key) {
  // No permitir seleccionar horarios ya confirmados
  if (confirmedSlots.value.includes(key)) {
    return
  }
  
  // Si es el mismo que ya estaba seleccionado, deseleccionar
  if (currentSelectedSlot.value === key) {
    currentSelectedSlot.value = null
    showConfirmation.value = false
  } else {
    // Seleccionar el nuevo horario
    currentSelectedSlot.value = key
    showConfirmation.value = true
  }
}

function closeConfirmation() {
  showConfirmation.value = false
}

function confirmSelection() {
  if (currentSelectedSlot.value) {
    confirmedSlots.value.push(currentSelectedSlot.value)
    
    // Emit reservation confirmed event with day and time
    emit('reservation-confirmed', {
      day: props.day,
      time: currentSelectedSlot.value,
      date: props.day,
    })
    
    currentSelectedSlot.value = null
    showConfirmation.value = false
  }
}

function searchAnother() {
  currentSelectedSlot.value = null
  closeConfirmation()
}
</script>

<style scoped>
.day-picked {
  max-width: 420px;
  margin: 0 auto;
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 10px;
  background: #fafafa;
}

header {
  margin-bottom: 1rem;
}

.day-label {
  margin: 0.25rem 0 0;
  color: #555;
}

.controls {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.controls label {
  display: flex;
  flex-direction: column;
  font-size: 0.95rem;
  color: #333;
}

.controls select {
  margin-top: 0.4rem;
  padding: 0.4rem;
  border: 1px solid #bbb;
  border-radius: 6px;
  background: #fff;
}

.schedule {
  display: grid;
  gap: 0.5rem;
}

.slot {
  padding: 0.8rem;
  border-radius: 8px;
  background: #10b981;
  color: #fff;
  cursor: pointer;
  transition: background 0.2s ease;
  font-weight: 500;
}

.slot:hover:not(.disabled) {
  background: #059669;
}

.slot.reserved {
  background: #ef4444;
  color: #fff;
  font-weight: 600;
  cursor: default;
}

.slot.reserved:hover {
  background: #ef4444;
}

.slot.selected {
  background: #2563eb;
  color: #fff;
  font-weight: 600;
}

.slot.disabled {
  cursor: not-allowed;
}

.hint {
  margin-top: 1rem;
  color: #666;
  font-size: 0.9rem;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  max-width: 400px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  text-align: center;
}

.modal-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 1.5rem 0;
}

.selected-slots {
  background: #f3f4f6;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1.5rem;
  max-height: 200px;
  overflow-y: auto;
}

.slot-item {
  padding: 0.5rem;
  margin: 0.25rem 0;
  background: white;
  border-radius: 6px;
  font-size: 0.95rem;
  color: #2563eb;
  font-weight: 500;
}

.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.btn-confirm-modal,
.btn-another-modal {
  padding: 0.8rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease;
}

.btn-confirm-modal {
  background: #2563eb;
  color: white;
}

.btn-confirm-modal:hover {
  background: #1d4ed8;
}

.btn-another-modal {
  background: #e5e7eb;
  color: #111827;
}

.btn-another-modal:hover {
  background: #d1d5db;
}
</style>
