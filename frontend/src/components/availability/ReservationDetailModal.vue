<script setup>
import { computed } from 'vue'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus,
  isReservationCancelable
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

const isAvailabilityBlock = computed(() => {
  return props.reservation?.type === 'blocked' ||
    props.reservation?.isAvailabilityBlock === true
})

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

const displayStatus = computed(() => {
  if (isAvailabilityBlock.value) {
    return {
      label: 'Bloqueo',
      className: 'scheduled'
    }
  }

  if (props.reservation?.isScheduledActivity) {
    return {
      label: 'Programación institucional',
      className: 'scheduled'
    }
  }

  if (props.reservation?.isWorkshop) {
    return {
      label: 'Taller programado',
      className: 'workshop'
    }
  }

  return getReservationDisplayStatus(props.reservation)
})

const modalTitle = computed(() => {
  if (isAvailabilityBlock.value) {
    return 'Bloqueo de disponibilidad'
  }

  if (props.reservation?.isScheduledActivity) {
    return 'Detalle de programación'
  }

  if (props.reservation?.isWorkshop) {
    return 'Detalle de taller'
  }

  return 'Detalle de reserva'
})

const modalDescription = computed(() => {
  if (
    isAvailabilityBlock.value ||
    props.reservation?.isScheduledActivity ||
    props.reservation?.isWorkshop
  ) {
    return 'Información del bloque seleccionado.'
  }

  return 'Información de la reserva seleccionada.'
})

const showCancelAction = computed(() => {
  return (
    props.canCancel &&
    !isAvailabilityBlock.value &&
    !props.reservation?.isWorkshop &&
    !props.reservation?.isScheduledActivity &&
    isReservationCancelable(props.reservation)
  )
})

const warningMessage = computed(() => {
  if (isAvailabilityBlock.value) {
    return 'Este horario no está disponible para reservas.'
  }

  if (props.reservation?.isScheduledActivity) {
    return 'Este bloque corresponde a programación institucional y ocupa la instalación durante ese horario.'
  }

  if (props.reservation?.isWorkshop) {
    return 'Este bloque corresponde a un taller deportivo y ocupa la instalación durante ese horario.'
  }

  if (showCancelAction.value) {
    return 'Esta acción cambiará la reserva a estado cancelada.'
  }

  return 'Solo el administrador o quien creó la reserva puede cancelarla.'
})

const details = computed(() => {
  if (!props.reservation) {
    return []
  }

  return [
    {
      label:
        isAvailabilityBlock.value
          ? 'Bloqueo'
          : props.reservation.isScheduledActivity
          ? 'Programación'
          : props.reservation.isWorkshop
            ? 'Taller'
            : 'Actividad',
      value: props.reservation.title || 'Reserva'
    },
    {
      label: 'Recurso',
      value: props.reservation.resourceName || 'Recurso'
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
      value: displayStatus.value.label,
      statusClass: displayStatus.value.className,
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
              {{ modalTitle }}
            </h2>

            <p>
              {{ modalDescription }}
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
                :class="[
                  { status: detail.label === 'Estado' },
                  detail.statusClass
                ]"
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
          {{ warningMessage }}

          <template v-if="false">
            Esta acción cambiará la reserva a estado cancelada.
          </template>

          <template v-if="false">
            Solo el administrador o quien creó la reserva puede cancelarla.
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

  padding: 20px;

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

  font-size: 23px;
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

  padding: 12px 15px;

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

.status.confirmed,
.status.completed,
.status.ongoing {
  color: var(--color-success);
}

.status.pending,
.status.workshop,
.status.scheduled {
  color: var(--color-warning);
}

.status.cancelled,
.status.rejected,
.status.expired {
  color: var(--color-error);
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
