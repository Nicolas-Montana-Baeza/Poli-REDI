<script setup>
import { computed, onMounted, ref } from 'vue'

import {
  CalendarDays,
  Clock,
  Timer,
  XCircle
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  parseReservationDateTime
} from '@/utils/reservationTime'

const reservationsStore = useReservationsStore()
const cancellingId = ref(null)

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

const isPast = (reservation) => {
  const end = getReservationEnd(reservation)

  return end ? end.getTime() < Date.now() : false
}

const reservations = computed(() => {
  return reservationsStore.myReservations
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
        Mis Reservas
      </h1>

      <p>
        Revisa tus reservas registradas en el sistema.
      </p>

    </header>

    <div
      v-if="reservationsStore.myLoading"
      aria-label="Cargando reservas"
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
        Aun no tienes reservas registradas.
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
                class="detail-link"
                :to="`/reservations/${reservation.id}`"
              >
                Detalle
              </RouterLink>

              <button
                v-if="canCancel(reservation)"
                class="cancel-button"
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

  color: #0f172a;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.content,
.reservations-list {
  display: flex;
  flex-direction: column;

  gap: 14px;
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

.state-card.success {
  background: #dcfce7;

  color: #166534;

  border-color: #bbf7d0;
}

.reservation-card {
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

.card-header {
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

.reservation-card h2 {
  margin: 12px 0 0;

  color: #0f172a;

  font-size: 20px;
  font-weight: 800;
}

.reservation-card p {
  margin: 6px 0 0;

  color: #64748b;
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
  background: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 14px;

  color: #b91c1c;

  cursor: pointer;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  padding: 10px 14px;

  font-weight: 800;

  white-space: nowrap;

  transition: 0.2s;
}

.detail-link {
  background: #eff6ff;

  border-color: #bfdbfe;

  color: #1d4ed8;

  text-decoration: none;
}

.detail-link:hover {
  background: #dbeafe;
}

.cancel-button:hover:not(:disabled) {
  background: #fecaca;
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
