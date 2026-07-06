<script setup>
import { computed, onMounted, ref } from 'vue'

import {
  CalendarDays,
  Clock,
  Timer
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  parseReservationDateTime
} from '@/utils/reservationTime'

const reservationsStore = useReservationsStore()
const authStore = useAuthStore()
const statusFilter = ref('ALL')
const fromDate = ref('')
const toDate = ref('')

onMounted(async () => {
  if (!authStore.user) {
    await authStore.loadAuthUser()
  }

  if (authStore.user?.isAdmin) {
    await reservationsStore.fetchReservations()
    return
  }

  await reservationsStore.fetchMyReservations()
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
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source
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

const isLoading = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.loading
    : reservationsStore.myLoading
})

const loadingError = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.loadingError
    : reservationsStore.myLoadingError
})

const emptyMessage = computed(() => {
  return authStore.user?.isAdmin
    ? 'Aun no hay reservas historicas registradas.'
    : 'Aun no tienes reservas historicas.'
})

const filteredReservations = computed(() => {
  return reservations.value.filter((reservation) => {
    const start = parseReservationDateTime(reservation.startTime)
    const reservationDate = start
      ? start.toISOString().slice(0, 10)
      : ''

    if (
      statusFilter.value !== 'ALL' &&
      reservation.status !== statusFilter.value
    ) {
      return false
    }

    if (fromDate.value && reservationDate < fromDate.value) {
      return false
    }

    if (toDate.value && reservationDate > toDate.value) {
      return false
    }

    return true
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
        {{ authStore.user?.isAdmin
          ? 'Revisa todo el historial de reservas del sistema.'
          : 'Revisa tus reservas pasadas o canceladas.' }}
      </p>

    </header>

    <section class="filters">

      <label>
        Estado
        <select v-model="statusFilter">
          <option value="ALL">
            Todos
          </option>
          <option value="CONFIRMED">
            Finalizadas
          </option>
          <option value="CANCELLED">
            Canceladas
          </option>
          <option value="PENDING">
            Pendientes
          </option>
        </select>
      </label>

      <label>
        Desde
        <input
          v-model="fromDate"
          type="date"
        />
      </label>

      <label>
        Hasta
        <input
          v-model="toDate"
          type="date"
        />
      </label>

    </section>

    <div
      v-if="isLoading"
      aria-label="Cargando historial"
    >
      <SkeletonLoader
        variant="reservations"
        :items="4"
      />
    </div>

    <div
      v-else-if="loadingError"
      class="state-card error"
    >
      {{ loadingError }}
    </div>

    <div
      v-else-if="!reservations.length"
      class="state-card"
    >
      {{ emptyMessage }}
    </div>

    <div
      v-else-if="!filteredReservations.length"
      class="state-card"
    >
      No hay reservas que coincidan con los filtros.
    </div>

    <section
      v-else
      class="history-list"
    >

      <article
        v-for="reservation in filteredReservations"
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

.filters {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 18px;

  display: grid;
  grid-template-columns:
    repeat(3, minmax(0, 1fr));

  gap: 14px;
}

.filters label {
  color: #334155;

  display: flex;
  flex-direction: column;

  gap: 7px;

  font-size: 13px;
  font-weight: 800;
}

.filters input,
.filters select {
  width: 100%;
  height: 42px;

  border: 1px solid #dbe2ea;
  border-radius: 12px;

  padding: 0 12px;

  box-sizing: border-box;
  outline: none;
}

.filters input:focus,
.filters select:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
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

  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
