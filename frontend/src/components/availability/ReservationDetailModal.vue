<script setup>
import { computed } from 'vue'
import {
  formatReservationDate,
  formatReservationTimeRange
} from '@/utils/reservationTime'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },

  reservation: {
    type: Object,
    default: null
  },

  errorMessage: {
    type: String,
    default: ''
  },

  canCancel: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits([
  'close',
  'cancel'
])

const reservationDate = computed(() => {
  return formatReservationDate(
    props.reservation?.startTime
  )
})

const reservationTime = computed(() => {
  const duration =
    props.reservation?.durationMinutes || 60

  return formatReservationTimeRange(
    props.reservation?.startTime,
    duration
  )
})

const statusLabel = computed(() => {
  switch (props.reservation?.status) {
    case 'CONFIRMED':
      return 'Confirmada'

    case 'PENDING':
      return 'Pendiente'

    case 'CANCELLED':
      return 'Cancelada'

    default:
      return 'Sin estado'
  }
})

const showCancelAction = computed(() => {
  return (
    props.canCancel &&
    props.reservation?.status !== 'CANCELLED'
  )
})

const details = computed(() => {
  if (!props.reservation) {
    return []
  }

  return [
    {
      label: 'Actividad',
      value: props.reservation.title || 'Reserva'
    },
    {
      label: 'Fecha',
      value: reservationDate.value
    },
    {
      label: 'Horario',
      value: reservationTime.value
    },
    {
      label: 'Duración',
      value: `${props.reservation.durationMinutes} minutos`
    },
    {
      label: 'Estado',
      value: statusLabel.value,
      wide: true
    }
  ]
})

const handleCancel = () => {
  emit('cancel')
}
</script>

<template>
  <Teleport to="body">

    <div
      v-if="visible"
      class="overlay"
      @click.self="emit('close')"
    >

      <div class="modal">

        <header class="modal-header">

          <div>

            <h2>
              Detalle de reserva
            </h2>

            <p>
              Información de la reserva seleccionada.
            </p>

          </div>

          <button
            class="close-btn"
            type="button"
            aria-label="Cerrar"
            @click="emit('close')"
          >
            x
          </button>

        </header>

        <section
          v-if="reservation"
          class="detail-panel"
        >

          <dl class="detail-list">

            <div
              v-for="detail in details"
              :key="detail.label"
              class="detail-row"
            >

              <dt>
                {{ detail.label }}
              </dt>

              <dd
                :class="{ status: detail.label === 'Estado' }"
              >
                {{ detail.value }}
              </dd>

            </div>

          </dl>

        </section>

        <div
          v-if="errorMessage"
          class="error"
        >
          {{ errorMessage }}
        </div>

        <div class="warning">
          <template v-if="showCancelAction">
            Esta acción cambiará la reserva a estado cancelada.
          </template>

          <template v-else>
            Solo el administrador o quien creo la reserva puede cancelarla.
          </template>
        </div>

        <footer class="actions">

          <button
            class="secondary-btn app-button secondary"
            type="button"
            @click="emit('close')"
          >
            Cerrar
          </button>

          <button
            v-if="showCancelAction"
            class="danger-btn app-button danger"
            type="button"
            @click="handleCancel"
          >
            Cancelar reserva
          </button>

        </footer>

      </div>

    </div>

  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15, 23, 42, 0.55);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 18px;

  z-index: 9999;

  backdrop-filter: blur(4px);
}

.modal {
  width: min(100%, 500px);
  max-height: min(88vh, 640px);

  background: var(--color-surface);

  border-radius: var(--radius-xl);

  padding: 22px;

  display: flex;
  flex-direction: column;

  gap: 16px;

  overflow: hidden;

  box-shadow: var(--shadow-modal);
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.modal-header h2 {
  margin: 0;

  font-size: 24px;
  font-weight: 800;

  color: var(--color-text);
}

.modal-header p {
  margin: 4px 0 0;

  color: var(--color-text-muted);

  font-size: 13px;
}

.close-btn {
  width: 38px;
  height: 38px;

  border: none;

  border-radius: var(--radius-md);

  background: var(--color-surface-soft);

  color: #334155;

  cursor: pointer;

  font-size: 18px;
  font-weight: 800;
}

.close-btn:hover {
  background: #e2e8f0;
}

.detail-panel {
  padding: 4px 0;

  overflow-y: auto;
}

.detail-list {
  margin: 0;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  overflow: hidden;

  background: var(--color-surface-muted);
}

.detail-row {
  display: grid;
  grid-template-columns: 126px minmax(0, 1fr);
  align-items: center;

  gap: 16px;

  padding: 13px 16px;

  border-bottom: 1px solid var(--color-border);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-row dt {
  margin: 0;

  color: var(--color-text-muted);

  font-size: 13px;
  font-weight: 800;
}

.detail-row dd {
  margin: 0;

  color: var(--color-text);

  font-size: 14px;
  font-weight: 800;

  overflow-wrap: anywhere;
}

.status {
  color: var(--color-primary);
}

.warning,
.error {
  padding: 11px 14px;

  border-radius: var(--radius-md);

  font-size: 13px;
  font-weight: 800;
}

.warning {
  background: var(--color-warning-soft);

  border: 1px solid var(--color-warning-border);

  color: var(--color-warning);
}

.error {
  background: var(--color-error-soft);

  border: 1px solid var(--color-error-border);

  color: var(--color-error);
}

.actions {
  display: flex;
  justify-content: flex-end;

  gap: 12px;
}

.actions button {
  height: 44px;

  padding: 0 18px;

  font-size: 14px;
  font-weight: 800;

  cursor: pointer;
}

.secondary-btn {
  min-width: 116px;
}

.danger-btn {
  min-width: 160px;
}

@media (max-width: 520px) {
  .overlay {
    align-items: flex-end;

    padding: 12px;
  }

  .modal {
    width: 100%;
    max-height: 92vh;

    padding: 16px;

    border-radius: 20px;
  }

  .modal-header h2 {
    font-size: 20px;
  }

  .detail-row {
    grid-template-columns: 1fr;

    gap: 5px;

    padding: 12px 14px;
  }

  .actions {
    flex-direction: column;
  }

  .actions button {
    width: 100%;
  }
}
</style>
