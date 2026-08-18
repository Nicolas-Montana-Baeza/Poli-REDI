<script setup>
import { computed } from 'vue'

const props = defineProps({
  participantCount: {
    type: Number,
    default: 0
  },

  minimumParticipants: {
    type: Number,
    default: 0
  },

  capacity: {
    type: Number,
    default: 0
  },

  status: {
    type: String,
    default: 'PENDING'
  },

  groupCondition: {
    type: String,
    default: 'PENDING_MINIMUM'
  }
})

// ------------------------------------------------------------
// Progreso visual del grupo.
// ------------------------------------------------------------
//
// La barra representa el avance hacia el mínimo requerido para confirmar
// la reserva, no hacia la capacidad máxima del recurso.
//
// Ejemplo:
//
//     7 participantes / mínimo 10 = 70 %
//
// Una vez alcanzado el mínimo se mantiene en 100 %, aunque posteriormente
// existan 15 participantes en una cancha con capacidad para 22.
const progressPercentage = computed(() => {
  if (props.minimumParticipants <= 0) {
    return 0
  }

  return Math.min(
    100,
    Math.max(
      0,
      Math.round(
        (props.participantCount / props.minimumParticipants) * 100
      )
    )
  )
})

// Cantidad de participantes que todavía faltan para alcanzar el mínimo.
//
// Nunca se muestran números negativos una vez confirmado el grupo.
const missingParticipants = computed(() => {
  return Math.max(
    0,
    props.minimumParticipants - props.participantCount
  )
})

// ------------------------------------------------------------
// Estado grupal.
// ------------------------------------------------------------
//
// groupCondition describe la salud actual del grupo y es independiente de
// reservation.status.
//
// Esto es especialmente importante para la regla:
//
//     reservation.status = CONFIRMED
//     groupCondition      = AT_RISK
//
// cuando una reserva ya confirmada vuelve a quedar bajo el mínimo.
const conditionLabel = computed(() => {
  switch (props.groupCondition) {
    case 'HEALTHY':
      return 'Grupo completo'

    case 'AT_RISK':
      return 'Grupo en riesgo'

    case 'PENDING_MINIMUM':
      return 'Pendiente de participantes'

    default:
      return 'Sin seguimiento grupal'
  }
})

const conditionClass = computed(() => {
  switch (props.groupCondition) {
    case 'HEALTHY':
      return 'participants-progress--healthy'

    case 'AT_RISK':
      return 'participants-progress--risk'

    default:
      return 'participants-progress--pending'
  }
})

const description = computed(() => {
  if (props.groupCondition === 'HEALTHY') {
    return 'La reserva cumple con el mínimo de participantes.'
  }

  if (props.groupCondition === 'AT_RISK') {
    return `Faltan ${missingParticipants.value} participante${
      missingParticipants.value === 1 ? '' : 's'
    } para recuperar el mínimo requerido.`
  }

  if (missingParticipants.value === 0) {
    return 'El grupo alcanzó el mínimo requerido.'
  }

  return `Faltan ${missingParticipants.value} participante${
    missingParticipants.value === 1 ? '' : 's'
  } para confirmar la reserva.`
})
</script>

<template>
  <section
    class="participants-progress"
    :class="conditionClass"
    aria-label="Progreso de participantes"
  >
    <div class="participants-progress__header">
      <div>
        <p class="participants-progress__eyebrow">
          Participantes
        </p>

        <strong class="participants-progress__count">
          {{ participantCount }} / {{ minimumParticipants }}
        </strong>
      </div>

      <span class="participants-progress__status">
        {{ conditionLabel }}
      </span>
    </div>

    <div
      class="participants-progress__track"
      role="progressbar"
      :aria-valuenow="progressPercentage"
      aria-valuemin="0"
      aria-valuemax="100"
    >
      <div
        class="participants-progress__bar"
        :style="{ width: `${progressPercentage}%` }"
      />
    </div>

    <div class="participants-progress__footer">
      <p class="participants-progress__description">
        {{ description }}
      </p>

      <span
        v-if="capacity > 0"
        class="participants-progress__capacity"
      >
        Capacidad máxima: {{ capacity }}
      </span>
    </div>
  </section>
</template>

<style scoped>
.participants-progress {
  display: grid;
  gap: var(--space-3);

  padding: var(--space-4);

  background: var(--color-surface);
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-xl);
}

.participants-progress__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.participants-progress__eyebrow {
  margin-bottom: var(--space-1);

  color: var(--color-text-muted);
  font-size: var(--text-help);
  font-weight: 600;
}

.participants-progress__count {
  color: var(--color-text);
  font-size: 24px;
  line-height: 1;
}

.participants-progress__status {
  padding: 6px 10px;

  border: 1px solid;
  border-radius: var(--radius-pill);

  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

.participants-progress__track {
  width: 100%;
  height: 10px;

  overflow: hidden;

  background: var(--color-surface-soft);
  border-radius: var(--radius-pill);
}

.participants-progress__bar {
  height: 100%;

  border-radius: inherit;

  transition: width 180ms ease;
}

.participants-progress__footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.participants-progress__description,
.participants-progress__capacity {
  color: var(--color-text-muted);
  font-size: var(--text-help);
  line-height: 1.45;
}

.participants-progress__capacity {
  flex-shrink: 0;
}

/* ----------------------------------------------------------
 * Estado pendiente.
 * ---------------------------------------------------------- */

.participants-progress--pending .participants-progress__status {
  color: var(--color-warning);
  background: var(--color-warning-soft);
  border-color: var(--color-warning-border);
}

.participants-progress--pending .participants-progress__bar {
  background: var(--color-primary);
}

/* ----------------------------------------------------------
 * Grupo saludable.
 * ---------------------------------------------------------- */

.participants-progress--healthy .participants-progress__status {
  color: var(--color-success);
  background: var(--color-success-soft);
  border-color: var(--color-success-border);
}

.participants-progress--healthy .participants-progress__bar {
  background: var(--color-success);
}

/* ----------------------------------------------------------
 * Grupo confirmado pero bajo el mínimo.
 * ---------------------------------------------------------- */

.participants-progress--risk .participants-progress__status {
  color: var(--color-error);
  background: var(--color-error-soft);
  border-color: var(--color-error-border);
}

.participants-progress--risk .participants-progress__bar {
  background: var(--color-error-strong);
}

@media (max-width: 640px) {
  .participants-progress__header,
  .participants-progress__footer {
    flex-direction: column;
  }
}
</style>
