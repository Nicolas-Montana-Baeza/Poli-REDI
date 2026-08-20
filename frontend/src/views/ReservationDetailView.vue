<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  ArrowLeft,
  CalendarDays,
  Clock,
  Copy,
  KeyRound,
  RotateCw,
  Timer,
  UserRound,
  UsersRound,
  XCircle
} from 'lucide-vue-next'
import ParticipantsProgress from '@/components/ui/ParticipantsProgress.vue'
import {reservationsService} from '@/services/reservations.service'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus,
  isReservationCancelable
} from '@/utils/reservationTime'
const participants = ref([])
const participantsLoading = ref(false)
const participantsError = ref('')
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const reservationsStore = useReservationsStore()
const cancelling = ref(false)
const cancelConfirmationOpen = ref(false)
const rotatingJoinCode = ref(false)
const joinCode = ref('')
const joinCodeCopied = ref(false)
const groupActionError = ref('')
const reservationId = computed(() => {
  return Number(route.params.id)
})

onMounted(async () => {
  await authStore.loadAuthUser()

  await reservationsStore.fetchReservationDetail(
    reservationId.value
  )

  await loadParticipants()
})
const reservation = computed(() => {
  return reservationsStore.reservationDetail
})

const isLoading = computed(() => {
  return authStore.loading || reservationsStore.detailLoading
})

const loadingError = computed(() => {
  return authStore.error || reservationsStore.detailLoadingError
})

const canCancel = computed(() => {
  return isReservationCancelable(reservation.value)
})

const statusLabel = computed(() => {
  return getReservationDisplayStatus(reservation.value).label
})

const statusClass = computed(() => {
  return getReservationDisplayStatus(reservation.value).className
})

const isGroupReservation = computed(() => {
  return reservation.value?.isGroupReservation === true
})
const activeParticipants = computed(() => {
  return participants.value.filter(
    participant => participant.status === 'CONFIRMED'
  )
})
const canManageGroup = computed(() => {
  if (!reservation.value || !authStore.user) {
    return false
  }

  return (
    authStore.user.isAdmin ||
    reservation.value.userId === authStore.user.id
  )
})
// ------------------------------------------------------------
// Participantes de la reserva grupal.
// ------------------------------------------------------------
//
// La lista completa solo se solicita cuando el usuario tiene permisos
// de gestión sobre la reserva. El backend vuelve a validar owner/admin,
// por lo que esta condición frontend es únicamente una mejora de UX.
const loadParticipants = async () => {
  if (
    !reservation.value ||
    !isGroupReservation.value ||
    !canManageGroup.value
  ) {
    return
  }

  participantsLoading.value = true
  participantsError.value = ''

  try {
    participants.value =
      await reservationsService.getParticipants(
        reservation.value.id
      )
  } catch (error) {
    participantsError.value =
      error?.response?.data?.error ||
      'No fue posible cargar los participantes'
  } finally {
    participantsLoading.value = false
  }
}
const reservationUserName = computed(() => {
  if (authStore.user?.isAdmin) {
    return (
      reservation.value?.userFullName ||
      reservation.value?.userEmail ||
      `Usuario #${reservation.value?.userId}`
    )
  }

  return (
    authStore.user?.fullName ||
    authStore.user?.email ||
    'Usuario'
  )
})

const reservationUserRut = computed(() => {
  if (authStore.user?.isAdmin) {
    return reservation.value?.userRut || 'RUT no registrado'
  }

  return authStore.user?.rut || 'RUT no registrado'
})

const cancelReservation = async () => {
  if (!reservation.value) {
    return
  }

  cancelling.value = true

  try {
    await reservationsStore.cancelReservation(
      reservation.value.id
    )

    cancelConfirmationOpen.value = false

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store conserva el mensaje de error para la vista.
  } finally {
    cancelling.value = false
  }
}

// ------------------------------------------------------------
// Gestión del código de invitación grupal.
// ------------------------------------------------------------
//
// El código solo vive temporalmente en memoria después de rotarlo.
// No se recupera desde backend ni se persiste en localStorage.
const rotateJoinCode = async () => {
  if (!reservation.value || rotatingJoinCode.value) {
    return
  }

  rotatingJoinCode.value = true
  groupActionError.value = ''
  joinCode.value = ''
  joinCodeCopied.value = false

  try {
    const result = await reservationsService.rotateJoinCode(
      reservation.value.id
    )

    joinCode.value = result.joinCode || ''
  } catch (error) {
    groupActionError.value =
      error?.response?.data?.error ||
      'No fue posible generar un nuevo código'
  } finally {
    rotatingJoinCode.value = false
  }
}

const copyJoinCode = async () => {
  if (!joinCode.value) {
    return
  }

  try {
    await navigator.clipboard.writeText(joinCode.value)
    joinCodeCopied.value = true
  } catch {
    groupActionError.value =
      'No fue posible copiar el código al portapapeles'
  }
}

const goBack = () => {
  if (route.query.from === 'history') {
    router.push('/history')
    return
  }

  router.push('/reservations')
}
</script>

<template>
  <main class="detail-view">

    <button
      class="back-button"
      type="button"
      @click="goBack"
    >
      <ArrowLeft :size="18" />
      Volver
    </button>

    <div
      v-if="isLoading"
      aria-label="Cargando reserva"
    >
      <SkeletonLoader
        variant="resources"
        :items="4"
      />
    </div>

    <div
      v-else-if="loadingError"
      class="state-card error"
    >
      {{ loadingError }}
    </div>

    <div
      v-else-if="!reservation"
      class="state-card"
    >
      No se encontro la reserva solicitada.
    </div>

    <section
      v-else
      class="detail-panel"
    >

      <header class="detail-header">

        <div>

          <span
            class="status"
            :class="statusClass"
          >
            {{ statusLabel }}
          </span>

          <h1>
            {{ reservation.title || 'Reserva' }}
          </h1>

          <p>
            {{ reservation.resourceName || 'Recurso' }}
          </p>

        </div>

        <button
          v-if="canCancel"
          class="cancel-button"
          type="button"
          :disabled="cancelling"
          @click="cancelConfirmationOpen = true"
        >
          <XCircle :size="18" />
          Cancelar
        </button>

      </header>

      <section
        v-if="canCancel && cancelConfirmationOpen"
        class="cancel-confirmation"
        role="alert"
      >
        <div>
          <strong>
            ¿Cancelar esta reserva?
          </strong>

          <p>
            Esta acción no se puede deshacer.
            La reserva dejará de estar activa.
          </p>
        </div>

        <div class="cancel-confirmation__actions">

          <button
            type="button"
            class="cancel-confirmation__back"
            :disabled="cancelling"
            @click="cancelConfirmationOpen = false"
          >
            Volver
          </button>

          <button
            type="button"
            class="cancel-button"
            :disabled="cancelling"
            @click="cancelReservation"
          >
            <XCircle :size="18" />

            {{
              cancelling
                ? 'Cancelando...'
                : 'Sí, cancelar reserva'
            }}
          </button>

        </div>
      </section>

      <div
        v-if="reservationsStore.actionError"
        class="state-card error"
      >
        {{ reservationsStore.actionError }}
      </div>

      <div
        v-if="reservationsStore.actionSuccess"
        class="state-card success"
      >
        {{ reservationsStore.actionSuccess }}
      </div>

      <section class="details-grid">

        <article class="detail-item">

          <CalendarDays :size="22" />

          <span>
            Fecha
          </span>

          <strong>
            {{ formatReservationDate(reservation.startTime) }}
          </strong>

        </article>

        <article class="detail-item">

          <Clock :size="22" />

          <span>
            Horario
          </span>

          <strong>
            {{ formatReservationTimeRange(
              reservation.startTime,
              reservation.durationMinutes
            ) }}
          </strong>

        </article>

        <article class="detail-item">

          <Timer :size="22" />

          <span>
            Duración
          </span>

          <strong>
            {{ reservation.durationMinutes }} minutos
          </strong>

        </article>

        <article class="detail-item">

          <UserRound :size="22" />

          <span>
            Usuario
          </span>

          <strong>
            {{ reservationUserName }}
          </strong>

          <small>
            {{ reservationUserRut }}
          </small>

        </article>

      </section>
     <section
  v-if="isGroupReservation"
  class="group-panel"
>
  <div class="group-header">
    <div>
      <h2>Reserva grupal</h2>
      <p>
        Seguimiento del mínimo requerido y gestión del código de invitación.
      </p>
    </div>

    <KeyRound :size="22" />
  </div>

  <ParticipantsProgress
    :participant-count="reservation.participantCount"
    :minimum-participants="reservation.minimumParticipants"
    :capacity="reservation.capacity"
    :status="reservation.status"
    :group-condition="reservation.groupCondition"
  />
<section
  v-if="canManageGroup"
  class="participants-panel"
>
  <div class="participants-header">
    <div>
      <h3>Participantes</h3>

      <p>
        {{ activeParticipants.length }}
        participantes confirmados
      </p>
    </div>

    <UsersRound :size="22" />
  </div>

  <div
    v-if="participantsLoading"
    class="participants-state"
  >
    Cargando participantes...
  </div>

  <div
    v-else-if="participantsError"
    class="state-card error"
  >
    {{ participantsError }}
  </div>

  <div
    v-else-if="activeParticipants.length === 0"
    class="participants-state"
  >
    No hay participantes confirmados.
  </div>

  <div
    v-else
    class="participants-list"
  >
    <article
      v-for="participant in activeParticipants"
      :key="participant.userId"
      class="participant-item"
    >
      <div>
        <strong>
          {{ participant.fullName || participant.email }}
        </strong>

        <span>
          {{ participant.email }}
        </span>

        <small v-if="participant.rut">
          {{ participant.rut }}
        </small>
      </div>

      <span
        v-if="participant.isOwner"
        class="owner-badge"
      >
        Responsable
      </span>
    </article>
  </div>
</section>
  <div
    v-if="groupActionError"
    class="state-card error"
  >
    {{ groupActionError }}
  </div>

  <div
    v-if="canManageGroup"
    class="group-actions"
  >
    <div>
      <strong>Código de invitación</strong>

      <p>
        Por seguridad, el código actual no puede recuperarse.
        Puedes generar uno nuevo cuando lo necesites.
      </p>
    </div>

    <button
      class="rotate-code-button"
      type="button"
      :disabled="rotatingJoinCode"
      @click="rotateJoinCode"
    >
      <RotateCw
        :size="18"
        :class="{ spinning: rotatingJoinCode }"
      />

      {{ rotatingJoinCode
        ? 'Generando...'
        : 'Generar nuevo código'
      }}
    </button>
  </div>

  <div
    v-if="joinCode"
    class="join-code-result"
  >
    <div>
      <span>Nuevo código</span>

      <strong>
        {{ joinCode }}
      </strong>

      <small>
        Este código reemplaza al anterior.
      </small>
    </div>

    <button
      class="copy-code-button"
      type="button"
      @click="copyJoinCode"
    >
      <Copy :size="18" />

      {{ joinCodeCopied
        ? 'Copiado'
        : 'Copiar'
      }}
    </button>
  </div>
</section>
    </section>

  </main>
</template>

<style scoped>
.detail-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 18px;
}
.participants-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;

  padding-top: 16px;

  border-top: 1px solid var(--color-border);
}

.participants-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.participants-header h3 {
  margin: 0;

  color: var(--color-text);

  font-size: 18px;
  font-weight: 800;
}

.participants-header p {
  margin: 5px 0 0;

  color: var(--color-text-muted);

  font-size: 13px;
}

.participants-header svg {
  color: var(--color-primary);
}

.participants-list {
  display: flex;
  flex-direction: column;

  gap: 10px;
}

.participant-item {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 14px;

  padding: 12px 14px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  background: #f8fafc;
}

.participant-item div {
  display: flex;
  flex-direction: column;

  gap: 3px;
}

.participant-item strong {
  color: var(--color-text);
}

.participant-item span,
.participant-item small,
.participants-state {
  color: var(--color-text-muted);

  font-size: 13px;
}

.owner-badge {
  flex-shrink: 0;

  padding: 5px 9px;

  border-radius: var(--radius-pill);

  background: #dbeafe;

  color: #1d4ed8 !important;

  font-size: 12px !important;
  font-weight: 750;
}
.back-button,
.cancel-button {
  border: none;
  border-radius: var(--radius-md);

  cursor: pointer;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  padding: 10px 13px;

  font-weight: 750;

  transition: 0.2s;
}

.back-button {
  width: fit-content;

  background: #f1f5f9;

  color: #334155;
}

.back-button:hover {
  background: #e2e8f0;
}

.cancel-button {
  background: #fee2e2;
  border: 1px solid #fecaca;

  color: #b91c1c;

  white-space: nowrap;
}

.cancel-button:hover:not(:disabled) {
  background: #fecaca;
}

.cancel-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.state-card,
.detail-panel,
.detail-item {
  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.state-card {
  padding: var(--space-4);

  color: #334155;

  font-weight: 650;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.state-card.success {
  background: #dcfce7;

  color: #166534;

  border-color: #bbf7d0;
}

.detail-panel {
  padding: var(--space-5);

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow: var(--shadow-card);
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.status {
  display: inline-flex;

  padding: 6px 10px;

  border-radius: var(--radius-pill);

  background: #eff6ff;
  color: #1d4ed8;

  font-size: 12px;
  font-weight: 750;
}

.status.cancelled {
  background: #fee2e2;

  color: #b91c1c;
}

.status.pending {
  background: #fef3c7;

  color: #92400e;
}

.status.completed {
  background: #e2e8f0;

  color: #475569;
}

.status.ongoing {
  background: #dcfce7;

  color: #166534;
}

.detail-header h1 {
  margin: 12px 0 0;

  color: var(--color-text);

  font-size: 28px;
  font-weight: 850;
}

.detail-header p {
  margin: 7px 0 0;

  color: var(--color-text-muted);
}

.details-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));

  gap: 14px;
}

.detail-item {
  padding: var(--space-4);

  display: flex;
  flex-direction: column;

  gap: 9px;
}

.detail-item svg {
  color: var(--color-primary);
}

.detail-item span {
  color: var(--color-text-muted);

  font-size: 13px;
  font-weight: 700;
}

.detail-item strong {
  color: var(--color-text);

  overflow-wrap: anywhere;
}

.detail-item small {
  color: var(--color-text-muted);

  font-size: 13px;
  font-weight: 650;

  overflow-wrap: anywhere;
}
.group-panel {
  display: flex;
  flex-direction: column;
  gap: 18px;

  padding: var(--space-4);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  background: var(--color-surface);
}

.group-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.group-header h2 {
  margin: 0;

  color: var(--color-text);

  font-size: 20px;
  font-weight: 800;
}

.group-header p,
.group-actions p {
  margin: 6px 0 0;

  color: var(--color-text-muted);

  font-size: 14px;
}

.group-header svg {
  color: var(--color-primary);
}

.group-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  padding-top: 16px;

  border-top: 1px solid var(--color-border);
}

.group-actions strong {
  color: var(--color-text);
}

.rotate-code-button,
.copy-code-button {
  border: none;
  border-radius: var(--radius-md);

  cursor: pointer;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  padding: 10px 14px;

  font-weight: 750;
}

.rotate-code-button {
  background: var(--color-primary);

  color: white;
}

.rotate-code-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.copy-code-button {
  background: #e2e8f0;

  color: #334155;
}

.join-code-result {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  padding: var(--space-4);

  border: 1px solid #bfdbfe;
  border-radius: var(--radius-md);

  background: #eff6ff;
}

.join-code-result div {
  display: flex;
  flex-direction: column;

  gap: 5px;
}

.join-code-result span,
.join-code-result small {
  color: #475569;

  font-size: 13px;
}

.join-code-result strong {
  color: #1d4ed8;

  font-family: monospace;
  font-size: 20px;
  letter-spacing: 1px;
}

.spinning {
  animation: group-code-spin 0.8s linear infinite;
}

@keyframes group-code-spin {
  to {
    transform: rotate(360deg);
  }
}
@media (max-width: 768px) {
  .group-actions,
  .join-code-result {
    align-items: stretch;
    flex-direction: column;
  }
  .participant-item {
    align-items: flex-start;
    flex-direction: column;
  }
  .rotate-code-button,
  .copy-code-button {
    width: 100%;
  }
  .detail-header {
    flex-direction: column;
  }

  .detail-header h1 {
    font-size: 26px;
  }

  .cancel-button {
    width: 100%;
  }
}

.cancel-confirmation {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;

  padding: 16px 18px;

  background: var(--color-error-soft);
  border: 1px solid var(--color-error-border);
  border-radius: var(--radius-lg);
}

.cancel-confirmation strong {
  color: var(--color-error);
}

.cancel-confirmation p {
  margin: 4px 0 0;

  color: var(--color-text-muted);

  font-size: 14px;
}

.cancel-confirmation__actions {
  display: flex;
  gap: 10px;

  flex-shrink: 0;
}

.cancel-confirmation__back {
  padding: 10px 16px;

  background: var(--color-surface);
  color: var(--color-text);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  font-weight: 700;
  cursor: pointer;
}

.cancel-confirmation__back:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

</style>
