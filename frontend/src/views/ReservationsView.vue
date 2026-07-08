<script setup>
import { computed, onMounted, ref } from 'vue'

import ReservationListCard from '@/components/reservations/ReservationListCard.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  isReservationActionable,
  isReservationCancelable,
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

const reservations = computed(() => {
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source
    .filter((reservation) => {
      if (authStore.user?.isAdmin) {
        return true
      }

      return isReservationActionable(reservation)
    })
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
    ? 'Aún no hay reservas registradas.'
    : 'Aún no tienes reservas registradas.'
})

const canCancel = (reservation) => {
  return isReservationCancelable(reservation)
}

const cancelReservation = async (reservation) => {
  const confirmed = window.confirm(
    '¿Deseas cancelar esta reserva?'
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

        <ReservationListCard
          v-for="reservation in reservations"
          :key="reservation.id"
          :reservation="reservation"
          :detail-to="`/reservations/${reservation.id}`"
          :show-cancel="canCancel(reservation)"
          :cancel-disabled="cancellingId === reservation.id"
          @cancel="cancelReservation"
        />

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

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }
}
</style>
