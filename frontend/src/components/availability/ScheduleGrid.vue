<script setup>
import { computed, ref } from 'vue'

import ResourceColumn from './ResourceColumn.vue'

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

const emit = defineEmits([
  'slot-selected',
  'reservation-created'
])


/* HOURS */
const hours = computed(() => {
  const arr = []

  for (
    let i = props.startHour;
    i < props.endHour;
    i++
  ) {

    arr.push(
      `${String(i).padStart(2, '0')}:00`
    )
  }

  return arr
})

/* BUILD SLOTS */
const buildSlots = (resourceId) => {
  return hours.value.map((hour) => {

    const reservation =
      props.reservations.find(
        r =>
          r.resourceId === resourceId &&
          r.hour === hour
      )

    return {
      time: hour,

      available: !reservation,

      reserved: !!reservation,

      title:
        reservation?.title || null,

      type:
        reservation?.type || null
    }
  })
}

/* SLOT SELECT */
const handleSlotSelect = (
  resource,
  time
) => {

  emit('slot-selected', {
    resource,
    hour: time
  })
}

/* CLOSE */
const closeReservationForm = () => {
  showReservationForm.value = false
}

/* SUBMIT */
const submitReservation = (
  reservation
) => {

  emit(
    'reservation-created',
    reservation
  )

  console.log(
    'Reserva creada:',
    reservation
  )

  showReservationForm.value = false
}
</script>

<template>
  <section class="schedule-section">

    <!-- HEADER -->
    <div class="section-header">

      <div>

        <h2>
          Disponibilidad
        </h2>

        <p>
          Selecciona un horario disponible.
        </p>

      </div>

    </div>

    <!-- COLUMNS -->
    <div class="columns-wrapper">

      <ResourceColumn
        v-for="resource in resources"
        :key="resource.id"

        :resource="resource"

        :slots="
          buildSlots(resource.id)
        "

        @slot-selected="
          (time) =>
            handleSlotSelect(
              resource,
              time
            )
        "
      />

    </div>

      "
    />

  </section>
</template>

<style scoped>
.schedule-section {
  display: flex;
  flex-direction: column;

  gap: 20px;
}

/* HEADER */
.section-header h2 {
  margin: 0;

  font-size: 24px;
  font-weight: 700;

  color: #0f172a;
}

.section-header p {
  margin-top: 4px;

  font-size: 14px;

  color: #64748b;
}

/* WRAPPER */
.columns-wrapper {
  display: flex;

  gap: 20px;

  overflow-x: auto;

  padding-bottom: 10px;

  scroll-behavior: smooth;
}

/* SCROLLBAR */
.columns-wrapper::-webkit-scrollbar {
  height: 8px;
}

.columns-wrapper::-webkit-scrollbar-thumb {
  background: #cbd5e1;

  border-radius: 999px;
}

/* MOBILE */
@media (max-width: 768px) {
  .columns-wrapper {
    gap: 16px;
  }
}
</style>