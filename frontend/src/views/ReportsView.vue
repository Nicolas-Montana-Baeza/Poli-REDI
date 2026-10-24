<script setup>
import { computed, onMounted, ref } from 'vue'
import { BarChart3, Clock, LayoutGrid } from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import {
  calculateReportMetrics,
  formatReservedMinutes
} from '@/utils/reportMetrics'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()
const refreshedAt = ref(Date.now())

onMounted(async () => {
  await Promise.all([
    resourcesStore.fetchResources(),
    reservationsStore.fetchReservations()
  ])
  refreshedAt.value = Date.now()
})

const report = computed(() => calculateReportMetrics(
  reservationsStore.reservations,
  resourcesStore.resources,
  refreshedAt.value
))

const resourceUsage = computed(() => report.value.resourceUsage)
const statusSummary = computed(() => report.value.statusSummary)
const peakHours = computed(() => report.value.peakHours)

const stats = computed(() => [
  {
    label: 'Reservas vigentes',
    value: report.value.activeCount,
    icon: BarChart3
  },
  {
    label: 'Tiempo reservado vigente',
    value: formatReservedMinutes(report.value.totalMinutes),
    icon: Clock
  },
  {
    label: 'Recursos con reservas',
    value: report.value.resourceCount,
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
        Indicadores de reservas vigentes y distribución de estados.
      </p>

    </header>

    <div
      v-if="resourcesStore.loading || reservationsStore.loading"
      aria-label="Cargando reportes"
    >
      <SkeletonLoader
        variant="resources"
        :items="6"
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

      <section v-if="!resourcesStore.error && !reservationsStore.loadingError" class="stats-grid">

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

      <section v-if="!resourcesStore.error && !reservationsStore.loadingError" class="reports-grid">

        <article class="report-panel">

          <header>

            <h2>
              Reservas vigentes por recurso
            </h2>

            <p>
              Reservas pendientes o confirmadas que aún no finalizan.
            </p>

          </header>

          <div
            v-if="!resourceUsage.length"
            class="empty-state"
          >
            No hay reservas vigentes.
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
                {{ formatReservedMinutes(item.minutes) }}
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
              Distribución de estados de las reservas consultadas.
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
              Horarios de inicio más frecuentes
            </h2>

            <p>
              Horas de inicio de reservas vigentes; no mide ocupación por franja.
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
