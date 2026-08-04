<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  ArrowLeft,
  CalendarDays,
  Clipboard,
  Clock,
  KeyRound,
  Minus,
  Plus,
  RefreshCw,
  Share2,
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
const initializing = ref(true)
let joinOperationGeneration = 0
let componentMounted = true

const reservationId = computed(() => {
  return Number(route.params.id)
})

onMounted(async () => {
  try {
    const user = await authStore.loadAuthUser()

    if (user?.isAdmin) {
      await reservationsStore.fetchReservations()
      return
    }

    await reservationsStore.fetchMyReservations()
  } finally {
    initializing.value = false
  }
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
watch(reservation, value => { targetValue.value = value?.targetParticipants ?? null }, { immediate: true })
watch(reservationId, clearJoinCodeState)
onBeforeUnmount(() => {
  componentMounted = false
  clearJoinCodeState()
})

const isLoading = computed(() => {
  const queryInitialLoading = authStore.user?.isAdmin
    ? (reservationsStore.initialLoading ??
      (reservationsStore.loading && !reservationsStore.hasLoaded))
    : (reservationsStore.myInitialLoading ??
      (reservationsStore.myLoading && !reservationsStore.myHasLoaded))

  return initializing.value || queryInitialLoading
})

const loadingError = computed(() => {
  return authStore.user?.isAdmin
    ? reservationsStore.loadingError
    : reservationsStore.myLoadingError
})

const canCancel = computed(() => {
  return canUserCancelReservation(reservation.value, authStore.user)
})
const isGroupReservation = computed(() => (
  reservation.value?.targetParticipants != null ||
  reservation.value?.capacity != null
))
const canManageJoinCode = computed(() => {
  const status = reservation.value?.status
  return (
    isGroupReservation.value &&
    Number(reservation.value?.userId) === Number(authStore.user?.id) &&
    (status === 'PENDING' || status === 'CONFIRMED')
  )
})
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
const shareJoinLink = async () => {
  const url = `${window.location.origin}/join/${encodeURIComponent(joinCode.value)}`
  if (navigator.share) {
    try {
      await navigator.share({ title: 'Invitación a reserva grupal', url })
      copyFeedback.value = 'Invitación compartida.'
      return
    } catch (error) {
      if (error?.name === 'AbortError') return
    }
  }
  await copyJoinLink()
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
const minimumTarget = computed(() => Math.max(
  Number(reservation.value?.minimumParticipants) || 0,
  Number(reservation.value?.participantCount) || 0
))
const changeTarget = amount => {
  const current = Number(targetValue.value) || minimumTarget.value
  const capacity = Number(reservation.value?.capacity) || current
  targetValue.value = Math.min(capacity, Math.max(minimumTarget.value, current + amount))
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
      role="status"
      aria-live="polite"
    >
      <SkeletonLoader
        variant="detail"
        :items="1"
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
          Cancelar reserva
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
            Responsable
          </span>

          <strong>
            {{ reservationUserName }}
          </strong>

        </article>

      </section>

      <section v-if="canManageJoinCode" class="group-panel">
        <ParticipantsProgress :progress="ownerProgress" :busy="groupBusy" :show-status="false" />
        <div class="group-management">
          <form v-if="reservation.canEditTarget" class="target-card" @submit.prevent="updateTarget">
            <div class="management-copy">
              <strong>Objetivo de participantes</strong>
              <span>Puedes cambiarlo hasta una hora antes</span>
            </div>
            <div class="target-actions">
              <div class="stepper">
                <button type="button" aria-label="Disminuir objetivo" :disabled="groupBusy || targetValue <= minimumTarget" @click="changeTarget(-1)">
                  <Minus :size="18" />
                </button>
                <input id="owner-target" v-model.number="targetValue" aria-label="Objetivo de participantes" type="number" :min="minimumTarget" :max="reservation.capacity" required>
                <button type="button" aria-label="Aumentar objetivo" :disabled="groupBusy || targetValue >= reservation.capacity" @click="changeTarget(1)">
                  <Plus :size="18" />
                </button>
              </div>
              <button class="save-target" :disabled="groupBusy">Guardar cambios</button>
            </div>
          </form>

          <section class="invitation-card">
            <div class="invitation-heading">
              <div class="invitation-icon"><KeyRound :size="21" /></div>
              <div class="management-copy">
                <strong>Invita a tu grupo</strong>
              </div>
              <button class="invitation-toggle" type="button" :aria-expanded="joinCodeOpen" @click="toggleJoinCode">
                Código de invitación
              </button>
            </div>
            <div v-if="joinCodeOpen" class="invitation-content">
              <p v-if="groupBusy" class="loading-code">Cargando…</p>
              <div v-if="joinCode" class="code-row">
                <output aria-label="Código de invitación">{{ joinCode }}</output>
                <button type="button" title="Copiar código" @click="copyJoinCode"><Clipboard :size="18" /> Copiar código</button>
              </div>
              <div v-if="joinCode" class="invitation-actions">
                <button type="button" @click="shareJoinLink"><Share2 :size="18" /> Compartir invitación</button>
                <button type="button" :disabled="groupBusy" @click="rotateModalOpen = true"><RefreshCw :size="18" /> Generar código nuevo</button>
              </div>
              <template v-else-if="joinCodeUnavailable && !groupBusy">
                <p class="empty-code">Esta reserva todavía no tiene un código de invitación.</p>
                <button class="generate-code" type="button" @click="rotateModalOpen = true">Generar código</button>
              </template>
              <p v-if="copyFeedback" class="copy-feedback" role="status">{{ copyFeedback }}</p>
            </div>
          </section>
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
.group-panel { display: grid; gap: 16px; margin-top: 4px; }
.group-management { display: grid; grid-template-columns: minmax(280px, .85fr) minmax(360px, 1.15fr); gap: 14px; }
.target-card, .invitation-card { min-width: 0; padding: 18px; border: 1px solid var(--color-border); border-radius: var(--radius-lg); background: var(--color-surface); }
.target-card { display: flex; flex-direction: column; justify-content: space-between; gap: 18px; }
.management-copy { display: grid; gap: 4px; }
.management-copy strong { color: var(--color-text); font-size: 16px; }
.management-copy span { color: var(--color-text-muted); font-size: 13px; line-height: 1.45; }
.target-actions { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.stepper { display: grid; grid-template-columns: 42px 64px 42px; min-height: 44px; overflow: hidden; border: 1px solid var(--color-border); border-radius: var(--radius-md); background: var(--color-surface); }
.stepper button, .stepper input { min-width: 0; border: 0; background: transparent; color: var(--color-text); }
.stepper button { display: grid; place-items: center; cursor: pointer; }
.stepper button:hover:not(:disabled) { background: var(--color-surface-muted); color: var(--color-primary); }
.stepper button:disabled { cursor: not-allowed; opacity: .4; }
.stepper input { width: 100%; border-inline: 1px solid var(--color-border); text-align: center; font-size: 16px; font-weight: 800; appearance: textfield; }
.stepper input::-webkit-inner-spin-button, .stepper input::-webkit-outer-spin-button { margin: 0; appearance: none; }
.save-target, .invitation-toggle, .generate-code { min-height: 44px; padding: 10px 16px; border: 0; border-radius: var(--radius-md); background: var(--color-primary); color: #fff; cursor: pointer; font-weight: 750; }
.save-target:hover:not(:disabled), .invitation-toggle:hover:not(:disabled), .generate-code:hover:not(:disabled) { filter: brightness(.94); }
.save-target:disabled, .invitation-toggle:disabled, .generate-code:disabled { cursor: not-allowed; opacity: .6; }
.invitation-card { display: grid; gap: 16px; }
.invitation-heading { display: grid; grid-template-columns: auto 1fr auto; align-items: center; gap: 12px; }
.invitation-icon { width: 42px; height: 42px; display: grid; place-items: center; border-radius: 12px; background: #eff6ff; color: var(--color-primary); }
.invitation-content { display: grid; gap: 12px; padding-top: 14px; border-top: 1px solid var(--color-border); }
.code-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 8px; }
.code-row output { min-width: 0; padding: 13px 14px; overflow-wrap: anywhere; border: 1px dashed #93c5fd; border-radius: var(--radius-md); background: #eff6ff; color: #1e3a8a; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-weight: 750; }
.code-row button, .invitation-actions button { min-height: 44px; padding: 10px 13px; display: inline-flex; align-items: center; justify-content: center; gap: 7px; border: 1px solid var(--color-border); border-radius: var(--radius-md); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-weight: 700; }
.code-row button:hover, .invitation-actions button:hover:not(:disabled) { border-color: #93c5fd; background: #eff6ff; color: #1d4ed8; }
.invitation-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.loading-code, .empty-code, .copy-feedback { margin: 0; color: var(--color-text-muted); line-height: 1.5; }
.copy-feedback { color: #166534; font-weight: 650; }

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

  .group-management { grid-template-columns: 1fr; }
  .invitation-heading { grid-template-columns: auto 1fr; }
  .invitation-toggle { grid-column: 1 / -1; width: 100%; }
  .code-row { grid-template-columns: 1fr; }
  .target-actions, .save-target, .invitation-actions button { width: 100%; }
  .stepper { flex: 1; }
}
</style>
