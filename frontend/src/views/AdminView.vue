<script setup>
import { computed, onMounted } from 'vue'

import {
  BarChart3,
  CalendarCheck,
  CalendarDays,
  LayoutGrid,
  SlidersHorizontal,
  Users
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
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
    label: 'Políticas de reserva',
    detail: 'Configuración institucional',
    to: '/admin/reservation-policies',
    icon: SlidersHorizontal
  },
  {
    label: 'Reportes',
    detail: 'Indicadores del sistema',
    to: '/reports',
    icon: BarChart3
  }
]

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
    >
      <SkeletonLoader
        variant="resources"
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

        <article
          v-for="stat in stats"
          :key="stat.label"
          class="stat-card"
          :class="stat.variant"
        >

          <component
            :is="stat.icon"
            :size="26"
          />

          <span>
            {{ stat.label }}
          </span>

          <strong>
            {{ stat.value }}
          </strong>

          <p>
            {{ stat.detail }}
          </p>

        </article>

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

              <small>
                {{ statusLabel(reservation.status) }}
              </small>

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
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.page-header h1 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-page-title);
  font-weight: 800;
  line-height: 1.15;
}

.page-header p {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  line-height: 1.45;
}

.content {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr));
  gap: var(--space-4);
}

.stat-card,
.panel {
  min-width: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-5);
}

.stat-card svg { color: var(--color-primary-strong); }
.stat-card.green svg { color: var(--color-success); }
.stat-card.amber svg { color: var(--color-warning); }

.stat-card span {
  color: var(--color-text-muted);
  font-size: var(--text-help);
  font-weight: 700;
}

.stat-card strong {
  color: var(--color-text);
  font-size: 34px;
  font-weight: 800;
}

.stat-card p {
  margin: 0;
  color: var(--color-text-muted);
}

.admin-grid {
  display: grid;
  grid-template-columns: minmax(220px, 340px) minmax(0, 1fr);
  gap: var(--space-4);
}

.panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-5);
}

.panel header h2 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-section-title);
  font-weight: 800;
}

.panel header p {
  margin: var(--space-2) 0 0;
  color: var(--color-text-muted);
}

.actions,
.reservation-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.action-link {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text);
  text-decoration: none;
  transition: background-color 0.18s ease, border-color 0.18s ease;
}

.action-link:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
}

.action-link:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.action-link svg {
  flex-shrink: 0;
  color: var(--color-primary-strong);
}

.action-link span {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: var(--space-1);
}

.action-link small,
.empty-state {
  color: var(--color-text-muted);
}

.empty-state { font-weight: 700; }

.reservation-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(180px, 260px) auto;
  align-items: center;
  min-width: 0;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-md);
}

.reservation-row div { min-width: 0; }

.reservation-row strong,
.reservation-row span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reservation-row strong { color: var(--color-text); }

.reservation-row span,
.reservation-row p {
  color: var(--color-text-muted);
}

.reservation-row p {
  margin: 0;
  font-size: var(--text-help);
}

.reservation-row small {
  width: fit-content;
  padding: 6px var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
  font-weight: 750;
}

@media (max-width: 900px) {
  .admin-grid { grid-template-columns: 1fr; }
  .reservation-row { grid-template-columns: 1fr; }
}

@media (max-width: 480px) {
  .panel,
  .stat-card { padding: var(--space-4); }
}
</style>
