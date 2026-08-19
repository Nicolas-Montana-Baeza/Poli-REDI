<script setup>
import {
  computed,
  onMounted,
  reactive,
  ref
} from 'vue'

import {
  AlertTriangle,
  ArrowLeft,
  CalendarClock,
  Check,
  CircleOff,
  Clock3,
  RefreshCw,
  ShieldCheck
} from 'lucide-vue-next'

import {
  useSchedulingConflictsStore
} from '@/stores/schedulingConflicts'

const store =
  useSchedulingConflictsStore()

const selectedId = ref(null)

const forms = reactive({})

// ============================================================================
// CONFIRMACIÓN DE DECISIONES
// ============================================================================
//
// Las resoluciones administrativas son finales para una ocurrencia.
// Antes de enviar el PATCH exigimos una confirmación explícita para evitar
// decisiones accidentales, especialmente ALLOW, CANCEL y RESCHEDULE.

const pendingDecision = ref(null)

const filters = [
  {
    value: 'PENDING',
    label: 'Pendientes'
  },
  {
    value: 'RESOLVED',
    label: 'Resueltos'
  },
  {
    value: 'ALL',
    label: 'Todos'
  }
]

const resolutionLabels = {
  PENDING: 'Pendiente',
  KEEP: 'Mantener',
  ALLOW: 'Permitir convivencia',
  CANCEL: 'Cancelar',
  RESCHEDULE: 'Reprogramar'
}

const confirmationMessages = {
  KEEP:
    'La ocurrencia seleccionada se mantendrá en su horario actual.',

  ALLOW:
    'Se autorizará explícitamente esta convivencia pese a la incompatibilidad detectada.',

  CANCEL:
    'La ocurrencia seleccionada será cancelada para resolver el conflicto.',

  RESCHEDULE:
    'La ocurrencia seleccionada será movida al nuevo horario indicado.'
}

const entityLabels = {
  INSTITUTIONAL_ACTIVITY:
    'Actividad institucional',

  RESERVATION:
    'Reserva'
}

onMounted(async () => {
  await store.fetchConflicts('PENDING')
})

const selectedConflict = computed(
  () => store.selectedConflict
)

const pendingItems = computed(
  () =>
    selectedConflict.value?.items?.filter(
      item =>
        item.resolution === 'PENDING'
    ) || []
)

const ensureForm = (itemId) => {
  if (!forms[itemId]) {
    forms[itemId] = {
      resolutionNote: '',
      newDate: '',
      newStartTime: '',
      newEndTime: ''
    }
  }

  return forms[itemId]
}

// ============================================================================
// FECHAS
// ============================================================================

const formatDateTime = (value) => {
  if (!value) {
    return ''
  }

  const date = new Date(value)

  return new Intl.DateTimeFormat(
    'es-CL',
    {
      dateStyle: 'medium',
      timeStyle: 'short',
      timeZone: 'America/Santiago'
    }
  ).format(date)
}

const formatDate = (value) => {
  if (!value) {
    return ''
  }

  const date = new Date(value)

  return new Intl.DateTimeFormat(
    'es-CL',
    {
      dateStyle: 'medium',
      timeZone: 'America/Santiago'
    }
  ).format(date)
}

// ============================================================================
// NAVEGACIÓN INTERNA
// ============================================================================

const loadConflict = async (conflict) => {
  selectedId.value = conflict.id

  store.clearMessages()

  const detail =
    await store.fetchConflict(
      conflict.id
    )

  for (
    const item of detail.items || []
  ) {
    ensureForm(item.id)
  }
}

const closeDetail = () => {
  selectedId.value = null
  store.selectedConflict = null
  store.clearMessages()
}

const changeFilter = async (status) => {
  closeDetail()

  await store.fetchConflicts(status)
}

// ============================================================================
// RESOLUCIÓN
// ============================================================================

const prepareResolution = (
  item,
  resolution
) => {
  const form = ensureForm(item.id)

  const resolutionNote =
    form.resolutionNote.trim()

  store.clearMessages()

  if (!resolutionNote) {
    store.actionError =
      'Debes ingresar una nota administrativa antes de resolver.'

    return
  }

  if (resolution === 'RESCHEDULE') {
    if (
      !form.newDate ||
      !form.newStartTime ||
      !form.newEndTime
    ) {
      store.actionError =
        'Para reprogramar debes indicar fecha, hora de inicio y hora de término.'

      return
    }
  }

  pendingDecision.value = {
    item,
    resolution,

    payload: {
      resolution,
      resolutionNote,

      ...(resolution === 'RESCHEDULE'
        ? {
            newDate:
              form.newDate,

            newStartTime:
              form.newStartTime,

            newEndTime:
              form.newEndTime
          }
        : {})
    }
  }
}

const cancelPendingDecision = () => {
  pendingDecision.value = null
}

const confirmResolution = async () => {
  const decision =
    pendingDecision.value

  if (!decision) {
    return
  }

  const {
    item,
    payload
  } = decision

  try {
    await store.resolveItem(
      selectedConflict.value.id,
      item.id,
      payload
    )

    const form =
      ensureForm(item.id)

    // Una justificación nunca se reutiliza automáticamente en otra decisión.
    form.resolutionNote = ''
    form.newDate = ''
    form.newStartTime = ''
    form.newEndTime = ''

    pendingDecision.value = null

    if (
      store.selectedConflict.status ===
      'RESOLVED'
    ) {
      await store.fetchConflicts(
        store.statusFilter
      )
    }
  } catch {
    // El store conserva el error retornado por el backend.
    // El modal permanece abierto para que el administrador pueda revisarlo.
  }
}

</script>

<template>
  <main class="conflicts-view">

    <template v-if="!selectedId">

      <header class="page-header">
        <div>
          <h1>
            Conflictos de programación
          </h1>

          <p>
            Revisa y resuelve incompatibilidades
            entre reservas y programación institucional.
          </p>
        </div>

        <button
          type="button"
          class="refresh-button"
          :disabled="store.loading"
          @click="
            store.fetchConflicts(
              store.statusFilter
            )
          "
        >
          <RefreshCw :size="17" />
          Actualizar
        </button>
      </header>

      <section class="filters">
        <button
          v-for="filter in filters"
          :key="filter.value"
          type="button"
          class="filter-button"
          :class="{
            active:
              store.statusFilter ===
              filter.value
          }"
          @click="
            changeFilter(filter.value)
          "
        >
          {{ filter.label }}
        </button>
      </section>

      <div
        v-if="store.error"
        class="state-card error"
      >
        {{ store.error }}
      </div>

      <div
        v-if="store.loading"
        class="state-card"
      >
        Cargando conflictos...
      </div>

      <section
        v-else-if="!store.conflicts.length"
        class="state-card"
      >
        No existen conflictos para este filtro.
      </section>

      <section
        v-else
        class="conflict-grid"
      >
        <article
          v-for="conflict in store.conflicts"
          :key="conflict.id"
          class="conflict-card"
        >
          <div class="card-heading">
            <div class="conflict-icon">
              <AlertTriangle :size="22" />
            </div>

            <div>
              <span class="eyebrow">
                Conflicto #{{ conflict.id }}
              </span>

              <h2>
                {{
                  conflict.resourceName ||
                  `Recurso ${conflict.resourceId}`
                }}
              </h2>
            </div>

            <span
              class="status-pill"
              :class="
                conflict.status.toLowerCase()
              "
            >
              {{ conflict.status }}
            </span>
          </div>

          <div class="card-body">
            <p>
              <strong>
                {{
                  conflict.items?.length || 0
                }}
              </strong>
              elementos involucrados
            </p>

            <p
              v-if="
                conflict.resolutionSummary
              "
              class="summary"
            >
              {{ conflict.resolutionSummary }}
            </p>

            <small>
              Creado:
              {{
                formatDate(
                  conflict.createdAt
                )
              }}
            </small>
          </div>

          <button
            type="button"
            class="primary-button"
            @click="loadConflict(conflict)"
          >
            {{
              conflict.status === 'PENDING'
                ? 'Resolver conflicto'
                : 'Ver resolución'
            }}
          </button>
        </article>
      </section>

    </template>

    <template v-else>

      <button
        type="button"
        class="back-button"
        @click="closeDetail"
      >
        <ArrowLeft :size="18" />
        Volver a conflictos
      </button>

      <div
        v-if="store.loadingDetail"
        class="state-card"
      >
        Cargando detalle...
      </div>

      <template
        v-else-if="selectedConflict"
      >
        <header class="detail-header">
          <div>
            <span class="eyebrow">
              Conflicto
              #{{ selectedConflict.id }}
            </span>

            <h1>
              {{
                selectedConflict.resourceName ||
                `Recurso ${selectedConflict.resourceId}`
              }}
            </h1>

            <p>
              {{
                selectedConflict.items?.length ||
                0
              }}
              elementos involucrados ·
              {{ pendingItems.length }}
              pendientes
            </p>
          </div>

          <span
            class="status-pill large"
            :class="
              selectedConflict.status.toLowerCase()
            "
          >
            {{ selectedConflict.status }}
          </span>
        </header>

        <div
          v-if="store.actionError"
          class="state-card error"
        >
          {{ store.actionError }}
        </div>

        <div
          v-if="store.actionSuccess"
          class="state-card success"
        >
          {{ store.actionSuccess }}
        </div>

        <section
          v-if="
            selectedConflict.resolutionSummary
          "
          class="resolution-summary"
        >
          <ShieldCheck :size="20" />

          <div>
            <strong>
              Resultado
            </strong>

            <p>
              {{
                selectedConflict.resolutionSummary
              }}
            </p>
          </div>
        </section>

        <section class="items-list">
          <article
            v-for="item in selectedConflict.items"
            :key="item.id"
            class="item-card"
          >
            <div class="item-heading">
              <div>
                <span class="entity-type">
                  {{
                    entityLabels[
                      item.entityType
                    ] ||
                    item.entityType
                  }}
                </span>

                <h2>
                  {{
                    item.title ||
                    `Elemento ${item.id}`
                  }}
                </h2>

                <p v-if="item.unitName">
                  {{ item.unitName }}
                </p>
              </div>

              <span
                class="resolution-pill"
                :class="
                  item.resolution
                    .toLowerCase()
                "
              >
                {{
                  resolutionLabels[
                    item.resolution
                  ] ||
                  item.resolution
                }}
              </span>
            </div>

            <div class="schedule">
              <CalendarClock :size="18" />

              <span>
                {{
                  formatDateTime(
                    item.occurrenceStart
                  )
                }}
              </span>

              <span>→</span>

              <span>
                {{
                  formatDateTime(
                    item.occurrenceEnd
                  )
                }}
              </span>
            </div>

            <div
              v-if="
                item.resolution !==
                'PENDING'
              "
              class="resolved-info"
            >
              <Check :size="18" />

              <div>
                <strong>
                  {{
                    resolutionLabels[
                      item.resolution
                    ]
                  }}
                </strong>

                <p v-if="item.resolutionNote">
                  {{ item.resolutionNote }}
                </p>

                <small
                  v-if="item.resolvedAt"
                >
                  {{
                    formatDateTime(
                      item.resolvedAt
                    )
                  }}
                </small>
              </div>
            </div>

            <template v-else>

              <label class="field">
                <span>
                  Nota administrativa
                </span>

                <textarea
                  v-model="
                    ensureForm(item.id)
                      .resolutionNote
                  "
                  rows="3"
                  placeholder="Justifica la decisión tomada"
                />
              </label>

              <details class="reschedule-box">
                <summary>
                  <Clock3 :size="17" />
                  Reprogramar ocurrencia
                </summary>

                <div class="reschedule-fields">
                  <label class="field">
                    <span>Nueva fecha</span>

                    <input
                      v-model="
                        ensureForm(item.id)
                          .newDate
                      "
                      type="date"
                    >
                  </label>

                  <label class="field">
                    <span>Inicio</span>

                    <input
                      v-model="
                        ensureForm(item.id)
                          .newStartTime
                      "
                      type="time"
                    >
                  </label>

                  <label class="field">
                    <span>Término</span>

                    <input
                      v-model="
                        ensureForm(item.id)
                          .newEndTime
                      "
                      type="time"
                    >
                  </label>
                </div>

                <button
                  type="button"
                  class="decision-button reschedule"
                  :disabled="
                    store.resolvingItemId ===
                    item.id
                  "
                  @click="
                    prepareResolution(
                      item,
                      'RESCHEDULE'
                    )
                  "
                >
                  Reprogramar
                </button>
              </details>

              <div class="decisions">
                <button
                  type="button"
                  class="decision-button keep"
                  :disabled="
                    store.resolvingItemId ===
                    item.id
                  "
                  @click="
                    prepareResolution(
                      item,
                      'KEEP'
                    )
                  "
                >
                  <Check :size="17" />
                  Mantener
                </button>

                <button
                  type="button"
                  class="decision-button allow"
                  :disabled="
                    store.resolvingItemId ===
                    item.id
                  "
                  @click="
                    prepareResolution(
                      item,
                      'ALLOW'
                    )
                  "
                >
                  <ShieldCheck :size="17" />
                  Permitir
                </button>

                <button
                  type="button"
                  class="decision-button cancel"
                  :disabled="
                    store.resolvingItemId ===
                    item.id
                  "
                  @click="
                    prepareResolution(
                      item,
                      'CANCEL'
                    )
                  "
                >
                  <CircleOff :size="17" />
                  Cancelar
                </button>
              </div>

            </template>
          </article>
        </section>
      </template>

    </template>

    <Teleport to="body">
      <div
        v-if="pendingDecision"
        class="confirmation-overlay"
        @click.self="cancelPendingDecision"
      >
        <section
          class="confirmation-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="conflict-confirmation-title"
        >
          <span class="eyebrow">
            Decisión administrativa
          </span>

          <h2 id="conflict-confirmation-title">
            Confirmar
            {{
              resolutionLabels[
                pendingDecision.resolution
              ]
            }}
          </h2>

          <p class="confirmation-description">
            {{
              confirmationMessages[
                pendingDecision.resolution
              ]
            }}
          </p>

          <div class="confirmation-item">
            <strong>
              {{
                pendingDecision.item.title ||
                `Elemento ${pendingDecision.item.id}`
              }}
            </strong>

            <span
              v-if="pendingDecision.item.unitName"
            >
              {{ pendingDecision.item.unitName }}
            </span>

            <small>
              {{
                formatDateTime(
                  pendingDecision.item.occurrenceStart
                )
              }}
              →
              {{
                formatDateTime(
                  pendingDecision.item.occurrenceEnd
                )
              }}
            </small>
          </div>

          <div class="confirmation-note">
            <strong>
              Nota administrativa
            </strong>

            <p>
              {{
                pendingDecision.payload.resolutionNote
              }}
            </p>
          </div>

          <div
            v-if="
              pendingDecision.resolution ===
              'RESCHEDULE'
            "
            class="confirmation-note"
          >
            <strong>
              Nuevo horario
            </strong>

            <p>
              {{
                pendingDecision.payload.newDate
              }}
              ·
              {{
                pendingDecision.payload.newStartTime
              }}
              →
              {{
                pendingDecision.payload.newEndTime
              }}
            </p>
          </div>

          <p class="irreversible-warning">
            Esta decisión quedará registrada y no podrá editarse
            posteriormente desde este conflicto.
          </p>

          <div class="confirmation-actions">
            <button
              type="button"
              class="secondary-button"
              :disabled="
                store.resolvingItemId ===
                pendingDecision.item.id
              "
              @click="cancelPendingDecision"
            >
              Volver
            </button>

            <button
              type="button"
              class="primary-button confirm-button"
              :disabled="
                store.resolvingItemId ===
                pendingDecision.item.id
              "
              @click="confirmResolution"
            >
              Confirmar decisión
            </button>
          </div>
        </section>
      </div>
    </Teleport>

  </main>
</template>

<style scoped>
.conflicts-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.page-header,
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
}

h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 30px;
  font-weight: 900;
}

.page-header p,
.detail-header p {
  margin-top: 8px;
  color: var(--color-text-muted);
}

.filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-button,
.refresh-button,
.back-button {
  min-height: 40px;
  padding: 0 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text);
  font: inherit;
  font-weight: 800;
  cursor: pointer;
}

.refresh-button,
.back-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.filter-button.active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.conflict-grid {
  display: grid;
  grid-template-columns:
    repeat(auto-fill, minmax(310px, 1fr));
  gap: 16px;
}

.conflict-card,
.item-card,
.resolution-summary {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  background: var(--color-surface);
  box-shadow: var(--shadow-card);
}

.conflict-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 18px;
}

.card-heading,
.item-heading {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.card-heading > div:nth-child(2),
.item-heading > div {
  flex: 1;
}

.conflict-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-lg);
  background: var(--color-error-soft);
  color: var(--color-error);
}

.eyebrow,
.entity-type {
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 900;
  text-transform: uppercase;
}

.card-heading h2,
.item-heading h2 {
  margin: 3px 0 0;
  color: var(--color-text);
  font-size: 18px;
}

.item-heading p {
  margin: 5px 0 0;
  color: var(--color-text-muted);
}

.status-pill,
.resolution-pill {
  flex: 0 0 auto;
  padding: 6px 10px;
  border-radius: var(--radius-pill);
  font-size: 11px;
  font-weight: 900;
}

.status-pill.pending,
.resolution-pill.pending {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.status-pill.resolved,
.resolution-pill.keep,
.resolution-pill.allow {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.resolution-pill.cancel {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.resolution-pill.reschedule {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.card-body {
  flex: 1;
  color: var(--color-text-muted);
}

.card-body p {
  margin: 0 0 8px;
}

.summary {
  font-size: 13px;
}

.primary-button {
  min-height: 42px;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
  background: var(--color-primary);
  color: var(--color-primary-contrast);
  font: inherit;
  font-weight: 900;
  cursor: pointer;
}

.back-button {
  align-self: flex-start;
}

.status-pill.large {
  font-size: 12px;
  padding: 8px 12px;
}

.resolution-summary {
  display: flex;
  gap: 12px;
  padding: 16px;
  color: var(--color-success);
}

.resolution-summary p {
  margin: 4px 0 0;
  color: var(--color-text-muted);
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.item-card {
  padding: 20px;
}

.schedule {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin: 18px 0;
  color: var(--color-text-muted);
  font-weight: 700;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 7px;
  margin-top: 14px;
}

.field span {
  color: var(--color-text);
  font-size: 13px;
  font-weight: 800;
}

.field textarea,
.field input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text);
  padding: 10px 12px;
  font: inherit;
}

.decisions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.decision-button {
  min-height: 40px;
  padding: 0 13px;
  border-radius: var(--radius-md);
  font: inherit;
  font-weight: 900;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  cursor: pointer;
}

.decision-button.keep {
  border: 1px solid var(--color-success);
  background: var(--color-success-soft);
  color: var(--color-success);
}

.decision-button.allow {
  border: 1px solid var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.decision-button.cancel {
  border: 1px solid var(--color-error-border);
  background: var(--color-error-soft);
  color: var(--color-error);
}

.decision-button.reschedule {
  margin-top: 12px;
  border: 1px solid var(--color-primary);
  background: var(--color-primary);
  color: var(--color-primary-contrast);
}

.decision-button:disabled {
  opacity: 0.55;
  cursor: wait;
}

.reschedule-box {
  margin-top: 16px;
  padding: 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.reschedule-box summary {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-primary);
  font-weight: 900;
  cursor: pointer;
}

.reschedule-fields {
  display: grid;
  grid-template-columns:
    repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.resolved-info {
  display: flex;
  gap: 10px;
  padding: 14px;
  border-radius: var(--radius-lg);
  background: var(--color-surface-soft);
  color: var(--color-success);
}

.resolved-info p {
  margin: 5px 0;
  color: var(--color-text-muted);
}

.state-card {
  border-radius: var(--radius-lg);
}

.state-card.error {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.state-card.success {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.confirmation-overlay {
  position: fixed;
  inset: 0;
  z-index: 1200;

  display: grid;
  place-items: center;

  padding: 20px;

  background: rgba(15, 23, 42, 0.52);
}

.confirmation-modal {
  width: min(520px, 100%);

  box-sizing: border-box;

  padding: 24px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);

  background: var(--color-surface);
  color: var(--color-text);

  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.28);
}

.confirmation-modal h2 {
  margin: 6px 0 10px;
}

.confirmation-description {
  color: var(--color-text-muted);
}

.confirmation-item,
.confirmation-note {
  display: flex;
  flex-direction: column;
  gap: 5px;

  margin-top: 16px;
  padding: 13px;

  border-radius: var(--radius-lg);
  background: var(--color-surface-soft);
}

.confirmation-item span,
.confirmation-item small,
.confirmation-note p {
  margin: 0;
  color: var(--color-text-muted);
}

.irreversible-warning {
  margin-top: 16px;
  padding: 12px;

  border: 1px solid var(--color-error-border);
  border-radius: var(--radius-lg);

  background: var(--color-error-soft);
  color: var(--color-error);

  font-size: 13px;
  font-weight: 800;
}

.confirmation-actions {
  display: flex;
  justify-content: flex-end;
  gap: 9px;

  margin-top: 20px;
}

.secondary-button {
  min-height: 42px;
  padding: 0 15px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  background: var(--color-surface);
  color: var(--color-text);

  font: inherit;
  font-weight: 900;

  cursor: pointer;
}

.confirm-button {
  padding: 0 16px;
}


@media (max-width: 768px) {
  .page-header,
  .detail-header {
    flex-direction: column;
  }

  .reschedule-fields {
    grid-template-columns: 1fr;
  }

  h1 {
    font-size: 26px;
  }
}
</style>
