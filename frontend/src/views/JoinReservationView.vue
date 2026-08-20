<script setup>
import { computed, ref } from 'vue'

import ParticipantsProgress from '../components/ui/ParticipantsProgress.vue'
import { reservationsService } from '../services/reservations.service'
import { getReservationDisplayStatus } from '../utils/reservationTime'

// ------------------------------------------------------------
// Estado de la búsqueda.
// ------------------------------------------------------------

const joinCode = ref('')
const progress = ref(null)

const loading = ref(false)
const actionLoading = ref(false)

const errorMessage = ref('')
const successMessage = ref('')

// Normalizamos solo para presentación y envío.
//
// El backend vuelve a normalizar el valor, por lo que esta lógica no
// representa una regla de seguridad; únicamente mejora la experiencia.
const normalizedJoinCode = computed(() => {
  return joinCode.value
    .trim()
    .toUpperCase()
})

const canSearch = computed(() => {
  return normalizedJoinCode.value.length > 0 && !loading.value
})

const reservationStatusLabel = computed(() => {
  return getReservationDisplayStatus(
    progress.value
  ).label
})

// ------------------------------------------------------------
// Mensajes de error provenientes de la API.
// ------------------------------------------------------------

const resolveApiError = (error) => {
  return (
    error?.response?.data?.error ||
    error?.message ||
    'No fue posible completar la operación.'
  )
}

// ------------------------------------------------------------
// Consulta del grupo.
// ------------------------------------------------------------

const searchReservation = async () => {
  if (!canSearch.value) {
    return
  }

  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  progress.value = null

  try {
    progress.value = await reservationsService.getGroupProgress(
      normalizedJoinCode.value
    )
  } catch (error) {
    errorMessage.value = resolveApiError(error)
  } finally {
    loading.value = false
  }
}

// ------------------------------------------------------------
// Incorporación y retiro.
// ------------------------------------------------------------
//
// Las respuestas de POST y DELETE contienen el progreso actualizado,
// por lo que no necesitamos hacer un segundo GET después de cada acción.
const joinReservation = async () => {
  actionLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    progress.value = await reservationsService.joinGroup(
      normalizedJoinCode.value
    )

    successMessage.value = 'Te uniste correctamente a la reserva.'
  } catch (error) {
    errorMessage.value = resolveApiError(error)
  } finally {
    actionLoading.value = false
  }
}

const leaveReservation = async () => {
  actionLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    progress.value = await reservationsService.leaveGroup(
      normalizedJoinCode.value
    )

    successMessage.value = 'Saliste correctamente de la reserva.'
  } catch (error) {
    errorMessage.value = resolveApiError(error)
  } finally {
    actionLoading.value = false
  }
}
</script>

<template>
  <main class="join-reservation">
    <header class="join-reservation__header">
      <p class="join-reservation__eyebrow">
        Reserva grupal
      </p>

      <h1>Unirse a una reserva</h1>

      <p class="join-reservation__description">
        Ingresa el código entregado por la persona que creó la reserva.
      </p>
    </header>

    <section class="join-card">
      <form
        class="join-form"
        @submit.prevent="searchReservation"
      >
        <label
          for="join-code"
          class="join-form__label"
        >
          Código de unión
        </label>

        <div class="join-form__controls">
          <input
            id="join-code"
            v-model="joinCode"
            type="text"
            autocomplete="off"
            maxlength="11"
            placeholder="ABCDE-FGHIJ"
            class="join-form__input"
            @input="joinCode = joinCode.toUpperCase()"
          >

          <button
            type="submit"
            class="button button--primary"
            :disabled="!canSearch"
          >
            {{ loading ? 'Buscando...' : 'Buscar reserva' }}
          </button>
        </div>
      </form>

      <p
        v-if="errorMessage"
        class="message message--error"
        role="alert"
      >
        {{ errorMessage }}
      </p>

      <p
        v-if="successMessage"
        class="message message--success"
      >
        {{ successMessage }}
      </p>

      <div
        v-if="progress"
        class="reservation-result"
      >
        <ParticipantsProgress
          :participant-count="progress.participantCount"
          :minimum-participants="progress.minimumParticipants"
          :capacity="progress.capacity"
          :status="progress.status"
          :group-condition="progress.groupCondition"
        />

        <div class="reservation-result__meta">
          <div>
            <span>Estado de reserva</span>
            <strong>{{ reservationStatusLabel }}</strong>
          </div>

          <div>
            <span>Tu participación</span>
            <strong>
              {{ progress.isMember ? 'Participando' : 'No participas' }}
            </strong>
          </div>
        </div>

        <div class="reservation-result__actions">
          <button
            v-if="!progress.isMember"
            type="button"
            class="button button--primary"
            :disabled="actionLoading"
            @click="joinReservation"
          >
            {{ actionLoading ? 'Procesando...' : 'Unirme a la reserva' }}
          </button>

          <button
            v-else-if="!progress.isOwner"
            type="button"
            class="button button--danger"
            :disabled="actionLoading"
            @click="leaveReservation"
          >
            {{ actionLoading ? 'Procesando...' : 'Retirarme' }}
          </button>

          <p
            v-else
            class="reservation-result__owner"
          >
            Eres el propietario de esta reserva. Para abandonarla debes
            cancelar la reserva completa.
          </p>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.join-reservation {
  width: min(760px, 100%);
  margin: 0 auto;
  padding: var(--space-8) var(--space-4);
}

.join-reservation__header {
  margin-bottom: var(--space-6);
}

.join-reservation__eyebrow {
  margin-bottom: var(--space-2);

  color: var(--color-primary);
  font-size: var(--text-help);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.join-reservation h1 {
  margin-bottom: var(--space-2);

  color: var(--color-text);
  font-size: var(--text-page-title);
}

.join-reservation__description {
  color: var(--color-text-muted);
  line-height: 1.5;
}

.join-card {
  display: grid;
  gap: var(--space-5);

  padding: var(--space-6);

  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-card);
}

.join-form {
  display: grid;
  gap: var(--space-2);
}

.join-form__label {
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.join-form__controls {
  display: flex;
  gap: var(--space-2);
}

.join-form__input {
  flex: 1;
  min-width: 0;

  padding: 11px 13px;

  color: var(--color-text);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  font-weight: 700;
  letter-spacing: 0.08em;
}

.join-form__input:focus {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
  outline: none;
}

.button {
  padding: 10px 16px;

  border: 0;
  border-radius: var(--radius-md);

  font-weight: 700;
  cursor: pointer;
}

.button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.button--primary {
  color: var(--color-primary-contrast);
  background: var(--color-primary);
}

.button--primary:hover:not(:disabled) {
  background: var(--color-primary-strong);
}

.button--danger {
  color: #fff;
  background: var(--color-error-strong);
}

.message {
  padding: var(--space-3);

  border: 1px solid;
  border-radius: var(--radius-md);

  font-size: var(--text-help);
}

.message--error {
  color: var(--color-error);
  background: var(--color-error-soft);
  border-color: var(--color-error-border);
}

.message--success {
  color: var(--color-success);
  background: var(--color-success-soft);
  border-color: var(--color-success-border);
}

.reservation-result {
  display: grid;
  gap: var(--space-4);
}

.reservation-result__meta {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-3);
}

.reservation-result__meta div {
  display: grid;
  gap: var(--space-1);

  padding: var(--space-3);

  background: var(--color-surface-muted);
  border-radius: var(--radius-md);
}

.reservation-result__meta span {
  color: var(--color-text-muted);
  font-size: var(--text-help);
}

.reservation-result__actions {
  display: flex;
  align-items: center;
}

.reservation-result__owner {
  color: var(--color-text-muted);
  font-size: var(--text-help);
  line-height: 1.5;
}

@media (max-width: 640px) {
  .join-form__controls {
    flex-direction: column;
  }

  .reservation-result__meta {
    grid-template-columns: 1fr;
  }

  .button {
    width: 100%;
  }
}
</style>
