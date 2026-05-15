<script setup>
import { computed } from 'vue'

import {
  CalendarDays,
  Clock3,
  Users,
  MapPin,
  Trash2,
  Eye
} from 'lucide-vue-next'

const props = defineProps({
  title: {
    type: String,
    default: 'Cancha 1'
  },

  sport: {
    type: String,
    default: 'Fútbol'
  },

  date: {
    type: String,
    default: 'Lunes 12 Mayo'
  },

  time: {
    type: String,
    default: '18:00 - 19:00'
  },

  status: {
    type: String,
    default: 'pending'
  },

  participants: {
    type: Number,
    default: 7
  },

  maxParticipants: {
    type: Number,
    default: 10
  }
})

const progress = computed(() => {
  return (props.participants / props.maxParticipants) * 100
})

const statusLabel = computed(() => {
  switch (props.status) {
    case 'confirmed':
      return 'Confirmada'

    case 'cancelled':
      return 'Cancelada'

    default:
      return 'Pendiente'
  }
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'confirmed':
      return 'confirmed'

    case 'cancelled':
      return 'cancelled'

    default:
      return 'pending'
  }
})
</script>

<template>
  <div class="card">

    <!-- TOP -->
    <div class="top">

      <div class="resource">

        <div class="icon">
          <MapPin :size="18" />
        </div>

        <div>

          <h3>
            {{ title }}
          </h3>

          <p>
            {{ sport }}
          </p>

        </div>

      </div>

      <!-- STATUS -->
      <div
        class="status"
        :class="statusClass"
      >
        {{ statusLabel }}
      </div>

    </div>

    <!-- INFO -->
    <div class="info">

      <div class="info-item">

        <CalendarDays :size="16" />

        <span>
          {{ date }}
        </span>

      </div>

      <div class="info-item">

        <Clock3 :size="16" />

        <span>
          {{ time }}
        </span>

      </div>

    </div>

    <!-- PARTICIPANTS -->
    <div class="participants">

      <div class="participants-header">

        <div class="label">

          <Users :size="16" />

          <span>
            Participantes
          </span>

        </div>

        <strong>
          {{ participants }}/{{ maxParticipants }}
        </strong>

      </div>

      <!-- Progress -->
      <div class="progress">

        <div
          class="progress-fill"
          :style="{ width: `${progress}%` }"
        />

      </div>

    </div>

    <!-- ACTIONS -->
    <div class="actions">

      <button class="secondary">

        <Eye :size="16" />

        Ver

      </button>

      <button class="danger">

        <Trash2 :size="16" />

        Cancelar

      </button>

    </div>

  </div>
</template>

<style scoped>
.card {
  background: white;

  border-radius: 22px;

  padding: 20px;

  display: flex;
  flex-direction: column;
  gap: 18px;

  border: 1px solid #e2e8f0;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);

  transition: 0.2s;
}

.card:hover {
  transform: translateY(-2px);

  box-shadow:
    0 10px 24px rgba(0,0,0,0.08);
}

/* TOP */
.top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;

  gap: 16px;
}

.resource {
  display: flex;
  gap: 14px;
}

.icon {
  width: 42px;
  height: 42px;

  border-radius: 14px;

  background: #dbeafe;
  color: #2563eb;

  display: flex;
  align-items: center;
  justify-content: center;

  flex-shrink: 0;
}

.resource h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 700;

  color: #0f172a;
}

.resource p {
  margin: 4px 0 0;

  font-size: 14px;

  color: #64748b;
}

/* STATUS */
.status {
  padding: 6px 12px;

  border-radius: 999px;

  font-size: 12px;
  font-weight: 700;
}

.confirmed {
  background: #dcfce7;
  color: #15803d;
}

.pending {
  background: #fed7aa;
  color: #c2410c;
}

.cancelled {
  background: #e2e8f0;
  color: #475569;
}

/* INFO */
.info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 10px;

  color: #475569;

  font-size: 14px;
}

/* PARTICIPANTS */
.participants {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.participants-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  display: flex;
  align-items: center;
  gap: 8px;

  color: #475569;

  font-size: 14px;
}

.progress {
  width: 100%;
  height: 10px;

  background: #e2e8f0;

  border-radius: 999px;

  overflow: hidden;
}

.progress-fill {
  height: 100%;

  background: linear-gradient(
    90deg,
    #2563eb,
    #f97316
  );

  border-radius: 999px;

  transition: width 0.3s ease;
}

/* ACTIONS */
.actions {
  display: flex;
  gap: 12px;
}

.actions button {
  flex: 1;

  border: none;

  border-radius: 14px;

  padding: 12px;

  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  font-size: 14px;
  font-weight: 600;

  cursor: pointer;

  transition: 0.2s;
}

/* Secondary */
.secondary {
  background: #eff6ff;
  color: #2563eb;
}

.secondary:hover {
  background: #dbeafe;
}

/* Danger */
.danger {
  background: #fee2e2;
  color: #dc2626;
}

.danger:hover {
  background: #fecaca;
}

/* Mobile */
@media (max-width: 768px) {
  .card {
    padding: 18px;
  }

  .top {
    flex-direction: column;
  }

  .actions {
    flex-direction: column;
  }
}
</style>
