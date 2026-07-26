<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  ArrowLeft,
  CalendarDays,
  Clock,
  Timer,
  UserRound,
  XCircle
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import ParticipantsProgress from '@/components/ui/ParticipantsProgress.vue'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { reservationsService } from '@/services/reservations.service'
import { useAuthStore } from '@/stores/auth'
import { useReservationsStore } from '@/stores/reservations'
import {
  canUserCancelReservation,
  formatReservationDate,
  formatReservationTimeRange,
  getReservationDisplayStatus
} from '@/utils/reservationTime'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const reservationsStore = useReservationsStore()
const cancelling = ref(false)
const joinCode = ref('')
const joinCodeOpen = ref(false)
const groupBusy = ref(false)
const groupError = ref('')
const targetValue = ref(null)
const cancelModalOpen = ref(false)
const rotateModalOpen = ref(false)
const copyFeedback = ref('')
const joinCodeUnavailable = ref(false)
let joinOperationGeneration = 0
let componentMounted = true

const reservationId = computed(() => {
  return Number(route.params.id)
})

onMounted(async () => {
  const user = await authStore.loadAuthUser()

  if (user?.isAdmin) {
    await reservationsStore.fetchReservations()
    return
  }

  await reservationsStore.fetchMyReservations()
})

const reservation = computed(() => {
  const source = authStore.user?.isAdmin
    ? reservationsStore.reservations
    : reservationsStore.myReservations

  return source.find(
    (item) => item.id === reservationId.value
  )
})
const clearJoinCodeState = () => {
  joinOperationGeneration += 1
  joinCode.value = ''
  joinCodeOpen.value = false
  joinCodeUnavailable.value = false
  copyFeedback.value = ''
  groupError.value = ''
  rotateModalOpen.value = false
  groupBusy.value = false
}
watch(reservation, value => { targetValue.value = value?.targetParticipants ?? null })
watch(reservationId, clearJoinCodeState)
onBeforeUnmount(() => {
  componentMounted = false
  clearJoinCodeState()
})

const isLoading = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.loading
    : reservationsStore.myLoading
})

const loadingError = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.loadingError
    : reservationsStore.myLoadingError
})

const canCancel = computed(() => {
  return canUserCancelReservation(reservation.value, authStore.user)
})
const isOwnGroup = computed(() => Boolean(reservation.value?.targetParticipants) && reservation.value?.userId === authStore.user?.id)
const ownerProgress = computed(() => reservation.value ? { ...reservation.value, reservationId: reservation.value.id, isOwner: true, isMember: true } : null)
const toggleJoinCode = async () => {
  joinCodeOpen.value = !joinCodeOpen.value
  if (!joinCodeOpen.value) {
    clearJoinCodeState()
    return
  }
  if (!joinCodeOpen.value || joinCode.value) return
  const requestedReservationId = reservation.value.id
  const operationGeneration = ++joinOperationGeneration
  groupBusy.value = true; groupError.value = ''
  try {
    const response = await reservationsService.getJoinCode(requestedReservationId)
    if (!componentMounted || !joinCodeOpen.value || reservationId.value !== requestedReservationId || operationGeneration !== joinOperationGeneration) return
    joinCode.value = response.joinCode || ''
    joinCodeUnavailable.value = !joinCode.value
  } catch (error) {
    if (!componentMounted || !joinCodeOpen.value || reservationId.value !== requestedReservationId || operationGeneration !== joinOperationGeneration) return
    if (error?.response?.status === 404) joinCodeUnavailable.value = true
    else groupError.value = 'No se pudo recuperar el código.'
  } finally {
    if (componentMounted && operationGeneration === joinOperationGeneration) groupBusy.value = false
  }
}
const rotateJoinCode = async () => {
  const requestedReservationId = reservation.value.id
  const operationGeneration = ++joinOperationGeneration
  groupBusy.value = true; groupError.value = ''
  try {
    const response = await reservationsService.rotateJoinCode(requestedReservationId)
    if (!componentMounted || reservationId.value !== requestedReservationId || operationGeneration !== joinOperationGeneration) return
    joinCode.value = response.joinCode
    joinCodeUnavailable.value = false
    joinCodeOpen.value = true
    rotateModalOpen.value = false
    copyFeedback.value = 'Código nuevo generado. El código anterior dejó de funcionar.'
  }
  catch {
    if (!componentMounted || reservationId.value !== requestedReservationId || operationGeneration !== joinOperationGeneration) return
    groupError.value = 'No se pudo generar un código nuevo.'
  } finally {
    if (componentMounted && operationGeneration === joinOperationGeneration) groupBusy.value = false
  }
}
const copyJoinLink = async () => {
  const link = `${window.location.origin}/join/${encodeURIComponent(joinCode.value)}`
  try {
    await navigator.clipboard.writeText(link)
    copyFeedback.value = 'Enlace de invitación copiado.'
  } catch {
    copyFeedback.value = 'No se pudo copiar el enlace.'
  }
}
const copyJoinCode = async () => {
  try {
    await navigator.clipboard.writeText(joinCode.value)
    copyFeedback.value = 'Código copiado al portapapeles.'
  } catch {
    copyFeedback.value = 'No se pudo copiar. Selecciona el código manualmente.'
  }
}
const updateTarget = async () => {
  groupBusy.value = true; groupError.value = ''
  try {
    const updated = await reservationsService.updateTarget(reservation.value.id, Number(targetValue.value))
    Object.assign(reservation.value, updated)
  } catch (e) { groupError.value = e?.response?.data?.error || 'No se pudo actualizar el objetivo.' }
  finally { groupBusy.value = false }
}

const statusLabel = computed(() => {
  return getReservationDisplayStatus(reservation.value).label
})

const statusClass = computed(() => {
  return getReservationDisplayStatus(reservation.value).className
})
const statusBadgeStatus = computed(() => ({
  completed: 'INACTIVE',
  ongoing: 'ACTIVE'
}[statusClass.value] || reservation.value?.status))

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

  cancelModalOpen.value = false
  cancelling.value = true

  try {
    await reservationsStore.cancelReservation(
      reservation.value.id
    )

    reservationsStore.setActionSuccess(
      'Reserva cancelada correctamente'
    )
  } catch {
    // El store conserva el mensaje de error para la vista.
  } finally {
    cancelling.value = false
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

          <StatusBadge :status="statusBadgeStatus" :label="statusLabel" />

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
          @click="cancelModalOpen = true"
        >
          <XCircle :size="18" />
          Cancelar
        </button>

      </header>

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

      <section v-if="isOwnGroup" class="group-panel">
        <ParticipantsProgress :progress="ownerProgress" :busy="groupBusy" />
        <form v-if="reservation.canEditTarget" @submit.prevent="updateTarget">
          <label for="owner-target">Objetivo de participantes</label>
          <input id="owner-target" v-model.number="targetValue" type="number" :min="Math.max(reservation.minimumParticipants,reservation.participantCount)" :max="reservation.capacity" required>
          <button :disabled="groupBusy">Guardar objetivo</button>
        </form>
        <button type="button" :aria-expanded="joinCodeOpen" @click="toggleJoinCode">Código de invitación</button>
        <div v-if="joinCodeOpen">
          <p v-if="groupBusy">Cargando…</p>
          <output v-if="joinCode">{{ joinCode }}</output>
          <button v-if="joinCode" type="button" @click="copyJoinCode">Copiar</button>
          <button v-if="joinCode" type="button" @click="copyJoinLink">Copiar enlace</button>
          <button v-if="joinCode" type="button" :disabled="groupBusy" @click="rotateModalOpen = true">Generar código nuevo</button>
          <template v-else-if="joinCodeUnavailable && !groupBusy">
            <p>Esta reserva todavía no tiene un código de invitación.</p>
            <button type="button" @click="rotateModalOpen = true">Generar código</button>
          </template>
          <p v-if="copyFeedback" role="status">{{ copyFeedback }}</p>
        </div>
        <p v-if="groupError" class="state-card error" role="alert">{{ groupError }}</p>
      </section>

    </section>
    <ConfirmModal
      :show="cancelModalOpen"
      title="Cancelar reserva"
      message="Esta acción cancelará la reserva y no se puede deshacer."
      confirm-text="Sí, cancelar reserva"
      cancel-text="Mantener reserva"
      variant="danger"
      destructive
      :loading="cancelling"
      @confirm="cancelReservation"
      @cancel="cancelModalOpen = false"
    />
    <ConfirmModal
      :show="rotateModalOpen"
      title="Generar código nuevo"
      message="El código actual dejará de funcionar inmediatamente. Comparte el nuevo código con las personas invitadas."
      confirm-text="Generar código nuevo"
      cancel-text="Conservar código actual"
      destructive
      :loading="groupBusy"
      @confirm="rotateJoinCode"
      @cancel="rotateModalOpen = false"
    />

  </main>
</template>

<style scoped>
.detail-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 18px;
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
.group-panel{display:grid;gap:1rem;margin-top:1rem}.group-panel form{display:flex;gap:.75rem;align-items:end;flex-wrap:wrap}.group-panel input,.group-panel button{min-height:44px;padding:.65rem}.group-panel output{display:block;padding:.75rem;background:#f1f5f9;border-radius:.5rem;overflow-wrap:anywhere}

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

@media (max-width: 768px) {
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
</style>
