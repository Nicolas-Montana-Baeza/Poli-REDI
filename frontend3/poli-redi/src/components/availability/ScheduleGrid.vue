<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  resources: {
    type: Array,
    default: () => []
  },

  reservations: {
    type: Array,
    default: () => []
  },

  startHour: {
    type: Number,
    default: 8
  },

  endHour: {
    type: Number,
    default: 22
  }
})

const emit = defineEmits(['slot-selected'])

const selectedSlot = ref(null)

/* Horas dinámicas */
const hours = computed(() => {
  const arr = []

  for (let i = props.startHour; i < props.endHour; i++) {
    arr.push(`${i}:00`)
  }

  return arr
})

/* Selección */
const handleSelect = (resource, hour) => {
  const reservation = getReservation(resource.id, hour)

  // no seleccionar slots ocupados
  if (reservation) return

  const slot = {
    resource,
    hour
  }

  selectedSlot.value = slot

  emit('slot-selected', slot)
}

const isSelected = (resource, hour) => {
  return (
    selectedSlot.value &&
    selectedSlot.value.resource.id === resource.id &&
    selectedSlot.value.hour === hour
  )
}

/* Buscar reservas */
const getReservation = (resourceId, hour) => {
  return props.reservations.find(
    r =>
      r.resourceId === resourceId &&
      r.hour === hour
  )
}

/* Clase visual */
const getSlotClass = (resource, hour) => {
  const reservation = getReservation(resource.id, hour)

  if (reservation) {
    if (reservation.type === 'priority') {
      return 'priority'
    }

    return 'reserved'
  }

  if (isSelected(resource, hour)) {
    return 'selected'
  }

  return 'available'
}
</script>

<template>
  <div class="schedule-wrapper">

    <div
      class="schedule-grid"
      :style="{
        gridTemplateColumns: `80px repeat(${resources.length}, minmax(180px, 1fr))`
      }"
    >

      <!-- HEADER -->
      <div class="header-cell empty"></div>

      <div
        v-for="resource in resources"
        :key="resource.id"
        class="header-cell"
      >
        {{ resource.name }}
      </div>

      <!-- BODY -->
      <template
        v-for="hour in hours"
        :key="hour"
      >

        <!-- HORA -->
        <div class="time-cell">
          {{ hour }}
        </div>

        <!-- SLOTS -->
        <div
          v-for="resource in resources"
          :key="resource.id + hour"
          class="slot"
          :class="getSlotClass(resource, hour)"
          @click="handleSelect(resource, hour)"
        >

          <div
            v-if="getReservation(resource.id, hour)"
            class="reservation-content"
          >
            {{ getReservation(resource.id, hour).title }}
          </div>

        </div>

      </template>

    </div>

  </div>
</template>

<style scoped>
.schedule-wrapper {
  width: 100%;
  overflow-x: auto;

  background: white;
  border-radius: 14px;
  border: 1px solid #e5e7eb;
}

/* GRID */
.schedule-grid {
  display: grid;

  width: max-content;
  min-width: 100%;
}

/* HEADER */
.header-cell {
  background: #1e3a8a;
  color: white;

  padding: 16px;

  text-align: center;
  font-weight: 600;

  border-right: 1px solid rgba(255,255,255,0.1);

  min-width: 180px;
}

.empty {
  min-width: 80px;
}

/* TIME */
.time-cell {
  height: 72px;
  min-width: 80px;

  display: flex;
  align-items: center;
  justify-content: center;

  background: #f9fafb;

  border-right: 1px solid #eee;
  border-bottom: 1px solid #eee;

  font-size: 14px;
  color: #6b7280;

  position: sticky;
  left: 0;
  z-index: 2;
}

/* SLOT */
.slot {
  height: 72px;
  min-width: 180px;

  border-right: 1px solid #eee;
  border-bottom: 1px solid #eee;

  cursor: pointer;

  transition: 0.2s;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 13px;
  font-weight: 500;
}

/* ESTADOS */

.available {
  background: white;
}

.available:hover {
  background: #dbeafe;
}

.selected {
  background: #f97316;
  color: white;
}

.reserved {
  background: #eb2525;
  color: white;
}

.priority {
  background: #f97316;
  color: white;
}

/* CONTENIDO */
.reservation-content {
  padding: 6px 10px;
  text-align: center;
}

/* SCROLL */
.schedule-wrapper::-webkit-scrollbar {
  height: 8px;
}

.schedule-wrapper::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 20px;
}

/* MOBILE */
@media (max-width: 768px) {
  .header-cell,
  .slot {
    min-width: 140px;
  }
}
</style>