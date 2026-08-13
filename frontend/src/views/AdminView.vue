<script setup>
import { computed, onMounted } from 'vue'

import {
  BarChart3,
  CalendarCheck,
  CalendarDays,
  LayoutGrid,
  Users
} from '@lucide/vue'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import MetricCard from '@/components/admin/MetricCard.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getBusinessDateKey,
  getReservationDateKey,
  parseReservationDateTime
} from '@/utils/reservationTime'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()

onMounted(async () => {
  await Promise.all([
    resourcesStore.fetchResources(),
    reservationsStore.fetchReservations()
  ])
})

const todayKey = computed(() => {
  return getBusinessDateKey()
})

const activeResources = computed(() => {
  return resourcesStore.resources.filter(
    (resource) => resource.isActive !== false
  )
})

const confirmedReservations = computed(() => {
  return reservationsStore.reservations.filter(
    (reservation) => reservation.status === 'CONFIRMED'
  )
})

const todayReservations = computed(() => {
  return reservationsStore.reservations.filter(
    (reservation) =>
      getReservationDateKey(reservation.startTime) ===
      todayKey.value
  )
})

const upcomingReservations = computed(() => {
  return reservationsStore.reservations
    .filter((reservation) => {
      if (reservation.status === 'CANCELLED') {
        return false
      }

      const start =
        parseReservationDateTime(reservation.startTime)

      return start ? start.getTime() >= Date.now() : false
    })
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
    .slice(0, 5)
})

const stats = computed(() => [
  {
    label: 'Recursos activos',
    value: activeResources.value.length,
    detail: `${resourcesStore.resources.length} registrados`,
    icon: LayoutGrid,
    variant: 'blue'
  },
  {
    label: 'Reservas confirmadas',
    value: confirmedReservations.value.length,
    detail: `${reservationsStore.reservations.length} totales`,
    icon: CalendarCheck,
    variant: 'green'
  },
  {
    label: 'Reservas de hoy',
    value: todayReservations.value.length,
    detail: 'Agenda diaria',
    icon: CalendarDays,
    variant: 'amber'
  }
])

const actions = [
  {
    label: 'Recursos',
    detail: 'Revisar instalaciones',
    to: '/resources',
    icon: LayoutGrid
  },
  {
    label: 'Usuarios',
    detail: 'Gestion de cuentas',
    to: '/users',
    icon: Users
  },
  {
    label: 'Reportes',
    detail: 'Indicadores del sistema',
    to: '/reports',
    icon: BarChart3
  }
]

</script>

<template>
  <main class="admin-view">

    <header class="page-header">

      <h1>
        Administracion
      </h1>

      <p>
        Resumen operativo con datos actuales de la base.
      </p>

    </header>

    <div
      v-if="resourcesStore.loading || reservationsStore.loading"
      aria-label="Cargando panel administrativo"
      role="status"
      aria-live="polite"
    >
      <SkeletonLoader
        variant="metrics-table"
        :items="5"
      />
    </div>

    <section
      v-else
      class="content"
    >

      <div
        v-if="resourcesStore.error"
        class="state-card error"
      >
        {{ resourcesStore.error }}
      </div>

      <div
        v-if="reservationsStore.loadingError"
        class="state-card error"
      >
        {{ reservationsStore.loadingError }}
      </div>

      <section class="stats-grid">

        <MetricCard
          v-for="stat in stats"
          :key="stat.label"
          :title="stat.label"
          :value="stat.value"
          :subtitle="stat.detail"
          :icon="stat.icon"
        />

      </section>

      <section class="admin-grid">

        <article class="panel">

          <header>

            <h2>
              Accesos
            </h2>

            <p>
              Secciones administrativas.
            </p>

          </header>

          <div class="actions">

            <RouterLink
              v-for="action in actions"
              :key="action.label"
              class="action-link"
              :to="action.to"
            >

              <component
                :is="action.icon"
                :size="20"
              />

              <span>
                <strong>
                  {{ action.label }}
                </strong>

                <small>
                  {{ action.detail }}
                </small>
              </span>

            </RouterLink>

          </div>

        </article>

        <article class="panel">

          <header>

            <h2>
              Próximas reservas
            </h2>

            <p>
              Primeras reservas activas en agenda.
            </p>

          </header>

          <div
            v-if="!upcomingReservations.length"
            class="empty-state"
          >
            No hay reservas proximas.
          </div>

          <div
            v-else
            class="reservation-list"
          >

            <article
              v-for="reservation in upcomingReservations"
              :key="reservation.id"
              class="reservation-row"
            >

              <div>

                <strong>
                  {{ reservation.resourceName || 'Recurso' }}
                </strong>

                <span>
                  {{ reservation.title || 'Reserva' }}
                </span>

              </div>

              <p>
                {{ formatReservationDate(reservation.startTime) }}
                ·
                {{ formatReservationTimeRange(
                  reservation.startTime,
                  reservation.durationMinutes
                ) }}
              </p>

              <StatusBadge :status="reservation.status" />

            </article>

          </div>

        </article>

      </section>

    </section>

  </main>
</template>

<style scoped>
.admin-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  color: #0f172a;

  font-size: 30px;
  font-weight: 800;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.content {
  display: flex;
  flex-direction: column;

  gap: 18px;
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

.stats-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));

  gap: 18px;
}

.stat-card,
.panel {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.stat-card {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 9px;
}

.stat-card svg {
  color: #1d4ed8;
}

.stat-card.green svg {
  color: #15803d;
}

.stat-card.amber svg {
  color: #b45309;
}

.stat-card span {
  color: #64748b;

  font-size: 13px;
  font-weight: 800;
}

.stat-card strong {
  color: #0f172a;

  font-size: 34px;
  font-weight: 900;
}

.stat-card p {
  margin: 0;

  color: #475569;
}

.admin-grid {
  display: grid;

  grid-template-columns:
    minmax(240px, 340px)
    minmax(0, 1fr);

  gap: 18px;
}

.panel {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;
}

.panel header h2 {
  margin: 0;

  color: #0f172a;

  font-size: 20px;
  font-weight: 800;
}

.panel header p {
  margin: 6px 0 0;

  color: #64748b;
}

.actions,
.reservation-list {
  display: flex;
  flex-direction: column;

  gap: 10px;
}

.action-link {
  border: 1px solid #e2e8f0;
  border-radius: 14px;

  color: #0f172a;
  text-decoration: none;

  display: flex;
  align-items: center;
  gap: 12px;

  padding: 14px;

  transition: 0.2s;
}

.action-link:hover {
  border-color: #bfdbfe;

  background: #eff6ff;
}

.action-link svg {
  color: #1d4ed8;

  flex-shrink: 0;
}

.action-link span {
  display: flex;
  flex-direction: column;

  gap: 2px;
}

.action-link small {
  color: #64748b;
}

.empty-state {
  color: #64748b;

  font-weight: 700;
}

.reservation-row {
  border: 1px solid #e2e8f0;
  border-radius: 14px;

  padding: 14px;

  display: grid;

  grid-template-columns:
    minmax(0, 1fr)
    minmax(180px, 260px)
    auto;

  align-items: center;
  gap: 14px;
}

.reservation-row div {
  min-width: 0;
}

.reservation-row strong,
.reservation-row span {
  display: block;

  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reservation-row strong {
  color: #0f172a;
}

.reservation-row span,
.reservation-row p {
  color: #64748b;
}

.reservation-row p {
  margin: 0;

  font-size: 13px;
}

.reservation-row small {
  border-radius: 999px;

  background: #eff6ff;

  color: #1d4ed8;

  padding: 7px 10px;

  font-weight: 800;
}

@media (max-width: 900px) {
  .admin-grid {
    grid-template-columns: 1fr;
  }

  .reservation-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }
}
</style>
