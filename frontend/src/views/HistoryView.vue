<script setup>
import { computed, onMounted, ref } from 'vue'

import ReservationListCard from '@/components/reservations/ReservationListCard.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  isReservationHistorical,
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

const reservations = computed(() => {
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source
    .filter(isReservationHistorical)
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
    ? 'Aún no hay reservas históricas registradas.'
    : 'Aún no tienes reservas históricas.'
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

const getReservationDetailTo = (reservation) => {
  return {
    path: `/reservations/${reservation.id}`,
    query: {
      from: 'history'
    }
  }
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

    <section class="filters app-card">

      <label class="form-field">
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

      <label class="form-field">
        Desde
        <input
          v-model="fromDate"
          type="date"
        />
      </label>

      <label class="form-field">
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

      <ReservationListCard
        v-for="reservation in filteredReservations"
        :key="reservation.id"
        :reservation="reservation"
        mode="history"
        :detail-to="getReservationDetailTo(reservation)"
      />

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

.filters {
  padding: var(--space-5);

  display: grid;
  grid-template-columns:
    repeat(auto-fit, minmax(160px, 1fr));

  gap: 14px;
}

.filters input,
.filters select {
  width: 100%;
  height: 42px;

  padding: 0 12px;

  box-sizing: border-box;
}

.history-list {
  display: flex;
  flex-direction: column;

  gap: 14px;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }
}

@media (max-width: 520px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
