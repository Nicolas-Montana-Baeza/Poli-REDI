<template>
  <div class="schedule-picker">
    <header>
      <h2>Horario del día {{ isDateFullyBooked.value ? '(Completo)' : '' }}</h2>
      <p class="day-label">{{ formattedDay }}</p>
    </header>

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

    <p class="hint">{{ isDateFullyBooked.value ? 'Haz clic en una franja horaria para cancelar la reserva.' : 'Haz clic en una franja horaria para marcarla como reservada.' }}</p>

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

    <!-- Cancel Booking Modal -->
    <div v-if="showCancelConfirmation" class="modal-overlay" @click="closeCancelConfirmation">
      <div class="modal-content" @click.stop>
        <h3 class="modal-title">¿Cancelar esta reserva?</h3>
        <div class="selected-slots">
          <p class="slot-item">
            {{ slotToCancel }}
          </p>
        </div>
        <div class="modal-buttons">
          <button class="btn-cancel-modal" @click="cancelBooking">Cancelar reserva</button>
          <button class="btn-keep-modal" @click="closeCancelConfirmation">Mantener</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../services/api.js'

const props = defineProps({
  selectedDate: {
    type: Object,
    default: null
  }
})

const fullyBookedDates = ref([])
const bookingConfig = ref({ startHour: 8, endHour: 20 })
const confirmedSlots = ref([])
const currentSelectedSlot = ref(null)
const showConfirmation = ref(false)
const showCancelConfirmation = ref(false)
const slotToCancel = ref(null)

// Fetch booking config and booked dates from API
onMounted(async () => {
  try {
    bookingConfig.value = await api.getBookingConfig()
    fullyBookedDates.value = await api.getFullyBookedDates()
  } catch (error) {
    console.error('Error loading schedule config:', error)
  }
})

const formattedDate = computed(() => {
  if (!props.selectedDate) return ''
  const { date, month, year } = props.selectedDate
  return `${year}-${String(month + 1).padStart(2, '0')}-${String(date).padStart(2, '0')}`
})

const isDateFullyBooked = computed(() => {
  return fullyBookedDates.value.includes(formattedDate.value)
})

const formattedDay = computed(() => {
  if (!formattedDate.value || !props.selectedDate) {
    return 'Día no seleccionado'
  }

  const { date, month, year } = props.selectedDate
  const dateObj = new Date(year, month, date)

  return dateObj.toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})


const slots = computed(() => {
  const result = []
  for (let hour = bookingConfig.value.startHour; hour < bookingConfig.value.endHour; hour += 1) {
    const key = `${hour}-${hour + 1}`
    result.push({
      key,
      label: `${formatHour(hour)} - ${formatHour(hour + 1)}`
    })
  }
  return result
})

watch(() => formattedDate.value, () => {
  confirmedSlots.value = []
  currentSelectedSlot.value = null
  showConfirmation.value = false
  // If newly selected day is fully booked, fill all slots
  if (isDateFullyBooked.value) {
    confirmedSlots.value = slots.value.map(slot => slot.key)
  }
})

watch(isDateFullyBooked, (isFullyBooked) => {
  if (isFullyBooked) {
    // Pre-fill all slots as booked for a fully booked day
    confirmedSlots.value = slots.value.map(slot => slot.key)
  } else {
    confirmedSlots.value = []
  }
})

watch(slots, () => {
  // When slots change (hour range changes), update fully booked slots
  if (isDateFullyBooked.value) {
    confirmedSlots.value = slots.value.map(slot => slot.key)
  }
})

function formatHour(hour) {
  const padded = String(hour).padStart(2, '0')
  return `${padded}:00`
}

function toggleReservation(key) {
  // Si es un horario confirmado, mostrar modal de cancelación
  if (confirmedSlots.value.includes(key)) {
    slotToCancel.value = key
    showCancelConfirmation.value = true
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
    currentSelectedSlot.value = null
    showConfirmation.value = false
  }
}

function searchAnother() {
  currentSelectedSlot.value = null
  closeConfirmation()
}

function closeCancelConfirmation() {
  showCancelConfirmation.value = false
  slotToCancel.value = null
}

function cancelBooking() {
  if (slotToCancel.value) {
    const index = confirmedSlots.value.indexOf(slotToCancel.value)
    if (index > -1) {
      confirmedSlots.value.splice(index, 1)
    }
    closeCancelConfirmation()
  }
}
</script>

<style scoped>
.schedule-picker {
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
  cursor: pointer;
}

.slot.reserved:hover {
  background: #dc2626;
}

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
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  max-width: 400px;
  width: 90%;
}

.modal-title {
  margin: 0 0 1rem 0;
  color: #333;
  font-size: 1.1rem;
}

.selected-slots {
  margin-bottom: 1.5rem;
}

.slot-item {
  margin: 0;
  padding: 0.8rem;
  background: #f3f4f6;
  border-radius: 6px;
  color: #333;
  font-weight: 500;
}

.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.btn-confirm-modal,
.btn-another-modal,
.btn-cancel-modal,
.btn-keep-modal {
  padding: 0.6rem 1.2rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s ease;
}

.btn-confirm-modal {
  background: #10b981;
  color: white;
}

.btn-confirm-modal:hover {
  background: #059669;
}

.btn-another-modal,
.btn-keep-modal {
  background: #d1d5db;
  color: #333;
}

.btn-another-modal:hover,
.btn-keep-modal:hover {
  background: #9ca3af;
}

.btn-cancel-modal {
  background: #ef4444;
  color: white;
}

.btn-cancel-modal:hover {
  background: #dc2626;
}


.slot.selected {
  background: #2563eb;
  color: #fff;
  font-weight: 600;
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
