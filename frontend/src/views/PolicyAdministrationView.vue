<script setup>
import { computed, onMounted, ref } from 'vue'

import { reservationPolicyService } from '@/services/reservationPolicy.service'
import {
  validatePolicyDraft,
  getPolicyChanges
} from '@/utils/policyAdministration'
import { useResourcesStore } from '@/stores/resources'

const resourcesStore = useResourcesStore()

const loading = ref(false)
const error = ref(null)
const history = ref([])
const draft = ref(null)
const durationsText = ref('')
const screen = ref('current')
const step = ref('general')
const discardTarget = ref(null)

const fields = [
  {
    key: 'reservableWindowDays',
    label: 'Ventana de reserva',
    unit: 'días',
    min: 1
  },
  {
    key: 'requestFrequencyDays',
    label: 'Frecuencia de solicitudes',
    unit: 'días',
    min: 1
  },
  {
    key: 'confirmationDeadlineMinutes',
    label: 'Plazo de confirmación',
    unit: 'minutos',
    min: 0
  },
  {
    key: 'minimumParticipants',
    label: 'Mínimo global (compatibilidad)',
    unit: 'participantes',
    min: 1
  },
  {
    key: 'slotIntervalMinutes',
    label: 'Intervalo de horarios',
    unit: 'minutos',
    min: 1
  }
]

const currentPolicy = computed(() =>
  history.value.find((policy) => policy.effectiveTo == null) || null
)

const resources = computed(() =>
  [...resourcesStore.resources].sort((a, b) => a.id - b.id)
)

const resourceName = (id) =>
  resources.value.find((resource) => resource.id === id)?.name ||
  `Recurso #${id}`

const formatDate = (value) => {
  if (!value) return 'Sin fecha'

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) return 'Fecha no disponible'

  return new Intl.DateTimeFormat('es-CL', {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'America/Santiago'
  }).format(date)
}

const minuteToTime = (minute) => {
  const value = Number(minute)

  if (!Number.isFinite(value)) return ''

  return `${String(Math.floor(value / 60)).padStart(2, '0')}:${String(value % 60).padStart(2, '0')}`
}

const timeToMinute = (value) => {
  const match = /^(\d{2}):(\d{2})$/.exec(String(value))
  if (!match) return null

  return Number(match[1]) * 60 + Number(match[2])
}

const openingTime = computed({
  get: () => minuteToTime(draft.value?.openingMinute),
  set: (value) => {
    if (draft.value) draft.value.openingMinute = timeToMinute(value)
  }
})

const closingTime = computed({
  get: () => minuteToTime(draft.value?.closingMinute),
  set: (value) => {
    if (draft.value) draft.value.closingMinute = timeToMinute(value)
  }
})

const resetDraft = () => {
  if (!currentPolicy.value) {
    draft.value = null
    return
  }

  draft.value = {
    ...currentPolicy.value,
    allowedDurations: [...(currentPolicy.value.allowedDurations || [])],
    resourceIds: [...(currentPolicy.value.resourceIds || [])],
    groupResources: (currentPolicy.value.groupResources || []).map(
      (rule) => ({ ...rule })
    )
  }

  durationsText.value = draft.value.allowedDurations.join(', ')
}

const load = async () => {
  loading.value = true
  error.value = null

  try {
    const [policies] = await Promise.all([
      reservationPolicyService.getHistory(),
      resourcesStore.fetchResources()
    ])

    if (resourcesStore.error) {
      throw new Error(resourcesStore.error)
    }

    if (!Array.isArray(policies)) {
      throw new Error('El historial recibido no es válido.')
    }

    history.value = policies

    if (!currentPolicy.value) {
      throw new Error('No existe una política vigente.')
    }

    resetDraft()
  } catch (err) {
    history.value = []
    draft.value = null
    error.value =
      err.response?.data?.error ||
      err.message ||
      'No se pudieron cargar las políticas.'
  } finally {
    loading.value = false
  }
}

const isAllowed = (id) =>
  draft.value?.resourceIds.includes(id) === true

const groupRule = (id) =>
  draft.value?.groupResources.find((rule) => rule.resourceId === id)

const groupConfigurableResources = computed(() =>
  resources.value.filter((resource) =>
    resource.isActive === true &&
    resource.reservationMode === 'RESERVABLE' &&
    isAllowed(resource.id)
  )
)

const outOfScopeCount = computed(() =>
  resources.value.filter((resource) =>
    !isAllowed(resource.id)
  ).length
)

const setGroup = (resource, enabled) => {
  if (!draft.value) return

  draft.value.groupResources = draft.value.groupResources.filter(
    (rule) => rule.resourceId !== resource.id
  )

  if (enabled) {
    draft.value.groupResources.push({
      resourceId: resource.id,
      minimumParticipants: Number(draft.value.minimumParticipants) || 10
    })

    draft.value.groupResources.sort(
      (a, b) => a.resourceId - b.resourceId
    )
  }
}

const groupDescription = (policy) => {
  const rules = policy.groupResources || []

  return rules.length
    ? rules.map(
      (rule) =>
        `${resourceName(rule.resourceId)}: ${rule.minimumParticipants}`
    ).join(' · ')
    : 'Sin recursos sujetos a confirmación grupal'
}


const review = ref(null)
const publishing = ref(false)
const publicationError = ref(null)
const publicationSuccess = ref(null)

// Conserva la clave para reintentos del mismo contenido.
const pendingAttempt = ref(null)

const hasUnsavedChanges = computed(() => {
  if (!draft.value || !currentPolicy.value) return false

  const keys = [
    'reservableWindowDays', 'requestFrequencyDays',
    'confirmationDeadlineMinutes', 'minimumParticipants',
    'openingMinute', 'closingMinute', 'slotIntervalMinutes'
  ]

  return keys.some((key) =>
    String(draft.value[key]) !== String(currentPolicy.value[key])
  ) ||
    durationsText.value !==
      (currentPolicy.value.allowedDurations || []).join(', ') ||
    JSON.stringify(draft.value.resourceIds) !==
      JSON.stringify(currentPolicy.value.resourceIds) ||
    JSON.stringify(draft.value.groupResources) !==
      JSON.stringify(currentPolicy.value.groupResources)
})

const startEdit = () => {
  resetDraft()
  review.value = null
  publicationError.value = null
  screen.value = 'edit'
  step.value = 'general'
}

const switchScreen = (next) => {
  if (publishing.value || next === screen.value) return

  if (screen.value === 'edit' &&
      (hasUnsavedChanges.value || pendingAttempt.value)) {
    discardTarget.value = next
    return
  }

  if (next === 'edit') {
    startEdit()
  } else {
    screen.value = next
  }
}

const askReset = () => {
  if (publishing.value) return
  if (hasUnsavedChanges.value || pendingAttempt.value) {
    discardTarget.value = 'reset'
  } else {
    resetDraft()
    review.value = null
    step.value = 'general'
  }
}

const discardChanges = () => {
  const target = discardTarget.value
  discardTarget.value = null
  pendingAttempt.value = null
  publicationError.value = null
  review.value = null
  resetDraft()
  step.value = 'general'

  if (target !== 'reset') screen.value = target
}

const modeLabel = (mode) => ({
  RESERVABLE: 'Reservable',
  OPEN_USE: 'Uso libre',
  INFORMATIVE: 'Informativo',
  ADMIN_ONLY: 'Solo administración'
}[mode] || mode)

const draftValidation = computed(() => (
  draft.value
    ? validatePolicyDraft(
        draft.value,
        durationsText.value,
        resources.value
      )
    : { payload: null, errors: [] }
))

const prepareReview = () => {
  publicationError.value = null
  publicationSuccess.value = null
  review.value = null

  const { payload, errors } = draftValidation.value

  if (!payload) {
    publicationError.value =
      errors.join(' ') || 'El borrador contiene errores.'
    return
  }

  const changes = getPolicyChanges(
    currentPolicy.value,
    payload,
    resources.value
  )

  if (!changes.length) {
    publicationError.value =
      'No hay modificaciones respecto a la política vigente.'
    return
  }

  review.value = {
    expectedPolicyId: currentPolicy.value.id,
    payload: JSON.parse(JSON.stringify(payload)),
    changes
  }

  step.value = 'review'
}

const confirmPublication = async () => {
  if (!review.value || publishing.value) return

  publicationError.value = null
  publicationSuccess.value = null

  const selectedReview = review.value

  const publicationPayload = {
    ...selectedReview.payload,
    expectedPolicyId: selectedReview.expectedPolicyId
  }

  const serialized = JSON.stringify(publicationPayload)

  const retryingSamePayload =
    pendingAttempt.value?.serialized === serialized

  const unchanged = () => {
    const { payload } = validatePolicyDraft(
      draft.value,
      durationsText.value,
      resources.value
    )

    return payload &&
      JSON.stringify(payload) ===
        JSON.stringify(selectedReview.payload)
  }

  if (!unchanged()) {
    review.value = null
    publicationError.value =
      'El borrador cambió. Revisa nuevamente las modificaciones.'
    return
  }

  publishing.value = true

  try {
    // Comprobar que seguimos editando la versión consultada.
    if (!retryingSamePayload) {
      const latest = await reservationPolicyService.getHistory()
      const active = latest.find(
        (policy) => policy.effectiveTo == null
      )

      if (active?.id !== selectedReview.expectedPolicyId) {
        review.value = null
        publicationError.value =
          'La política vigente cambió. Actualiza antes de publicar.'
        return
      }
    }

    if (!unchanged()) {
      review.value = null
      publicationError.value =
        'El borrador cambió durante la revisión.'
      return
    }


    if (pendingAttempt.value?.serialized !== serialized) {
      pendingAttempt.value = {
        serialized,
        key: crypto.randomUUID()
      }
    }

    await reservationPolicyService.publish(
      publicationPayload,
      pendingAttempt.value.key
    )

    pendingAttempt.value = null
    review.value = null

    await load()
    screen.value = 'current'
    step.value = 'general'

    publicationSuccess.value = error.value
      ? 'Política publicada. No se pudo actualizar la vista; vuelve a cargarla.'
      : 'Nueva política publicada correctamente.'
  } catch (err) {
    // No descartar la clave: una respuesta perdida puede
    // ocurrir después de que PostgreSQL confirmó la operación.
    publicationError.value =
      err.response?.data?.error ||
      'No se pudo confirmar la publicación. Revisa el historial antes de reintentar.'
  } finally {
    publishing.value = false
  }
}

onMounted(load)
</script>

<template>
  <main class="policy-page">
    <header class="page-header">
      <div>
        <h1>Reglas de reserva</h1>
        <p>
          Consulta las reglas vigentes o prepara una nueva versión
          para las próximas solicitudes.
        </p>
      </div>

      <button
        v-if="screen !== 'edit'"
        type="button"
        class="app-button secondary"
        :disabled="loading"
        @click="load"
      >
        Actualizar
      </button>
    </header>

    <div v-if="loading" class="panel" role="status">
      Cargando políticas y recursos...
    </div>

    <div v-else-if="error" class="panel error" role="alert">
      {{ error }}
    </div>

    <template v-else-if="draft && currentPolicy">
      <nav class="policy-tabs" aria-label="Administrar políticas">
        <button type="button"
          :class="{ active: screen === 'current' }"
          :aria-pressed="screen === 'current'"
          :disabled="publishing"
          @click="switchScreen('current')">
          Política vigente
        </button>
        <button type="button"
          :class="{ active: screen === 'edit' }"
          :aria-pressed="screen === 'edit'"
          :disabled="publishing"
          @click="switchScreen('edit')">
          Configurar
        </button>
        <button type="button"
          :class="{ active: screen === 'history' }"
          :aria-pressed="screen === 'history'"
          :disabled="publishing"
          @click="switchScreen('history')">
          Historial
        </button>
      </nav>

      <div v-if="discardTarget" class="state-card warning discard-panel"
        role="alert">
        <strong>Hay cambios sin publicar</strong>
        <p>Si continúas, perderás la edición actual.</p>
        <div class="editor-actions">
          <button type="button" class="app-button secondary"
            @click="discardTarget = null">
            Seguir editando
          </button>
          <button type="button" class="app-button danger"
            @click="discardChanges">
            Descartar borrador
          </button>
        </div>
      </div>

      <div v-if="publicationSuccess" class="state-card success"
        role="status">
        {{ publicationSuccess }}
      </div>

      <div v-if="screen === 'edit'" class="workflow">
        <div class="workflow-title">
          <h2>Nueva versión de reglas</h2>
          <p>
            Los cambios no se aplican hasta confirmar la publicación.
          </p>
        </div>
        <ol class="workflow-steps" aria-label="Etapas de configuración">
          <li :class="{ active: step === 'general' }">
            <span>1</span> Reglas generales
          </li>
          <li :class="{ active: step === 'resources' }">
            <span>2</span> Instalaciones
          </li>
          <li :class="{ active: step === 'review' }">
            <span>3</span> Revisar y publicar
          </li>
        </ol>
        <div v-if="publicationError" class="state-card error"
          role="alert">
          {{ publicationError }}
        </div>
      </div>

      <section v-if="screen === 'current'" class="policy-grid">
        <article class="panel">
          <h2>Política vigente</h2>

          <p class="muted">
            Versión #{{ currentPolicy.id }}
            · Vigente desde {{ formatDate(currentPolicy.effectiveFrom) }}
          </p>

          <dl>
            <div v-for="field in fields" :key="field.key" class="fact">
              <dt>{{ field.label }}</dt>
              <dd>
                {{ currentPolicy[field.key] }} {{ field.unit }}
              </dd>
            </div>

            <div class="fact">
              <dt>Jornada</dt>
              <dd>
                {{ minuteToTime(currentPolicy.openingMinute) }}
                –
                {{ minuteToTime(currentPolicy.closingMinute) }}
              </dd>
            </div>

            <div class="fact">
              <dt>Duraciones</dt>
              <dd>
                {{ currentPolicy.allowedDurations.join(', ') }} min
              </dd>
            </div>

            <div class="fact">
              <dt>Recursos permitidos</dt>
              <dd>{{ currentPolicy.resourceIds.length }}</dd>
            </div>
          </dl>

          <h3>Confirmación grupal</h3>
          <p>{{ groupDescription(currentPolicy) }}</p>

          <div class="editor-actions current-actions">
            <button type="button" class="app-button primary"
              @click="startEdit">
              Crear nueva versión
            </button>
            <button type="button" class="app-button secondary"
              @click="switchScreen('history')">
              Ver historial
            </button>
          </div>
        </article>
      </section>

      <section
        v-if="screen === 'edit' && step === 'general'"
        class="panel policy-editor"
      >
          <div class="section-header">
            <div>
              <h2>Configura las reglas generales</h2>
              <p class="muted">
                Define cuándo y cómo se podrán solicitar las
                instalaciones deportivas.
              </p>
            </div>

            <button type="button" class="app-button secondary"
              @click="askReset">
              Restablecer borrador
            </button>
          </div>

          <div class="form-grid">
            <label
              v-for="field in fields.filter(
                item => item.key !== 'minimumParticipants'
              )"
              :key="field.key"
            >
              {{ field.label }} ({{ field.unit }})
              <input
                v-model.number="draft[field.key]"
                type="number"
                :min="field.min"
                step="1"
              />
            </label>

            <label>
              Apertura
              <input v-model="openingTime" type="time" />
            </label>

            <label>
              Cierre
              <input v-model="closingTime" type="time" />
            </label>

            <label class="wide">
              Duraciones permitidas (minutos, separadas por coma)
              <input
                v-model="durationsText"
                type="text"
                placeholder="30, 45, 60, 90"
              />
            </label>
          </div>

          <details class="advanced-settings">
            <summary>Configuración avanzada</summary>
            <label class="advanced-field">
              Mínimo global de participantes (compatibilidad)
              <input v-model.number="draft.minimumParticipants"
                type="number" min="1" step="1" />
            </label>
            <p class="muted">
              Los mínimos particulares por instalación prevalecen
              sobre este valor general.
            </p>
          </details>

          <div class="editor-actions">
            <button type="button" class="app-button secondary"
              @click="switchScreen('current')">
              Descartar edición
            </button>
            <button type="button" class="app-button primary"
              @click="step = 'resources'">
              Continuar a recursos
            </button>
          </div>
      </section>

      <section
        v-if="screen === 'edit' && step === 'resources'"
        class="panel"
      >
        <h2>Confirmación grupal por instalación</h2>
        <p class="muted">
          Configura los mínimos grupales de las instalaciones
          reservables ya incluidas en la política vigente.
          El modo y el estado se administran en Instalaciones deportivas.
        </p>

        <div class="resource-table-wrap">
          <table>
            <thead>
              <tr>
                <th>Recurso</th>

                <th>Capacidad</th>

                <th>Grupal</th>
                <th>Mínimo</th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="resource in groupConfigurableResources" :key="resource.id">
                <td>
                  <strong>{{ resource.name }}</strong>
                  <small v-if="!resource.isActive">
                    Inactivo
                  </small>
                </td>


                <td>{{ resource.capacity ?? 'Sin definir' }}</td>



                <td>
                  <input
                    type="checkbox"
                    :checked="!!groupRule(resource.id)"
                    :disabled="
                      !isAllowed(resource.id) ||
                      !resource.isActive ||
                      resource.reservationMode !== 'RESERVABLE'
                    "
                    :aria-label="`Confirmación grupal ${resource.name}`"
                    @change="setGroup(resource, $event.target.checked)"
                  />
                </td>

                <td>
                  <input
                    v-if="groupRule(resource.id)"
                    v-model.number="groupRule(resource.id).minimumParticipants"
                    type="number"
                    min="1"
                    :max="resource.capacity ?? undefined"
                    :aria-label="`Mínimo de participantes ${resource.name}`"
                  />
                  <span v-else>—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <p v-if="!groupConfigurableResources.length" class="muted">
          No hay instalaciones reservables habilitadas
          para configurar confirmación grupal.
        </p>

        <p class="muted">
          Las instalaciones de uso libre conservan su comportamiento
          actual y no requieren mínimos grupales.
        </p>

        <p v-if="outOfScopeCount > 0"
          class="state-card warning" role="note">
          Hay {{ outOfScopeCount }} instalación(es) fuera del alcance
          de la política vigente. Esta pantalla no las habilita
          automáticamente; esa simplificación queda pendiente para V2.
        </p>

        <p class="muted">
          La confirmación grupal solo puede utilizarse en
          instalaciones reservables con capacidad suficiente.
        </p>

        <div class="editor-actions">
          <button type="button" class="app-button secondary"
            @click="step = 'general'">
            Volver a reglas
          </button>
          <button type="button" class="app-button primary"
            @click="prepareReview">
            Revisar cambios
          </button>
        </div>
      </section>

      <section
        v-if="screen === 'edit' && step === 'review'"
        class="panel"
      >
        <h2>Revisa antes de publicar</h2>

        <p class="muted">
          Los cambios se aplicarán a las nuevas reservas.
          Las reservas existentes conservarán su política original.
        </p>

        <div v-if="review" class="review">
          <h3>Cambios propuestos</h3>

          <div class="resource-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Parámetro</th>
                  <th>Anterior</th>
                  <th>Nuevo</th>
                </tr>
              </thead>

              <tbody>
                <tr
                  v-for="change in review.changes"
                  :key="change.label"
                >
                  <td>{{ change.label }}</td>
                  <td>{{ change.before }}</td>
                  <td><strong>{{ change.after }}</strong></td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="state-card warning impact-message">
            <strong>Impacto de la publicación</strong>
            <p>
              La nueva versión tendrá vigencia inmediata para
              las solicitudes nuevas. Las reservas existentes
              conservarán su política original.
            </p>
          </div>

          <div class="review-actions">
            <button
              type="button"
              :disabled="publishing"
              class="app-button secondary"
              @click="step = 'resources'; review = null"
            >
              Volver a editar
            </button>

            <button
              type="button"
              class="app-button primary publish-button"
              :disabled="publishing"
              @click="confirmPublication"
            >
              {{
                publishing
                  ? 'Publicando...'
                  : 'Confirmar y publicar nueva versión'
              }}
            </button>
          </div>
        </div>
      </section>

      <section v-if="screen === 'history'" class="panel">
        <h2>Historial de políticas</h2>
        <p class="muted">
          Consulta las versiones publicadas, sus vigencias
          y la configuración grupal registrada.
        </p>

        <details
          v-for="policy in history"
          :key="policy.id"
          class="history-item"
        >
          <summary>
            Versión #{{ policy.id }}
            · {{ formatDate(policy.effectiveFrom) }}
            · {{ policy.effectiveTo == null ? 'Vigente' : 'Histórica' }}
          </summary>

          <p>
            Fin de vigencia:
            {{ policy.effectiveTo ? formatDate(policy.effectiveTo) : 'Actual' }}
          </p>

          <p>Creada por usuario ID: {{ policy.createdByUserId ?? 'Sistema' }}</p>

          <p>
            Ventana: {{ policy.reservableWindowDays }} días ·
            Frecuencia: {{ policy.requestFrequencyDays }} días ·
            Confirmación: {{ policy.confirmationDeadlineMinutes }} min
          </p>

          <p>
            Recursos permitidos:
            {{ policy.resourceIds.map(resourceName).join(', ') || 'Ninguno' }}
          </p>

          <p>{{ groupDescription(policy) }}</p>
        </details>
      </section>
    </template>
  </main>
</template>

<style scoped>
.policy-page {
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  color: var(--color-text);
}

.page-header,
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.page-header h1 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-page-title);
  font-weight: 800;
}

.page-header p,
.muted,
.workflow-title p,
.history-item p {
  color: var(--color-text-muted);
  line-height: 1.5;
}

.policy-tabs {
  display: flex;
  gap: var(--space-1);
  padding: var(--space-1);
  width: fit-content;
  max-width: 100%;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-soft);
}

.policy-tabs button {
  background: transparent;
  border: 0;
  border-radius: var(--radius-md);
  min-height: 42px;
  padding: 0 var(--space-4);
  font: inherit;
  font-weight: 700;
  color: var(--color-text-muted);
  cursor: pointer;
}

.policy-tabs button.active {
  background: var(--color-surface);
  color: var(--color-primary-strong);
  box-shadow: var(--shadow-card);
}

.policy-tabs button:hover:not(.active) {
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
}

.policy-tabs button:disabled {
  cursor: not-allowed;
  opacity: .6;
}

.policy-grid { min-width: 0; }

.panel {
  min-width: 0;
  padding: var(--space-5);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-card);
}

.panel h2,
.workflow-title h2 {
  margin: 0 0 var(--space-2);
  color: var(--color-text);
  font-size: var(--text-section-title);
  font-weight: 800;
}

.panel h3 { font-size: var(--text-body); }

.fact {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-border-soft);
}

.fact dt { color: var(--color-text-muted); }
.fact dd { margin: 0; font-weight: 750; }

.workflow {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.workflow-steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-2);
}

.workflow-steps li {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  padding: var(--space-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-weight: 650;
}

.workflow-steps li.active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
}

.workflow-steps span {
  display: inline-grid;
  place-items: center;
  width: 26px;
  height: 26px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--color-surface-soft);
  font-weight: 800;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
  margin: var(--space-4) 0;
}

.form-grid label,
.advanced-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
  font-size: var(--text-help);
  font-weight: 700;
}

.wide { grid-column: 1 / -1; }

.form-grid input,
.advanced-field input,
.resource-table-wrap input[type="number"] {
  width: 100%;
  min-width: 0;
  min-height: 42px;
  padding: 0 var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text);
  font: inherit;
}

.advanced-settings {
  margin: var(--space-4) 0;
  padding: var(--space-4);
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-md);
  background: var(--color-surface-muted);
}

.advanced-settings summary {
  cursor: pointer;
  font-weight: 750;
  color: var(--color-primary-strong);
}

.advanced-field {
  margin-top: var(--space-3);
  max-width: 360px;
}

.editor-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-5);
}

.resource-table-wrap {
  width: 100%;
  overflow-x: auto;
  margin-top: var(--space-4);
}

table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

th, td {
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border-soft);
}

th {
  color: var(--color-text-muted);
  font-size: var(--text-help);
}

td small {
  display: block;
  color: var(--color-error);
}

td input[type="number"] {
  max-width: 115px;
}

td input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.review { margin-top: var(--space-4); }

.impact-message,
.discard-panel {
  margin: var(--space-4) 0;
}

.impact-message p,
.discard-panel p { margin-bottom: 0; }

.history-item {
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--color-border-soft);
}

.history-item summary {
  cursor: pointer;
  font-weight: 750;
  color: var(--color-primary-strong);
}

.policy-page button:focus-visible,
.policy-page summary:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 3px;
}

@media (max-width: 720px) {
  .workflow-steps { grid-template-columns: 1fr; }
  .form-grid { grid-template-columns: 1fr; }
  .wide { grid-column: auto; }
  .policy-tabs { width: 100%; }
  .policy-tabs button { flex: 1; padding: var(--space-2); }
  .panel { padding: var(--space-4); }
}

@media (max-width: 480px) {
  .page-header > button { width: 100%; }
  .editor-actions .app-button { flex: 1 1 100%; }
}
</style>
