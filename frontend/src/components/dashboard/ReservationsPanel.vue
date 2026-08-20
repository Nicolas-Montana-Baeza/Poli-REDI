<script setup>
import ReservationListCard from '@/components/reservations/ReservationListCard.vue'

defineProps({
  reservations: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits([
  'open-detail'
])
</script>

<template>
  <section class="panel">

    <!-- HEADER -->
    <div class="header">

      <div>

        <h2>
          Mis Reservas
        </h2>

        <p>
          Gestiona tus reservas activas y revisa su estado.
        </p>

      </div>

      <RouterLink
        class="view-all app-button primary"
        to="/reservations"
      >
        Ver todas
      </RouterLink>

    </div>

    <!-- EMPTY -->
    <div
      v-if="!reservations.length"
      class="empty app-card"
    >

      <h3>
        No tienes reservas
      </h3>

      <p>
        Cuando reserves una instalación aparecerá aquí.
      </p>

    </div>

    <!-- LIST -->
    <div
      v-else
      class="reservations-grid"
    >

      <ReservationListCard
        v-for="reservation in reservations"
        :key="reservation.id"
        :reservation="reservation"
        @open-detail="
          emit('open-detail', $event)
        "
      />

    </div>

  </section>
</template>

<style scoped>
.panel {
  width: 100%;

  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* HEADER */
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;
}

.header h2 {
  margin: 0;

  font-size: 24px;
  font-weight: 700;

  color: #0f172a;
}

.header p {
  margin-top: 4px;

  font-size: 14px;

  color: #64748b;
}

/* Button */
.view-all {
  text-decoration: none;
}

/* GRID */
.reservations-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(320px, 1fr));

  gap: 20px;
}

/* EMPTY */
.empty {
  padding: 40px;

  text-align: center;
}

.empty h3 {
  margin: 0;

  font-size: 20px;

  color: #0f172a;
}

.empty p {
  margin-top: 8px;

  color: #64748b;
}

/* Mobile */
@media (max-width: 768px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .view-all {
    width: 100%;
  }

  .reservations-grid {
    grid-template-columns: 1fr;
  }
}
</style>
