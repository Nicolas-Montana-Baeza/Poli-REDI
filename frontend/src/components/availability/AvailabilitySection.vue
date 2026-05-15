<script setup>
import { ref } from 'vue'

import CalendarToolbar from './CalendarToolbar.vue'
import CalendarMini from './CalendarMini.vue'
import ScheduleGrid from './ScheduleGrid.vue'
import { useReservationsStore } from '@/stores/reservations'
import ReservationForm from '../forms/ReservationForm.vue'
import { useResourcesStore } from '@/stores/resources'

/*RESOURCES*/
const resourcesStore =
  useResourcesStore()

/* CURRENT DATE */
const currentDate = ref(
  'Lunes 12 Mayo'
)
/* RESERVATIONS */
const reservationsStore =
  useReservationsStore()
/* SELECTED SLOT */
const selectedSlot = ref(null)

/* MODAL */
const showReservationForm = ref(false)

/* SLOT SELECT */
const handleSlotSelected = (
  slot
) => {

  selectedSlot.value = slot

  showReservationForm.value = true

  console.log(
    'Slot seleccionado:',
    slot
  )
}

/* CLOSE */
const closeReservationForm = () => {

  showReservationForm.value = false
}

/* SUBMIT */
const submitReservation = (
  reservation
) => {

  console.log(
    'Reserva creada:',
    reservation
  )

  reservationsStore.addReservation({
    resourceId:
      reservation.resource.id,

    hour:
      reservation.hour,

    title:
      reservation.sport ||
      'Reserva',

    type:
      'normal'
  })

  showReservationForm.value = false
}

/* CALENDAR */
const handleDateSelect = (
  date
) => {

  console.log(
    'Fecha seleccionada:',
    date
  )
}

/* TOOLBAR */
const previousDay = () => {
  console.log('Prev day')
}

const nextDay = () => {
  console.log('Next day')
}

const goToday = () => {
  console.log('Today')
}
</script>

<template>
  <section class="availability-section">

    <!-- HEADER -->
    <div class="section-header">

      <div>

        <h2>
          Disponibilidad
        </h2>

        <p>
          Revisa horarios y
          selecciona un bloque
          disponible.
        </p>

      </div>

    </div>

    <!-- TOOLBAR -->
    <CalendarToolbar
      :current-date="
        currentDate
      "

      @prev-day="
        previousDay
      "

      @next-day="
        nextDay
      "

      @today="
        goToday
      "
    />

    <!-- CONTENT -->
    <div class="content">

      <!-- LEFT -->
      <div
        class="calendar-container"
      >

        <CalendarMini
          @select-date="
            handleDateSelect
          "
        />

        <!-- SLOT INFO -->
        <div
          v-if="selectedSlot"
          class="selection-card"
        >

          <h3>
            Selección actual
          </h3>

          <p>

            <strong>
              {{
                selectedSlot
                  .resource.name
              }}
            </strong>

          </p>

          <span>
            {{
              selectedSlot.hour
            }}
          </span>

        </div>

      </div>

      <!-- RIGHT -->
      <div
        class="grid-container"
      >

        <ScheduleGrid
          :resources="
            resourcesStore.resources
          "

          :reservations="
            reservationsStore.reservations
          "

          @slot-selected="
            handleSlotSelected
          "
        />

      </div>

    </div>

<ReservationForm
  :visible="showReservationForm"

  :slot="selectedSlot"

  :resources="resources"

  @close="closeReservationForm"

  @submit="submitReservation"
/>

  </section>
</template>

<style scoped>
.availability-section {
  display: flex;
  flex-direction: column;

  gap: 24px;
}

/* HEADER */
.section-header h2 {
  margin: 0;

  font-size: 28px;
  font-weight: 800;

  color: #0f172a;
}

.section-header p {
  margin-top: 6px;

  font-size: 15px;

  color: #64748b;
}

/* CONTENT */
.content {
  display: grid;

  grid-template-columns:
    340px 1fr;

  gap: 24px;

  align-items: start;
}

/* LEFT */
.calendar-container {
  display: flex;
  flex-direction: column;

  gap: 20px;

  position: sticky;

  top: 90px;
}

/* RIGHT */
.grid-container {
  min-width: 0;
}

/* CARD */
.selection-card {
  background: white;

  border-radius: 22px;

  padding: 20px;

  border: 1px solid #e2e8f0;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.selection-card h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 700;

  color: #0f172a;
}

.selection-card p {
  margin: 14px 0 6px;

  color: #334155;
}

.selection-card span {
  color: #2563eb;

  font-size: 14px;
  font-weight: 700;
}

/* TABLET */
@media (max-width: 1024px) {
  .content {
    grid-template-columns:
      1fr;
  }

  .calendar-container {
    position: relative;

    top: 0;
  }
}

/* MOBILE */
@media (max-width: 768px) {
  .availability-section {
    gap: 20px;
  }

  .section-header h2 {
    font-size: 24px;
  }

  .section-header p {
    font-size: 14px;
  }
}
</style>