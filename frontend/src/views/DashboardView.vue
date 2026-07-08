<script setup>
import { computed, onMounted } from 'vue'

import QuickActions from '../components/dashboard/QuickActions.vue'
import FacilityCarousel from '../components/dashboard/FacilityCarousel.vue'
import ReservationsPanel from '../components/dashboard/ReservationsPanel.vue'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'

import { useResourcesStore } from '@/stores/resources'
import { useReservationsStore } from '@/stores/reservations'
import {
  isReservationActionable,
  parseReservationDateTime
} from '@/utils/reservationTime'

const resourcesStore = useResourcesStore()
const reservationsStore = useReservationsStore()

onMounted(async () => {
  await Promise.all([
    resourcesStore.fetchResources(),
    reservationsStore.fetchMyReservations()
  ])
})

const facilities = computed(() => {
  return resourcesStore.resources.map((resource) => ({
    ...resource,
    image: ''
  }))
})

const reservations = computed(() => {
  return reservationsStore.myReservations
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
    .slice(0, 3)
})
</script>

<template>
  <main class="dashboard">

    <!-- HERO -->
    <section class="hero">

      <div>

        <h1>
          Panel Principal
        </h1>

        <p>
          Gestiona reservas,
          disponibilidad e
          instalaciones deportivas.
        </p>

      </div>

    </section>

    <!-- QUICK ACTIONS -->
    <QuickActions />

    <!-- FACILITIES -->
    <section class="section">

      <div class="section-header">

        <div>

          <h2>
            Instalaciones
          </h2>

          <p>
            Explora recursos
            disponibles.
          </p>

        </div>

      </div>

      <div
        v-if="resourcesStore.loading"
        aria-label="Cargando instalaciones"
      >
        <SkeletonLoader
          variant="dashboard"
          :items="3"
        />
      </div>

      <div
        v-else-if="resourcesStore.error"
        class="state-card error"
      >
        {{ resourcesStore.error }}
      </div>

      <FacilityCarousel
        v-else
        :facilities="facilities"
      />

    </section>

    <!-- RESERVATIONS -->
    <section class="section">

      <div class="section-header">

        <div>

          <h2>
            Proximas Reservas
          </h2>

          <p>
            Tus actividades
            agendadas.
          </p>

        </div>

      </div>

      <div
        v-if="reservationsStore.myLoading"
        aria-label="Cargando reservas"
      >
        <SkeletonLoader
          variant="reservations"
          :items="2"
        />
      </div>

      <div
        v-else-if="reservationsStore.myLoadingError"
        class="state-card error"
      >
        {{ reservationsStore.myLoadingError }}
      </div>

      <ReservationsPanel
        v-else
        :reservations="reservations"
      />

    </section>

  </main>
</template>

<style scoped>
.dashboard {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 32px;
}

/* HERO */
.hero {
  background:
    linear-gradient(
      135deg,
      #1e3a8a,
      #2563eb
    );

  border-radius: 28px;

  padding: 36px;

  color: white;

  overflow: hidden;

  position: relative;
}

.hero h1 {
  margin: 0;

  font-size: 36px;
  font-weight: 800;
}

.hero p {
  margin-top: 12px;

  max-width: 620px;

  font-size: 16px;
  line-height: 1.6;

  color: rgba(255,255,255,0.85);
}

/* SECTION */
.section {
  display: flex;
  flex-direction: column;

  gap: 18px;
}

.section-header h2 {
  margin: 0;

  font-size: 24px;
  font-weight: 800;

  color: #0f172a;
}

.section-header p {
  margin-top: 6px;

  color: #64748b;

  font-size: 14px;
}

.state-card {
  background: white;

  border-radius: 18px;

  padding: 20px;

  border: 1px solid #e2e8f0;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

/* TABLET */
@media (max-width: 1024px) {
  .hero {
    padding: 30px;
  }

  .hero h1 {
    font-size: 30px;
  }
}

/* MOBILE */
@media (max-width: 768px) {
  .dashboard {
    gap: 24px;
  }

  .hero {
    padding: 24px;

    border-radius: 22px;
  }

  .hero h1 {
    font-size: 26px;
  }

  .hero p {
    font-size: 14px;
  }

  .section-header h2 {
    font-size: 22px;
  }
}
</style>
