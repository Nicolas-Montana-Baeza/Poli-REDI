<script setup>
import QuickActions from
'../components/dashboard/QuickActions.vue'

import FacilityCarousel from
'../components/dashboard/FacilityCarousel.vue'

import ReservationsPanel from
'../components/dashboard/ReservationsPanel.vue'

/* FACILITIES */
const facilities = [
  {
    id: 1,
    name: 'Cancha 1',
    type: 'Fútbol',
    status: 'Disponible',
    image:
      'https://images.unsplash.com/photo-1574629810360-7efbbe195018?q=80&w=1200&auto=format&fit=crop'
  },

  {
    id: 2,
    name: 'Cancha 2',
    type: 'Básquetbol',
    status: 'Ocupada',
    image:
      'https://images.unsplash.com/photo-1546519638-68e109498ffc?q=80&w=1200&auto=format&fit=crop'
  },

  {
    id: 3,
    name: 'Piscina',
    type: 'Natación',
    status: 'Disponible',
    image:
      'https://images.unsplash.com/photo-1519315901367-f34ff9154487?q=80&w=1200&auto=format&fit=crop'
  },

  {
    id: 4,
    name: 'Gimnasio',
    type: 'Entrenamiento',
    status: 'Mantenimiento',
    image:
      'https://images.unsplash.com/photo-1517836357463-d25dfeac3438?q=80&w=1200&auto=format&fit=crop'
  }
]

/* RESERVATIONS */
const reservations = [
  {
    id: 1,
    title: 'Entrenamiento Fútbol',
    facility: 'Cancha 1',
    date: '12 Mayo',
    hour: '18:00',
    status: 'Confirmada'
  },

  {
    id: 2,
    title: 'Partido Básquetbol',
    facility: 'Cancha 2',
    date: '13 Mayo',
    hour: '20:00',
    status: 'Pendiente'
  },

  {
    id: 3,
    title: 'Natación Libre',
    facility: 'Piscina',
    date: '14 Mayo',
    hour: '16:00',
    status: 'Confirmada'
  }
]

/* ACTIONS */
const handleFacilitySelect = (
  facility
) => {

  console.log(
    'Facility selected:',
    facility
  )
}

const handleReservationClick = (
  reservation
) => {

  console.log(
    'Reservation:',
    reservation
  )
}
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

      <FacilityCarousel
        :facilities="
          facilities
        "

        @select="
          handleFacilitySelect
        "
      />

    </section>

    <!-- RESERVATIONS -->
    <section class="section">

      <div class="section-header">

        <div>

          <h2>
            Próximas Reservas
          </h2>

          <p>
            Tus actividades
            agendadas.
          </p>

        </div>

      </div>

      <ReservationsPanel
        :reservations="
          reservations
        "

        @select="
          handleReservationClick
        "
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