<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  ArrowLeft,
  CalendarDays,
  Clock,
  Timer,
  UserRound,
  XCircle
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  parseReservationDateTime
} from '@/utils/reservationTime'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const reservationsStore = useReservationsStore()
const cancelling = ref(false)

const reservationId = computed(() => {
  return Number(route.params.id)
})

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

const reservation = computed(() => {
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source.find(
    (item) => item.id === reservationId.value
  )
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

const getReservationEnd = (item) => {
  const start = parseReservationDateTime(item?.startTime)

  if (!start || !item) {
    return null
  }

  return new Date(
    start.getTime() +
    item.durationMinutes * 60000
  )
}

const isPast = computed(() => {
  const end = getReservationEnd(reservation.value)

  return end ? end.getTime() < Date.now() : false
})

const canCancel = computed(() => {
  return (
    reservation.value?.status !== 'CANCELLED' &&
    !isPast.value
  )
})

const statusLabel = computed(() => {
  switch (reservation.value?.status) {
    case 'CONFIRMED':
      return 'Confirmada'

    case 'PENDING':
      return 'Pendiente'

    case 'CANCELLED':
      return 'Cancelada'

    default:
      return reservation.value?.status || 'Reserva'
  }
})

const cancelReservation = async () => {
  if (!reservation.value) {
    return
  }

  const confirmed = window.confirm(
    'Deseas cancelar esta reserva?'
  )

  if (!confirmed) {
    return
  }

  cancelling.value = true

  try {
    await reservationsStore.cancelReservation(
      reservation.value.id
    )

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store conserva el mensaje de error para la vista.
  } finally {
    cancelling.value = false
  }
}

const goBack = () => {
  router.push('/reservations')
}
</script>

<template>
  <main class="detail-view">

    <button
      class="back-button"
      type="button"
      @click="goBack"
    >
      <ArrowLeft :size="18" />
      Volver
    </button>

    <div
      v-if="isLoading"
      aria-label="Cargando reserva"
    >
      <SkeletonLoader
        variant="resources"
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
      v-else-if="!reservation"
      class="state-card"
    >
      No se encontro la reserva solicitada.
    </div>

    <section
      v-else
      class="detail-panel"
    >

      <header class="detail-header">

        <div>

          <span
            class="status"
            :class="reservation.status?.toLowerCase()"
          >
            {{ statusLabel }}
          </span>

          <h1>
            {{ reservation.title || 'Reserva' }}
          </h1>

          <p>
            {{ reservation.resourceName || 'Recurso' }}
          </p>

        </div>

        <button
          v-if="canCancel"
          class="cancel-button"
          type="button"
          :disabled="cancelling"
          @click="cancelReservation"
        >
          <XCircle :size="18" />
          Cancelar
        </button>

      </header>

      <div
        v-if="reservationsStore.actionError"
        class="state-card error"
      >
        {{ reservationsStore.actionError }}
      </div>

      <div
        v-if="reservationsStore.actionSuccess"
        class="state-card success"
      >
        {{ reservationsStore.actionSuccess }}
      </div>

      <section class="details-grid">

        <article class="detail-item">

          <CalendarDays :size="22" />

          <span>
            Fecha
          </span>

          <strong>
            {{ formatReservationDate(reservation.startTime) }}
          </strong>

        </article>

        <article class="detail-item">

          <Clock :size="22" />

          <span>
            Horario
          </span>

          <strong>
            {{ formatReservationTimeRange(
              reservation.startTime,
              reservation.durationMinutes
            ) }}
          </strong>

        </article>

        <article class="detail-item">

          <Timer :size="22" />

          <span>
            Duracion
          </span>

          <strong>
            {{ reservation.durationMinutes }} minutos
          </strong>

        </article>

        <article class="detail-item">

          <UserRound :size="22" />

          <span>
            Usuario
          </span>

          <strong>
            {{ authStore.user?.isAdmin
              ? `Usuario #${reservation.userId}`
              : authStore.user?.fullName || authStore.user?.email || 'Usuario' }}
          </strong>

        </article>

      </section>

    </section>

  </main>
</template>

<style scoped>
.detail-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 18px;
}

.back-button,
.cancel-button {
  border: none;
  border-radius: 14px;

  cursor: pointer;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  padding: 11px 14px;

  font-weight: 800;

  transition: 0.2s;
}

.back-button {
  width: fit-content;

  background: #f1f5f9;

  color: #334155;
}

.back-button:hover {
  background: #e2e8f0;
}

.cancel-button {
  background: #fee2e2;
  border: 1px solid #fecaca;

  color: #b91c1c;

  white-space: nowrap;
}

.cancel-button:hover:not(:disabled) {
  background: #fecaca;
}

.cancel-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.state-card,
.detail-panel,
.detail-item {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;
}

.state-card {
  padding: 22px;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.state-card.success {
  background: #dcfce7;

  color: #166534;

  border-color: #bbf7d0;
}

.detail-panel {
  padding: 24px;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
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

.detail-header h1 {
  margin: 12px 0 0;

  color: #0f172a;

  font-size: 30px;
  font-weight: 900;
}

.detail-header p {
  margin: 7px 0 0;

  color: #64748b;
}

.details-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));

  gap: 14px;
}

.detail-item {
  padding: 18px;

  display: flex;
  flex-direction: column;

  gap: 9px;
}

.detail-item svg {
  color: #1d4ed8;
}

.detail-item span {
  color: #64748b;

  font-size: 13px;
  font-weight: 800;
}

.detail-item strong {
  color: #0f172a;

  overflow-wrap: anywhere;
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
  }

  .detail-header h1 {
    font-size: 26px;
  }

  .cancel-button {
    width: 100%;
  }
}
</style>
