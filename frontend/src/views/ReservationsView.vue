<script setup>
import {
  computed,
  onMounted,
  ref
} from 'vue'
import {
  useRoute,
  useRouter
} from 'vue-router'

import ReservationListCard from '@/components/reservations/ReservationListCard.vue'
import ReservationForm from '@/components/forms/ReservationForm.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'

import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'

import {
  getReservationDateKey,
  isReservationActionable,
  isReservationCancelable,
  isReservationHistorical,
  parseReservationDateTime
} from '@/utils/reservationTime'

const route = useRoute()
const router = useRouter()

const reservationsStore = useReservationsStore()
const authStore = useAuthStore()

const cancellingId = ref(null)
const selectedReservation = ref(null)

/* HISTORY FILTERS */
const statusFilter = ref('ALL')
const fromDate = ref('')
const toDate = ref('')

/* TAB */
const activeTab = computed(() => {
  return route.query.tab === 'history'
    ? 'history'
    : 'active'
})

const setTab = async (tab) => {
  const query = {
    ...route.query
  }

  if (tab === 'history') {
    query.tab = 'history'
  } else {
    delete query.tab
  }

  await router.replace({
    path: '/reservations',
    query
  })
}

/* LOAD */
onMounted(async () => {
  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()

  const user =
    await authStore.loadAuthUser()

  if (user?.isAdmin) {
    await reservationsStore.fetchReservations()
    return
  }

  await reservationsStore.fetchMyReservations()
})

/* SOURCE */
const sourceReservations = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations
})

/* ACTIVE */
const activeReservations = computed(() => {
  return sourceReservations.value
    .filter(isReservationActionable)
    .slice()
    .sort((first, second) => {
      const firstDate =
        parseReservationDateTime(first.startTime)

      const secondDate =
        parseReservationDateTime(second.startTime)

      return (
        (firstDate?.getTime() || 0) -
        (secondDate?.getTime() || 0)
      )
    })
})

/* HISTORY */
const historicalReservations = computed(() => {
  return sourceReservations.value
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

const filteredHistory = computed(() => {
  return historicalReservations.value.filter(
    reservation => {
      const start =
        parseReservationDateTime(
          reservation.startTime
        )

      const reservationDate =
        start
          ? getReservationDateKey(start)
          : ''

      if (
        statusFilter.value !== 'ALL' &&
        reservation.status !== statusFilter.value
      ) {
        return false
      }

      if (
        fromDate.value &&
        reservationDate < fromDate.value
      ) {
        return false
      }

      if (
        toDate.value &&
        reservationDate > toDate.value
      ) {
        return false
      }

      return true
    }
  )
})

const displayedReservations = computed(() => {
  return activeTab.value === 'history'
    ? filteredHistory.value
    : activeReservations.value
})

/* STATES */
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
  if (activeTab.value === 'history') {
    return authStore.user?.isAdmin
      ? 'Aún no hay reservas históricas registradas.'
      : 'Aún no tienes reservas en el historial.'
  }

  return authStore.user?.isAdmin
    ? 'No hay reservas activas registradas.'
    : 'No tienes reservas activas.'
})

const pageDescription = computed(() => {
  if (authStore.user?.isAdmin) {
    return 'Consulta y gestiona las reservas activas y el historial del sistema.'
  }

  return 'Consulta y gestiona tus reservas activas y revisa tu historial.'
})

/* DETAIL */
const openReservationDetail = async (
  reservation
) => {
  reservationsStore.clearActionError?.()
  reservationsStore.clearActionSuccess?.()

  selectedReservation.value =
    reservation

  const detail =
    await reservationsStore.fetchReservationDetail(
      reservation.id
    )

  if (
    detail &&
    String(selectedReservation.value?.id) ===
      String(reservation.id)
  ) {
    selectedReservation.value =
      detail
  }
}

const closeReservationDetail = () => {
  selectedReservation.value = null

  reservationsStore.clearActionError?.()
}

const canCancelSelected = computed(() => {
  return (
    activeTab.value === 'active' &&
    selectedReservation.value &&
    isReservationCancelable(
      selectedReservation.value
    )
  )
})

/* CANCEL */
const cancelReservation = async (
  reservation
) => {
  cancellingId.value =
    reservation.id

  try {
    const cancelled =
      await reservationsStore.cancelReservation(
        reservation.id
      )

    if (
      selectedReservation.value &&
      String(selectedReservation.value.id) ===
        String(reservation.id)
    ) {
      selectedReservation.value =
        cancelled
    }

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store mantiene el mensaje de error.
  } finally {
    cancellingId.value = null
  }
}
</script>

<template>
  <main class="reservations-view">

    <!-- HEADER -->
    <header class="page-header">

      <h1>
        Reservas
      </h1>

      <p>
        {{ pageDescription }}
      </p>

    </header>

    <!-- TABS -->
    <nav
      class="reservation-tabs"
      aria-label="Secciones de reservas"
    >

      <button
        type="button"
        :class="{
          active: activeTab === 'active'
        }"
        @click="setTab('active')"
      >
        Activas

        <span class="tab-count">
          {{ activeReservations.length }}
        </span>
      </button>

      <button
        type="button"
        :class="{
          active: activeTab === 'history'
        }"
        @click="setTab('history')"
      >
        Historial

        <span class="tab-count">
          {{ historicalReservations.length }}
        </span>
      </button>

    </nav>

    <!-- HISTORY FILTERS -->
    <section
      v-if="activeTab === 'history'"
      class="filters app-card"
    >

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

          <option value="REJECTED">
            Rechazadas
          </option>

          <option value="EXPIRED">
            Expiradas
          </option>
        </select>
      </label>

      <label class="form-field">
        Desde

        <input
          v-model="fromDate"
          type="date"
        >
      </label>

      <label class="form-field">
        Hasta

        <input
          v-model="toDate"
          type="date"
        >
      </label>

    </section>

    <!-- LOADING -->
    <div
      v-if="isLoading"
      aria-label="Cargando reservas"
    >
      <SkeletonLoader
        variant="reservations"
        :items="4"
      />
    </div>

    <!-- LOAD ERROR -->
    <div
      v-else-if="loadingError"
      class="state-card error"
    >
      {{ loadingError }}
    </div>

    <section
      v-else
      class="content"
    >

      <!-- ACTION ERROR -->
      <div
        v-if="
          reservationsStore.actionError &&
          !selectedReservation
        "
        class="state-card error"
      >
        {{ reservationsStore.actionError }}
      </div>

      <!-- ACTION SUCCESS -->
      <div
        v-if="reservationsStore.actionSuccess"
        class="state-card success"
      >
        {{ reservationsStore.actionSuccess }}
      </div>

      <!-- EMPTY -->
      <div
        v-if="
          activeTab === 'history' &&
          historicalReservations.length &&
          !filteredHistory.length
        "
        class="state-card"
      >
        No hay reservas que coincidan con los filtros.
      </div>

      <div
        v-else-if="!displayedReservations.length"
        class="state-card"
      >
        {{ emptyMessage }}
      </div>

      <!-- LIST -->
      <section
        v-else
        class="reservations-list"
      >

        <ReservationListCard
          v-for="reservation in displayedReservations"
          :key="reservation.id"
          :reservation="reservation"
          :mode="
            activeTab === 'history'
              ? 'history'
              : 'active'
          "
          @open-detail="openReservationDetail"
        />

      </section>

    </section>

    <!-- SHARED DETAIL MODAL -->
    <ReservationForm
      :visible="Boolean(selectedReservation)"
      mode="detail"
      :reservation="selectedReservation"
      :can-cancel="Boolean(canCancelSelected)"
      :cancel-disabled="
        cancellingId === selectedReservation?.id
      "
      :error-message="
        reservationsStore.actionError
      "
      @close="closeReservationDetail"
      @cancel="cancelReservation"
    />

  </main>
</template>

<style scoped>
.reservations-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: var(--color-text);
}

.page-header p {
  margin-top: 8px;

  color: var(--color-text-muted);
}

/* TABS */
.reservation-tabs {
  display: flex;
  gap: 8px;

  padding: 6px;

  width: fit-content;

  background: var(--color-surface-muted);

  border-radius: var(--radius-lg);
}

.reservation-tabs button {
  display: inline-flex;
  align-items: center;
  gap: 8px;

  min-height: 42px;

  padding: 0 16px;

  border: 0;
  border-radius: var(--radius-md);

  background: transparent;

  color: var(--color-text-muted);

  font-weight: 700;

  cursor: pointer;

  transition: 0.2s;
}

.reservation-tabs button:hover {
  color: var(--color-primary);
}

.reservation-tabs button.active {
  background: var(--color-surface);

  color: var(--color-primary);

  box-shadow: var(--shadow-card);
}

.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;

  min-width: 24px;
  height: 24px;

  padding: 0 6px;

  border-radius: var(--radius-pill);

  background: var(--color-primary-soft);

  color: var(--color-primary-strong);

  font-size: 12px;
}

/* FILTERS */
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

/* CONTENT */
.content,
.reservations-list {
  display: flex;
  flex-direction: column;

  gap: 14px;
}

.state-card {
  border-radius: var(--radius-lg);
}

.state-card.error {
  background: var(--color-error-soft);
  color: var(--color-error);
  border-color: var(--color-error-border);
}

.state-card.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border-color: var(--color-success-border);
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }

  .reservation-tabs {
    width: 100%;
  }

  .reservation-tabs button {
    flex: 1;
    justify-content: center;
  }
}

@media (max-width: 520px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
