<script setup>
import { computed, onMounted, ref } from 'vue'

import {
  CalendarDays,
  Clock,
  Timer,
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

const reservationsStore = useReservationsStore()
const authStore = useAuthStore()
const cancellingId = ref(null)

onMounted(async () => {
  reservationsStore.clearActionError()
  reservationsStore.clearActionSuccess()

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

const isPast = (reservation) => {
  const end = getReservationEnd(reservation)

  return end ? end.getTime() < Date.now() : false
}

const reservations = computed(() => {
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source
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
    ? 'Aun no hay reservas registradas.'
    : 'Aun no tienes reservas registradas.'
})

const statusLabel = (status) => {
  switch (status) {
    case 'CONFIRMED':
      return 'Confirmada'

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

const canCancel = (reservation) => {
  return (
    reservation.status !== 'CANCELLED' &&
    !isPast(reservation)
  )
}

const cancelReservation = async (reservation) => {
  const confirmed = window.confirm(
    'Deseas cancelar esta reserva?'
  )

  if (!confirmed) {
    return
  }

  cancellingId.value = reservation.id

  try {
    await reservationsStore.cancelReservation(
      reservation.id
    )

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store deja el mensaje listo para mostrar en la vista.
  } finally {
    cancellingId.value = null
  }
}
</script>

<template>
  <main class="reservations-view">

    <header class="page-header">

      <h1>
        {{ authStore.user?.isAdmin ? 'Reservas' : 'Mis Reservas' }}
      </h1>

      <p>
        {{ authStore.user?.isAdmin
          ? 'Revisa todas las reservas registradas en el sistema.'
          : 'Revisa tus reservas registradas en el sistema.' }}
      </p>

    </header>

    <div
      v-if="isLoading"
      aria-label="Cargando reservas"
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

    <section
      v-else
      class="content"
    >

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

      <div
        v-if="!reservations.length"
        class="state-card"
      >
        {{ emptyMessage }}
      </div>

      <section
        v-else
        class="reservations-list"
      >

        <article
          v-for="reservation in reservations"
          :key="reservation.id"
          class="reservation-card"
        >

          <div class="card-header">

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

            <div class="card-actions">

              <RouterLink
                class="detail-link app-button secondary"
                :to="`/reservations/${reservation.id}`"
              >
                Detalle
              </RouterLink>

              <button
                v-if="canCancel(reservation)"
                class="cancel-button app-button danger"
                type="button"
                :disabled="cancellingId === reservation.id"
                @click="cancelReservation(reservation)"
              >
                <XCircle :size="18" />
                Cancelar
              </button>

            </div>

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

    </section>

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

.reservation-card {
  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow: var(--shadow-card);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.status {
  display: inline-flex;

  padding: 6px 10px;

  border-radius: var(--radius-pill);

  background: var(--color-primary-soft);
  color: var(--color-primary-strong);

  font-size: 12px;
  font-weight: 800;
}

.status.cancelled {
  background: var(--color-error-soft);

  color: var(--color-error);
}

.status.pending {
  background: var(--color-warning-soft);

  color: var(--color-warning);
}

.reservation-card h2 {
  margin: 12px 0 0;

  color: var(--color-text);

  font-size: 20px;
  font-weight: 800;
}

.reservation-card p {
  margin: 6px 0 0;

  color: var(--color-text-muted);
}

.card-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;

  gap: 10px;
}

.detail-link,
.cancel-button {
  min-height: 40px;
  padding: 10px 14px;

  white-space: nowrap;
}

.detail-link {
  text-decoration: none;
}

.cancel-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
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

  background: var(--color-surface-muted);

  border-radius: var(--radius-pill);

  padding: 8px 11px;

  color: #475569;

  font-size: 13px;
  font-weight: 700;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }

  .card-header {
    flex-direction: column;
  }

  .card-actions,
  .detail-link,
  .cancel-button {
    width: 100%;
  }
}
</style>
