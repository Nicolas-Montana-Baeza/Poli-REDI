<script setup>
import { computed, onMounted } from 'vue'

import {
  CalendarDays,
  Clock,
  Timer
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  parseReservationDateTime
} from '@/utils/reservationTime'

const reservationsStore = useReservationsStore()

onMounted(() => {
  reservationsStore.fetchMyReservations()
})

const getReservationEnd = (reservation) => {
  const start = parseReservationDateTime(reservation.startTime)

  if (!start) {
    return null
  }

  return new Date(
    start.getTime() +
    reservation.durationMinutes * 60000
  )
}

const isHistorical = (reservation) => {
  const end = getReservationEnd(reservation)

  return (
    reservation.status === 'CANCELLED' ||
    (end ? end.getTime() < Date.now() : false)
  )
}

const reservations = computed(() => {
  return reservationsStore.myReservations
    .filter(isHistorical)
    .slice()
    .sort((first, second) => {
      const firstDate =
        parseReservationDateTime(first.startTime)

      const secondDate =
        parseReservationDateTime(second.startTime)

      return (
        (secondDate?.getTime() || 0) -
        (firstDate?.getTime() || 0)
      )
    })
})

const statusLabel = (status) => {
  switch (status) {
    case 'CONFIRMED':
      return 'Finalizada'

    case 'PENDING':
      return 'Pendiente'

    case 'CANCELLED':
      return 'Cancelada'

    default:
      return status || 'Reserva'
  }
}

const statusClass = (status) => {
  return String(status || 'default').toLowerCase()
}
</script>

<template>
  <main class="history-view">

    <header class="page-header">

      <h1>
        Historial
      </h1>

      <p>
        Revisa reservas pasadas o canceladas.
      </p>

    </header>

    <div
      v-if="reservationsStore.myLoading"
      aria-label="Cargando historial"
    >
      <SkeletonLoader
        variant="reservations"
        :items="4"
      />
    </div>

    <div
      v-else-if="reservationsStore.myLoadingError"
      class="state-card error"
    >
      {{ reservationsStore.myLoadingError }}
    </div>

    <div
      v-else-if="!reservations.length"
      class="state-card"
    >
      Aun no tienes reservas historicas.
    </div>

    <section
      v-else
      class="history-list"
    >

      <article
        v-for="reservation in reservations"
        :key="reservation.id"
        class="history-card"
      >

        <div>

          <span
            class="status"
            :class="statusClass(reservation.status)"
          >
            {{ statusLabel(reservation.status) }}
          </span>

          <h2>
            {{ reservation.title || 'Reserva' }}
          </h2>

          <p>
            {{ reservation.resourceName || 'Recurso' }}
          </p>

        </div>

        <div class="details">

          <span>
            <CalendarDays :size="17" />
            {{ formatReservationDate(reservation.startTime) }}
          </span>

          <span>
            <Clock :size="17" />
            {{ formatReservationTimeRange(
              reservation.startTime,
              reservation.durationMinutes
            ) }}
          </span>

          <span>
            <Timer :size="17" />
            {{ reservation.durationMinutes }} minutos
          </span>

        </div>

      </article>

    </section>

  </main>
</template>

<style scoped>
.history-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: #0f172a;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.state-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 22px;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.history-list {
  display: flex;
  flex-direction: column;

  gap: 14px;
}

.history-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.status {
  display: inline-flex;

  padding: 6px 10px;

  border-radius: 999px;

  background: #eff6ff;
  color: #1d4ed8;

  font-size: 12px;
  font-weight: 800;
}

.status.cancelled {
  background: #fee2e2;

  color: #b91c1c;
}

.status.pending {
  background: #fef3c7;

  color: #92400e;
}

.history-card h2 {
  margin: 12px 0 0;

  color: #0f172a;

  font-size: 20px;
  font-weight: 800;
}

.history-card p {
  margin: 6px 0 0;

  color: #64748b;
}

.details {
  display: flex;
  flex-wrap: wrap;

  gap: 10px;
}

.details span {
  display: inline-flex;
  align-items: center;
  gap: 8px;

  background: #f8fafc;

  border-radius: 999px;

  padding: 8px 11px;

  color: #475569;

  font-size: 13px;
  font-weight: 700;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }
}
</style>
