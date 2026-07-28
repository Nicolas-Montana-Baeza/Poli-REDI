<script setup>
import { computed, onMounted, ref } from 'vue'

import ReservationListCard from '@/components/reservations/ReservationListCard.vue'
import WorkshopEnrollmentHistoryCard from '@/components/workshops/WorkshopEnrollmentHistoryCard.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import { useWorkshopsStore } from '@/stores/workshops'
import {
  getReservationDateKey,
  isReservationHistorical,
  parseReservationDateTime
} from '@/utils/reservationTime'

const reservationsStore = useReservationsStore()
const workshopsStore = useWorkshopsStore()
const authStore = useAuthStore()
const typeFilter = ref('ALL')
const statusFilter = ref('ALL')
const fromDate = ref('')
const toDate = ref('')

onMounted(async () => {
  const user = await authStore.loadAuthUser()

  if (user?.isAdmin) {
    await Promise.all([
      reservationsStore.fetchReservations(),
      workshopsStore.fetchMyEnrollments()
    ])
    return
  }

  await Promise.all([
    reservationsStore.fetchMyReservations(),
    workshopsStore.fetchMyEnrollments()
  ])
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
  const reservationsLoading = authStore.user?.isAdmin
    ? reservationsStore.loading
    : reservationsStore.myLoading
  return reservationsLoading || workshopsStore.historyLoading
})

const loadingError = computed(() => {
  const reservationError = authStore.user?.isAdmin
    ? reservationsStore.loadingError
    : reservationsStore.myLoadingError
  return reservationError || workshopsStore.historyLoadingError
})

const emptyMessage = computed(() => {
  return authStore.user?.isAdmin
    ? 'Aún no hay reservas históricas ni inscripciones propias a talleres.'
    : 'Aún no tienes reservas históricas ni inscripciones a talleres.'
})

const historyItems = computed(() => {
  const reservationItems = reservations.value.map((reservation) => ({
    kind: 'RESERVATION',
    id: `reservation-${reservation.id}`,
    date: parseReservationDateTime(reservation.startTime),
    status: reservation.status,
    value: reservation
  }))
  const workshopItems = workshopsStore.myEnrollments.map((enrollment) => ({
    kind: 'WORKSHOP',
    id: `workshop-enrollment-${enrollment.id}`,
    date: parseReservationDateTime(enrollment.enrolledAt),
    status: enrollment.status,
    value: enrollment
  }))

  return [...reservationItems, ...workshopItems].sort((first, second) => (
    (second.date?.getTime() || 0) - (first.date?.getTime() || 0)
  ))
})

const filteredHistory = computed(() => {
  return historyItems.value.filter((item) => {
    const itemDate = item.date ? getReservationDateKey(item.date) : ''

    if (
      statusFilter.value !== 'ALL' &&
      item.status !== statusFilter.value
    ) {
      return false
    }

    if (typeFilter.value !== 'ALL' && item.kind !== typeFilter.value) {
      return false
    }

    if (fromDate.value && itemDate < fromDate.value) {
      return false
    }

    if (toDate.value && itemDate > toDate.value) {
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
          ? 'Revisa las reservas del sistema y tus inscripciones a talleres.'
          : 'Revisa tus reservas pasadas o canceladas e inscripciones a talleres.' }}
      </p>

    </header>

    <section class="filters app-card">

      <label class="form-field">
        Tipo
        <select v-model="typeFilter">
          <option value="ALL">Todos</option>
          <option value="RESERVATION">Reservas</option>
          <option value="WORKSHOP">Talleres</option>
        </select>
      </label>

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

      <p class="filter-hint">
        Para los talleres, el rango considera la fecha de inscripción.
      </p>

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
      v-else-if="!historyItems.length"
      class="state-card"
    >
      {{ emptyMessage }}
    </div>

    <div
      v-else-if="!filteredHistory.length"
      class="state-card"
    >
      No hay elementos que coincidan con los filtros.
    </div>

    <section
      v-else
      class="history-list"
    >

      <template v-for="item in filteredHistory" :key="item.id">
        <ReservationListCard
          v-if="item.kind === 'RESERVATION'"
          :reservation="item.value"
          mode="history"
          :detail-to="getReservationDetailTo(item.value)"
        />

        <WorkshopEnrollmentHistoryCard
          v-else
          :enrollment="item.value"
        />
      </template>

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

.filter-hint {
  grid-column: 1 / -1;
  margin: 0;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 500;
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
