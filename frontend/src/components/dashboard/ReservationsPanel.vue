<script setup>
import ReservationCard from './ReservationCard.vue'

defineProps({
  reservations: {
    type: Array,
    default: () => []
  }
})
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

      <button class="view-all">
        Ver todas
      </button>

    </div>

    <!-- EMPTY -->
    <div
      v-if="!reservations.length"
      class="empty"
    >

      <h3>
        No tienes reservas
      </h3>

      <p>
        Cuando reserves una instalacion aparecera aqui.
      </p>

    </div>

    <!-- LIST -->
    <div
      v-else
      class="reservations-grid"
    >

      <ReservationCard
        v-for="reservation in reservations"
        :key="reservation.id"
        :title="reservation.title"
        :sport="reservation.sport"
        :date="reservation.date"
        :time="reservation.time"
        :status="reservation.status"
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
  border: none;

  background: #1e3a8a;
  color: white;

  padding: 12px 18px;

  border-radius: 14px;

  font-size: 14px;
  font-weight: 600;

  cursor: pointer;

  transition: 0.2s;
}

.view-all:hover {
  background: #f97316;
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
  background: white;

  border-radius: 22px;

  padding: 40px;

  text-align: center;

  border: 1px solid #e2e8f0;
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
