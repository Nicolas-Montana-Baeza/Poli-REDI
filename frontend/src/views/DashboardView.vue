<script setup>
import { computed, onMounted, ref } from 'vue'
import {
  CalendarDays,
  Clock,
  KeyRound,
  Search
} from 'lucide-vue-next'

import FacilityCarousel from '@/components/dashboard/FacilityCarousel.vue'
import AsyncRegion from '@/components/ui/AsyncRegion.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { useAuthStore } from '@/stores/auth'
import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus,
  isReservationActionable,
  parseReservationDateTime
} from '@/utils/reservationTime'

const authStore = useAuthStore()
const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()
const search = ref('')
const typeFilter = ref('ALL')

onMounted(async () => {
  await Promise.all([
    authStore.loadAuthUser(),
    resourcesStore.fetchResources(),
    reservationsStore.fetchMyReservations()
  ])
})

const firstName = computed(() => {
  const fullName =
    authStore.user?.fullName ||
    authStore.account?.name ||
    'Usuario'

  return String(fullName).trim().split(/\s+/)[0] || 'Usuario'
})

const facilities = computed(() =>
  resourcesStore.resources.map((resource) => ({
    ...resource,
    image: resource.imageUrl || ''
  }))
)

const typeOptions = computed(() =>
  [...new Set(
    facilities.value
      .map((resource) => resource.type)
      .filter(Boolean)
  )].sort((first, second) => first.localeCompare(second, 'es'))
)

const filteredFacilities = computed(() => {
  const query = search.value.trim().toLocaleLowerCase('es')

  return facilities.value.filter((resource) => {
    const matchesQuery =
      !query ||
      `${resource.name || ''} ${resource.type || ''}`
        .toLocaleLowerCase('es')
        .includes(query)
    const matchesType =
      typeFilter.value === 'ALL' ||
      resource.type === typeFilter.value

    return matchesQuery && matchesType
  })
})

const nextReservation = computed(() =>
  reservationsStore.myReservations
    .filter(isReservationActionable)
    .slice()
    .sort((first, second) => {
      const firstDate = parseReservationDateTime(first.startTime)
      const secondDate = parseReservationDateTime(second.startTime)

      return (
        (firstDate?.getTime() || Number.MAX_SAFE_INTEGER) -
        (secondDate?.getTime() || Number.MAX_SAFE_INTEGER)
      )
    })[0] || null
)

const nextReservationStatus = computed(() =>
  nextReservation.value
    ? getReservationDisplayStatus(nextReservation.value)
    : null
)
</script>

<template>
  <main class="dashboard">
    <section class="dashboard-summary">
      <div class="hero">
        <div>
          <h1>Hola, {{ firstName }}</h1>
          <p>¿Qué actividad quieres realizar hoy?</p>
        </div>

        <div class="hero-actions">
          <RouterLink class="app-button hero-primary" to="/availability">
            <CalendarDays :size="18" />
            Reservar instalación
          </RouterLink>

          <RouterLink class="app-button hero-secondary" to="/join">
            <KeyRound :size="18" />
            Ingresar código
          </RouterLink>
        </div>
      </div>

      <aside class="next-reservation app-card" aria-labelledby="next-reservation-title">
        <div class="next-header">
          <h2 id="next-reservation-title">Tu próxima reserva</h2>
          <RouterLink class="all-reservations" to="/history">
            Ver todas mis reservas
          </RouterLink>
        </div>

        <AsyncRegion
          :loading="reservationsStore.myLoading"
          :error="reservationsStore.myLoadingError"
          :empty="!nextReservation"
          loading-label="Cargando próxima reserva"
          skeleton-variant="compact-rows"
          :skeleton-items="2"
        >
          <template #error="{ error }">
            <div class="compact-state error" role="alert">{{ error }}</div>
          </template>

          <template #empty>
            <div class="compact-state">
              <h3>Aún no tienes reservas próximas</h3>
              <p>Explora las instalaciones y agenda tu próxima actividad.</p>
            </div>
          </template>

          <div class="next-content">
            <StatusBadge
              :status="nextReservation.status"
              :label="nextReservationStatus.label"
            />

            <div>
              <h3>{{ nextReservation.title || 'Reserva' }}</h3>
              <p>{{ nextReservation.resourceName || 'Recurso' }}</p>
            </div>

            <div class="next-meta">
              <span>
                <CalendarDays :size="16" />
                {{ formatReservationDate(nextReservation.startTime) }}
              </span>
              <span>
                <Clock :size="16" />
                {{ formatReservationTimeRange(
                  nextReservation.startTime,
                  nextReservation.durationMinutes
                ) }}
              </span>
            </div>

            <RouterLink
              class="app-button secondary detail-link"
              :to="`/reservations/${nextReservation.id}`"
            >
              Ver detalle
            </RouterLink>
          </div>
        </AsyncRegion>
      </aside>
    </section>

    <section class="facilities-section" aria-labelledby="facilities-title">
      <div class="facilities-heading">
        <div>
          <h2 id="facilities-title">Explora instalaciones</h2>
          <p>Encuentra el espacio ideal para tu próxima actividad.</p>
        </div>

        <div class="facility-filters">
          <label class="search-control">
            <span class="sr-only">Buscar instalación</span>
            <Search :size="17" aria-hidden="true" />
            <input
              v-model="search"
              type="search"
              placeholder="Buscar instalación"
            />
          </label>

          <label class="type-control">
            <span class="sr-only">Filtrar por tipo</span>
            <select v-model="typeFilter" aria-label="Filtrar por tipo">
              <option value="ALL">Todas</option>
              <option
                v-for="type in typeOptions"
                :key="type"
                :value="type"
              >
                {{ type }}
              </option>
            </select>
          </label>
        </div>
      </div>

      <AsyncRegion
        :loading="resourcesStore.loading"
        :error="resourcesStore.error"
        :empty="!filteredFacilities.length"
        loading-label="Cargando instalaciones"
        skeleton-variant="media-grid"
        :skeleton-items="4"
        :skeleton-columns="4"
        mobile-carousel
      >
        <template #empty>
          <div class="state-card empty">
            <h3>No encontramos instalaciones</h3>
            <p>Prueba con otro nombre o selecciona una categoría diferente.</p>
          </div>
        </template>

        <FacilityCarousel :facilities="filteredFacilities" />
      </AsyncRegion>
    </section>
  </main>
</template>

<style scoped>
.dashboard {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.dashboard-summary {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(300px, 0.95fr);
  gap: 18px;
}

.hero {
  min-height: 245px;
  padding: clamp(28px, 4vw, 44px);
  border-radius: var(--radius-xl);
  color: white;
  background:
    radial-gradient(circle at 85% 35%, rgba(96, 165, 250, 0.2), transparent 30%),
    linear-gradient(135deg, #172f78, #2563eb);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 28px;
  box-shadow: 0 14px 32px rgba(30, 64, 175, 0.16);
}

.hero h1 {
  margin: 0;
  font-size: clamp(28px, 3vw, 38px);
  font-weight: 850;
}

.hero p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, 0.88);
  font-size: 16px;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-actions a {
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-decoration: none;
}

.hero-primary {
  background: white;
  color: #1e3a8a;
  border: 1px solid white;
}

.hero-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.55);
}

.next-reservation {
  min-height: 245px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.next-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.next-header h2,
.facilities-heading h2 {
  margin: 0;
  color: var(--color-text);
  font-size: 21px;
  font-weight: 800;
}

.all-reservations {
  color: var(--color-primary);
  font-size: 12px;
  font-weight: 750;
  text-decoration: none;
  text-align: right;
}

.next-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 13px;
}

.next-content h3,
.compact-state h3,
.state-card h3 {
  margin: 0;
  color: var(--color-text);
  font-size: 18px;
}

.next-content p,
.compact-state p,
.state-card p {
  margin: 4px 0 0;
  color: var(--color-text-muted);
  font-size: 14px;
}

.next-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 16px;
  color: #475569;
  font-size: 13px;
  font-weight: 650;
}

.next-meta span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.detail-link {
  margin-top: auto;
  min-height: 38px;
  padding: 8px 12px;
  text-decoration: none;
}

.compact-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.compact-state.error {
  color: var(--color-error);
}

.facilities-section {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.facilities-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
}

.facilities-heading p {
  margin: 6px 0 0;
  color: var(--color-text-muted);
  font-size: 14px;
}

.facility-filters {
  display: flex;
  gap: 10px;
}

.search-control,
.type-control {
  min-height: 42px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  display: flex;
  align-items: center;
}

.search-control {
  min-width: min(270px, 35vw);
  padding: 0 12px;
  gap: 8px;
  color: var(--color-text-muted);
}

.search-control input,
.type-control select {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--color-text);
  font: inherit;
}

.type-control {
  min-width: 135px;
  padding: 0 10px;
}

.search-control:focus-within,
.type-control:focus-within {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-soft);
}

.state-card {
  padding: 28px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
}

.state-card.empty {
  text-align: center;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 980px) {
  .dashboard-summary {
    grid-template-columns: 1fr;
  }

  .hero,
  .next-reservation {
    min-height: auto;
  }
}

@media (max-width: 700px) {
  .dashboard {
    min-width: 0;
    gap: 22px;
  }

  .hero {
    padding: 24px 20px;
    border-radius: 18px;
    gap: 20px;
  }

  .hero h1 {
    font-size: 27px;
  }

  .hero p {
    font-size: 15px;
  }

  .hero-actions {
    width: 100%;
  }

  .hero-actions a {
    flex: 1 1 160px;
    min-height: 46px;
  }

  .next-reservation {
    padding: 20px;
    gap: 15px;
  }

  .facilities-heading,
  .facility-filters {
    align-items: stretch;
    flex-direction: column;
  }

  .facility-filters,
  .search-control,
  .type-control {
    width: 100%;
  }

  .search-control {
    min-width: 0;
  }

  .next-header {
    flex-direction: column;
  }

  .all-reservations {
    text-align: left;
  }

  .facilities-section {
    min-width: 0;
    gap: 14px;
  }
}

@media (max-width: 420px) {
  .hero-actions {
    flex-direction: column;
  }

  .hero-actions a {
    width: 100%;
    flex-basis: auto;
  }

  .next-meta {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
