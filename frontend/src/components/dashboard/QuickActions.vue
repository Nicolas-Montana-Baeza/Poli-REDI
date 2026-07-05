<script setup>
import {
  Plus,
  CalendarDays,
  ListChecks,
  Clock3
} from 'lucide-vue-next'

const actions = [
  {
    id: 1,
    title: 'Nueva Reserva',
    description: 'Reserva una cancha o sala',
    icon: Plus,
    variant: 'primary',
    to: '/availability'
  },
  {
    id: 2,
    title: 'Disponibilidad',
    description: 'Ver horarios libres',
    icon: CalendarDays,
    variant: 'secondary',
    to: '/availability'
  },
  {
    id: 3,
    title: 'Mis Reservas',
    description: 'Ver y cancelar reservas',
    icon: ListChecks,
    variant: 'warning',
    to: '/reservations'
  },
  {
    id: 4,
    title: 'Historial',
    description: 'Revisar reservas pasadas',
    icon: Clock3,
    variant: 'neutral',
    to: '/history'
  }
]
</script>

<template>
  <section class="quick-actions">

    <!-- HEADER -->
    <div class="header">

      <div>

        <h2>
          Acciones rápidas
        </h2>

        <p>
          Accede rápidamente a las funciones principales.
        </p>

      </div>

    </div>

    <!-- ACTIONS -->
    <div class="actions-grid">

      <RouterLink
        v-for="action in actions"
        :key="action.id"
        class="action-card"
        :class="action.variant"
        :to="action.to"
      >

        <!-- Icon -->
        <div class="icon">

          <component
            :is="action.icon"
            :size="22"
          />

        </div>

        <!-- Text -->
        <div class="content">

          <h3>
            {{ action.title }}
          </h3>

          <p>
            {{ action.description }}
          </p>

        </div>

      </RouterLink>

    </div>

  </section>
</template>

<style scoped>
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* HEADER */
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

/* GRID */
.actions-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));

  gap: 18px;
}

/* CARD */
.action-card {
  border: none;

  border-radius: 22px;

  padding: 20px;

  display: flex;
  align-items: center;
  gap: 16px;

  cursor: pointer;

  transition: 0.25s;

  text-align: left;

  text-decoration: none;
}

.action-card:hover {
  transform: translateY(-4px);

  box-shadow:
    0 12px 28px rgba(0,0,0,0.08);
}

/* ICON */
.icon {
  width: 52px;
  height: 52px;

  border-radius: 16px;

  display: flex;
  align-items: center;
  justify-content: center;

  flex-shrink: 0;
}

/* CONTENT */
.content {
  display: flex;
  flex-direction: column;
}

.content h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 700;
}

.content p {
  margin-top: 4px;

  font-size: 13px;
  line-height: 1.4;
}

/* VARIANTS */
.primary {
  background: linear-gradient(
    135deg,
    #2563eb,
    #1d4ed8
  );

  color: white;
}

.primary .icon {
  background: rgba(255,255,255,0.18);
}

.primary p {
  color: rgba(255,255,255,0.8);
}

/* Secondary */
.secondary {
  background: white;

  border: 1px solid #dbeafe;
}

.secondary .icon {
  background: #dbeafe;
  color: #2563eb;
}

.secondary p {
  color: #64748b;
}

/* Warning */
.warning {
  background: linear-gradient(
    135deg,
    #f97316,
    #ea580c
  );

  color: white;
}

.warning .icon {
  background: rgba(255,255,255,0.18);
}

.warning p {
  color: rgba(255,255,255,0.82);
}

/* Neutral */
.neutral {
  background: white;

  border: 1px solid #e2e8f0;
}

.neutral .icon {
  background: #f1f5f9;
  color: #475569;
}

.neutral p {
  color: #64748b;
}

/* MOBILE */
@media (max-width: 768px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }

  .action-card {
    padding: 18px;
  }
}
</style>
