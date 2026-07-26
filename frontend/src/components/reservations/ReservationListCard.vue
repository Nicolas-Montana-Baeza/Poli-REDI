<script setup>
import { computed } from 'vue'
import {
  CalendarDays,
  Clock,
  Timer,
  XCircle
} from 'lucide-vue-next'

import {
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus
} from '@/utils/reservationTime'
import StatusBadge from '@/components/ui/StatusBadge.vue'

const props = defineProps({
  reservation: {
    type: Object,
    required: true
  },

  mode: {
    type: String,
    default: 'active',
    validator: (value) => ['active', 'history'].includes(value)
  },

  detailTo: {
    type: [String, Object],
    default: ''
  },

  detailLabel: {
    type: String,
    default: 'Detalle'
  },

  showCancel: {
    type: Boolean,
    default: false
  },

  cancelDisabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'open-detail',
  'cancel'
])
const displayStatus = computed(() => getReservationDisplayStatus(props.reservation))
const badgeStatus = computed(() => ({
  completed: 'INACTIVE',
  ongoing: 'ACTIVE'
}[displayStatus.value.className] || props.reservation.status))

const handleCardKeydown = (event) => {
  if (props.detailTo) {
    return
  }

  if (event.key !== 'Enter' && event.key !== ' ') {
    return
  }

  event.preventDefault()
  emit('open-detail', props.reservation)
}

const handleCancel = () => {
  emit('cancel', props.reservation)
}
</script>

<template>
  <article
    class="reservation-list-card"
    :class="{ interactive: !detailTo && mode === 'history' }"
    :role="!detailTo && mode === 'history' ? 'link' : undefined"
    :tabindex="!detailTo && mode === 'history' ? 0 : undefined"
    :aria-label="!detailTo && mode === 'history'
      ? `Ver detalle de ${reservation.title || 'reserva'}`
      : undefined"
    @click="!detailTo && mode === 'history' ? emit('open-detail', reservation) : null"
    @keydown="handleCardKeydown"
  >

    <div class="card-main">

      <div class="card-copy">

        <StatusBadge
          class="reservation-status"
          :status="badgeStatus"
          :label="displayStatus.label"
        />

        <h2>
          {{ reservation.title || 'Reserva' }}
        </h2>

        <p>
          {{ reservation.resourceName || 'Recurso' }}
        </p>

      </div>

      <div class="card-actions">

        <RouterLink
          v-if="detailTo"
          class="app-button secondary detail-link"
          :to="detailTo"
        >
          {{ detailLabel }}
        </RouterLink>

        <span
          v-else-if="mode === 'history'"
          class="detail-action"
          aria-hidden="true"
        >
          {{ detailLabel }}
        </span>

        <button
          v-if="showCancel"
          class="app-button danger cancel-button"
          type="button"
          :disabled="cancelDisabled"
          @click.stop="handleCancel"
        >
          <XCircle :size="18" />
          Cancelar
        </button>

      </div>

    </div>

    <div class="reservation-meta">

      <span>
        <CalendarDays :size="17" />
        {{ formatReservationDate(reservation.startTime) }}
      </span>

      <span>
        <Clock :size="17" />
        {{ formatReservationTimeRange(
          reservation.startTime,
          reservation.durationMinutes
        ) }}
      </span>

      <span>
        <Timer :size="17" />
        {{ reservation.durationMinutes }} minutos
      </span>

    </div>

  </article>
</template>

<style scoped>
.reservation-list-card {
  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  padding: var(--space-4);

  display: flex;
  flex-direction: column;

  gap: 16px;

  box-shadow: var(--shadow-card);

  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.reservation-list-card.interactive {
  cursor: pointer;
}

.reservation-list-card.interactive:hover,
.reservation-list-card.interactive:focus-visible {
  border-color: #bfdbfe;

  box-shadow: var(--shadow-card);

  outline: none;
}

.card-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: var(--space-4);
}

.card-copy {
  min-width: 0;
}

.reservation-status {
  cursor: default;
}

.reservation-status.cancelled {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.reservation-status.pending {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.reservation-status.completed {
  background: var(--color-border-soft);
  color: #475569;
}

.reservation-status.ongoing {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.reservation-status.rejected,
.reservation-status.expired {
  background: var(--color-error-soft);
  color: var(--color-error);
}

h2 {
  margin: 10px 0 0;

  color: var(--color-text);

  font-size: 18px;
  font-weight: 750;
  line-height: 1.2;
}

p {
  margin: var(--space-2) 0 0;

  color: var(--color-text-muted);
}

.card-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;

  gap: 10px;
}

.detail-link,
.cancel-button {
  min-height: 38px;
  padding: 8px 12px;

  text-decoration: none;

  white-space: nowrap;
}

.detail-action {
  align-self: flex-start;

  border-radius: var(--radius-pill);

  background: var(--color-primary-soft);

  padding: 7px 11px;

  color: var(--color-primary-strong);

  font-size: 13px;
  font-weight: 750;

  white-space: nowrap;
}

.reservation-meta {
  display: flex;
  flex-wrap: wrap;

  gap: 10px;
}

.reservation-meta span {
  display: inline-flex;
  align-items: center;
  gap: 8px;

  background: var(--color-surface-muted);

  border-radius: var(--radius-pill);

  padding: 7px 10px;

  color: #475569;

  font-size: 13px;
  font-weight: 650;
}

@media (max-width: 768px) {
  .card-main {
    flex-direction: column;
  }

  .card-actions,
  .detail-link,
  .cancel-button {
    width: 100%;
  }
}

@media (max-width: 520px) {
  .reservation-list-card {
    padding: var(--space-4);
  }

  .reservation-meta span,
  .detail-action {
    width: 100%;

    justify-content: center;
  }
}
</style>
