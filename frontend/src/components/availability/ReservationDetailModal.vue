<script setup>
import { computed } from 'vue'

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
  }
})

const emit = defineEmits([
  'close',
  'cancel'
])

const reservationDate = computed(() => {
  if (!props.reservation?.startTime) return ''

  const date = new Date(props.reservation.startTime)

  return date.toLocaleDateString('es-CL', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
})

const reservationTime = computed(() => {
  if (!props.reservation?.startTime) return ''

  const startDate = new Date(props.reservation.startTime)

  const duration =
    props.reservation.durationMinutes || 60

  const endDate =
    new Date(startDate.getTime() + duration * 60000)

  const start =
    startDate.toTimeString().slice(0, 5)

  const end =
    endDate.toTimeString().slice(0, 5)

  return `${start} - ${end}`
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

        <div class="modal-header">

          <div>

            <h2>
              Detalle de reserva
            </h2>

            <p>
              Revisa la información antes de cancelar.
            </p>

          </div>

          <button
            class="close-btn"
            @click="emit('close')"
          >
            ✕
          </button>

        </div>

        <div
          v-if="reservation"
          class="detail-card"
        >

          <div class="detail-row">

            <span>
              Actividad
            </span>

            <strong>
              {{ reservation.title || 'Reserva' }}
            </strong>

          </div>

          <div class="detail-row">

            <span>
              Fecha
            </span>

            <strong>
              {{ reservationDate }}
            </strong>

          </div>

          <div class="detail-row">

            <span>
              Horario
            </span>

            <strong>
              {{ reservationTime }}
            </strong>

          </div>

          <div class="detail-row">

            <span>
              Duración
            </span>

            <strong>
              {{ reservation.durationMinutes }} minutos
            </strong>

          </div>

          <div class="detail-row">

            <span>
              Estado
            </span>

            <strong class="status">
              {{ statusLabel }}
            </strong>

          </div>

        </div>

        <div
          v-if="errorMessage"
          class="error"
        >
          {{ errorMessage }}
        </div>

        <div class="warning">
          Esta acción cambiará la reserva a estado cancelada.
        </div>

        <div class="actions">

          <button
            class="secondary-btn"
            @click="emit('close')"
          >
            Cerrar
          </button>

          <button
            class="danger-btn"
            @click="handleCancel"
          >
            Cancelar reserva
          </button>

        </div>

      </div>

    </div>

  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15,23,42,0.55);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 20px;

  z-index: 9999;

  backdrop-filter: blur(4px);
}

.modal {
  width: 100%;
  max-width: 560px;

  background: white;

  border-radius: 28px;

  padding: 28px;

  display: flex;
  flex-direction: column;

  gap: 22px;

  box-shadow:
    0 24px 60px rgba(0,0,0,0.2);
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;
}

.modal-header h2 {
  margin: 0;

  font-size: 28px;
  font-weight: 800;

  color: #0f172a;
}

.modal-header p {
  margin-top: 6px;

  color: #64748b;

  font-size: 14px;
}

.close-btn {
  width: 44px;
  height: 44px;

  border: none;

  border-radius: 14px;

  background: #f1f5f9;

  color: #334155;

  cursor: pointer;

  font-size: 18px;
}

.detail-card {
  display: flex;
  flex-direction: column;

  gap: 12px;

  padding: 18px;

  border-radius: 20px;

  background: #f8fafc;

  border: 1px solid #e2e8f0;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;

  gap: 16px;

  padding-bottom: 10px;

  border-bottom: 1px solid #e2e8f0;
}

.detail-row:last-child {
  padding-bottom: 0;

  border-bottom: none;
}

.detail-row span {
  color: #64748b;

  font-size: 14px;
  font-weight: 700;
}

.detail-row strong {
  color: #0f172a;

  font-size: 14px;
  text-align: right;
}

.status {
  color: #2563eb;
}

.warning {
  padding: 14px 16px;

  border-radius: 16px;

  background: #fff7ed;

  border: 1px solid #fed7aa;

  color: #c2410c;

  font-size: 14px;
  font-weight: 700;
}

.error {
  padding: 14px 16px;

  border-radius: 16px;

  background: #fee2e2;

  border: 1px solid #fecaca;

  color: #b91c1c;

  font-size: 14px;
  font-weight: 700;
}

.actions {
  display: flex;
  justify-content: flex-end;

  gap: 14px;
}

.actions button {
  height: 50px;

  border: none;

  border-radius: 16px;

  padding: 0 22px;

  font-size: 14px;
  font-weight: 800;

  cursor: pointer;
}

.secondary-btn {
  background: #f1f5f9;

  color: #334155;
}

.danger-btn {
  background: #dc2626;

  color: white;
}

.danger-btn:hover {
  background: #b91c1c;
}

@media (max-width: 768px) {
  .modal {
    padding: 22px;

    border-radius: 24px;
  }

  .actions {
    flex-direction: column;
  }

  .actions button {
    width: 100%;
  }

  .detail-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .detail-row strong {
    text-align: left;
  }
}
</style>