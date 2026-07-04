<script setup>
import { computed, onMounted } from 'vue'

import {
  BarChart3,
  Clock,
  LayoutGrid
} from 'lucide-vue-next'

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import { parseReservationDateTime } from '@/utils/reservationTime'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()

onMounted(async () => {
  await Promise.all([
    resourcesStore.fetchResources(),
    reservationsStore.fetchReservations()
  ])
})

const activeReservations = computed(() => {
  return reservationsStore.reservations.filter(
    (reservation) => reservation.status !== 'CANCELLED'
  )
})

const resourceUsage = computed(() => {
  const usage = new Map()

  for (const resource of resourcesStore.resources) {
    usage.set(resource.id, {
      id: resource.id,
      name: resource.name,
      count: 0,
      minutes: 0
    })
  }

  for (const reservation of activeReservations.value) {
    const current = usage.get(reservation.resourceId) || {
      id: reservation.resourceId,
      name: reservation.resourceName || 'Recurso',
      count: 0,
      minutes: 0
    }

    current.count += 1
    current.minutes += reservation.durationMinutes || 0

    usage.set(reservation.resourceId, current)
  }

  return Array.from(usage.values())
    .filter((item) => item.count > 0)
    .sort((first, second) => second.count - first.count)
    .slice(0, 8)
})

const statusSummary = computed(() => {
  const labels = {
    CONFIRMED: 'Confirmadas',
    PENDING: 'Pendientes',
    CANCELLED: 'Canceladas'
  }

  const summary = new Map()

  for (const reservation of reservationsStore.reservations) {
    const key = reservation.status || 'OTHER'

    summary.set(key, (summary.get(key) || 0) + 1)
  }

  return Array.from(summary.entries()).map(([status, count]) => ({
    label: labels[status] || status,
    count
  }))
})

const peakHours = computed(() => {
  const summary = new Map()

  for (const reservation of activeReservations.value) {
    const start =
      parseReservationDateTime(reservation.startTime)

    if (!start) {
      continue
    }

    const hour = `${String(start.getHours()).padStart(2, '0')}:00`

    summary.set(hour, (summary.get(hour) || 0) + 1)
  }

  return Array.from(summary.entries())
    .map(([hour, count]) => ({ hour, count }))
    .sort((first, second) => second.count - first.count)
    .slice(0, 6)
})

const totalReservedHours = computed(() => {
  const minutes = activeReservations.value.reduce(
    (total, reservation) =>
      total + (reservation.durationMinutes || 0),
    0
  )

  return Math.round(minutes / 60)
})

const stats = computed(() => [
  {
    label: 'Reservas activas',
    value: activeReservations.value.length,
    icon: BarChart3
  },
  {
    label: 'Horas reservadas',
    value: totalReservedHours.value,
    icon: Clock
  },
  {
    label: 'Recursos con uso',
    value: resourceUsage.value.length,
    icon: LayoutGrid
  }
])
</script>

<template>
  <main class="reports-view">

    <header class="page-header">

      <h1>
        Reportes
      </h1>

      <p>
        Indicadores calculados desde reservas y recursos actuales.
      </p>

    </header>

    <div
      v-if="resourcesStore.loading || reservationsStore.loading"
      class="state-card"
    >
      Cargando reportes...
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
        >

          <component
            :is="stat.icon"
            :size="25"
          />

          <span>
            {{ stat.label }}
          </span>

          <strong>
            {{ stat.value }}
          </strong>

        </article>

      </section>

      <section class="reports-grid">

        <article class="report-panel">

          <header>

            <h2>
              Uso por recurso
            </h2>

            <p>
              Recursos con reservas activas.
            </p>

          </header>

          <div
            v-if="!resourceUsage.length"
            class="empty-state"
          >
            No hay uso registrado.
          </div>

          <div
            v-else
            class="rows"
          >

            <div
              v-for="item in resourceUsage"
              :key="item.id"
              class="report-row"
            >

              <span>
                {{ item.name }}
              </span>

              <strong>
                {{ item.count }} reservas
              </strong>

              <small>
                {{ Math.round(item.minutes / 60) }} h
              </small>

            </div>

          </div>

        </article>

        <article class="report-panel">

          <header>

            <h2>
              Estados
            </h2>

            <p>
              Distribucion de reservas.
            </p>

          </header>

          <div
            v-if="!statusSummary.length"
            class="empty-state"
          >
            No hay reservas registradas.
          </div>

          <div
            v-else
            class="rows"
          >

            <div
              v-for="item in statusSummary"
              :key="item.label"
              class="report-row"
            >

              <span>
                {{ item.label }}
              </span>

              <strong>
                {{ item.count }}
              </strong>

            </div>

          </div>

        </article>

        <article class="report-panel">

          <header>

            <h2>
              Horas punta
            </h2>

            <p>
              Horarios con mayor uso.
            </p>

          </header>

          <div
            v-if="!peakHours.length"
            class="empty-state"
          >
            No hay horas registradas.
          </div>

          <div
            v-else
            class="rows"
          >

            <div
              v-for="item in peakHours"
              :key="item.hour"
              class="report-row"
            >

              <span>
                {{ item.hour }}
              </span>

              <strong>
                {{ item.count }} reservas
              </strong>

            </div>

          </div>

        </article>

      </section>

    </section>

  </main>
</template>

<style scoped>
.reports-view {
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

.stats-grid,
.reports-grid {
  display: grid;

  gap: 18px;
}

.stats-grid {
  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));
}

.reports-grid {
  grid-template-columns:
    repeat(auto-fit, minmax(280px, 1fr));
}

.stat-card,
.report-panel {
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

.report-panel {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;
}

.report-panel header h2 {
  margin: 0;

  color: #0f172a;

  font-size: 20px;
  font-weight: 800;
}

.report-panel header p {
  margin: 6px 0 0;

  color: #64748b;
}

.rows {
  display: flex;
  flex-direction: column;

  gap: 10px;
}

.report-row {
  border: 1px solid #e2e8f0;
  border-radius: 14px;

  display: grid;
  grid-template-columns:
    minmax(0, 1fr)
    auto
    auto;

  align-items: center;
  gap: 12px;

  padding: 13px 14px;
}

.report-row span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  color: #0f172a;

  font-weight: 800;
}

.report-row strong {
  color: #1d4ed8;

  white-space: nowrap;
}

.report-row small {
  color: #64748b;

  font-weight: 800;
}

.empty-state {
  color: #64748b;

  font-weight: 700;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }

  .report-row {
    grid-template-columns: 1fr;
  }
}
</style>
