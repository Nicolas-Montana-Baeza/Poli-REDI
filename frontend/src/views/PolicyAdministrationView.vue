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

const setAllowed = (id, enabled) => {
  if (!draft.value) return

  if (enabled) {
    draft.value.resourceIds = [
      ...new Set([...draft.value.resourceIds, id])
    ].sort((a, b) => a - b)
  } else {
    draft.value.resourceIds = draft.value.resourceIds.filter(
      (value) => value !== id
    )

    draft.value.groupResources = draft.value.groupResources.filter(
      (rule) => rule.resourceId !== id
    )
  }
}

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
        <h1>Políticas de reserva</h1>
        <p>
          Administración de reglas institucionales y versiones históricas.
        </p>
      </div>

      <button type="button" :disabled="loading" @click="load">
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
      <section class="policy-grid">
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
        </article>

        <article class="panel">
          <div class="section-header">
            <div>
              <h2>Borrador de nueva política</h2>
              <p class="muted">
                Edición local. Todavía no se publica ni guarda.
              </p>
            </div>

            <button type="button" @click="resetDraft">
              Restablecer
            </button>
          </div>

          <div class="form-grid">
            <label v-for="field in fields" :key="field.key">
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

          <p class="muted">
            Los mínimos por recurso prevalecerán sobre el mínimo
            global cuando se publique la nueva versión.
          </p>
        </article>
      </section>

      <section class="panel">
        <h2>Recursos de la nueva política</h2>
        <p class="muted">
          Selecciona los recursos permitidos y cuáles requieren
          confirmación grupal.
        </p>

        <div class="resource-table-wrap">
          <table>
            <thead>
              <tr>
                <th>Recurso</th>
                <th>Modo</th>
                <th>Capacidad</th>
                <th>Permitido</th>
                <th>Grupal</th>
                <th>Mínimo</th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="resource in resources" :key="resource.id">
                <td>
                  <strong>{{ resource.name }}</strong>
                  <small v-if="!resource.isActive">
                    Inactivo
                  </small>
                </td>

                <td>{{ resource.reservationMode }}</td>
                <td>{{ resource.capacity ?? 'Sin definir' }}</td>

                <td>
                  <input
                    type="checkbox"
                    :checked="isAllowed(resource.id)"
                    :disabled="!resource.isActive"
                    :aria-label="`Permitir ${resource.name}`"
                    @change="setAllowed(resource.id, $event.target.checked)"
                  />
                </td>

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

        <p v-if="!resources.length" class="muted">
          No hay recursos registrados.
        </p>
      </section>

      <section class="panel">
        <h2>Comparación y publicación</h2>

        <p class="muted">
          Los cambios se aplicarán a las nuevas reservas.
          Las reservas existentes conservarán su política original.
        </p>

        <p
          v-if="publicationError"
          role="alert"
          class="publication-error"
        >
          {{ publicationError }}
        </p>

        <p
          v-if="publicationSuccess"
          role="status"
          class="publication-success"
        >
          {{ publicationSuccess }}
        </p>

        <button
          type="button"
          :disabled="publishing"
          @click="prepareReview"
        >
          Revisar cambios
        </button>

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

          <p class="muted">
            Confirma únicamente si estos cambios corresponden
            a la nueva política institucional.
          </p>

          <div class="review-actions">
            <button
              type="button"
              :disabled="publishing"
              @click="review = null"
            >
              Cancelar
            </button>

            <button
              type="button"
              class="publish-button"
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

      <section class="panel">
        <h2>Historial de políticas</h2>

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
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.page-header,
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

h1, h2, h3 {
  color: #0f172a;
}

h1 { margin: 0; font-size: 30px; }
h2 { margin: 0 0 10px; font-size: 20px; }
h3 { font-size: 16px; }

.muted,
.page-header p,
.history-item p {
  color: #64748b;
}

.policy-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.panel {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 22px;
  min-width: 0;
}

.panel.error {
  color: #b91c1c;
  border-color: #fca5a5;
}

.fact {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid #e2e8f0;
  padding: 10px 0;
}

.fact dt { color: #64748b; }
.fact dd { margin: 0; font-weight: 700; }

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-weight: 600;
  font-size: 14px;
}

.wide { grid-column: 1 / -1; }

input[type="number"],
input[type="time"],
input[type="text"] {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  padding: 9px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
}

button {
  cursor: pointer;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: white;
  padding: 10px 14px;
}

button:disabled { cursor: not-allowed; opacity: .6; }

.resource-table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

th, td {
  padding: 12px;
  border-bottom: 1px solid #e2e8f0;
}

td small {
  display: block;
  color: #b91c1c;
}

td input[type="number"] {
  max-width: 110px;
}

.history-item {
  padding: 13px 0;
  border-bottom: 1px solid #e2e8f0;
}

.history-item summary {
  cursor: pointer;
  font-weight: 700;
}

@media (max-width: 900px) {
  .policy-grid { grid-template-columns: 1fr; }
}

@media (max-width: 520px) {
  .form-grid { grid-template-columns: 1fr; }
  .wide { grid-column: auto; }
}

.review {
  margin-top: 18px;
}

.review-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
}

.publish-button {
  background: #1d4ed8;
  color: white;
  border-color: #1d4ed8;
}

.publication-error {
  color: #b91c1c;
}

.publication-success {
  color: #166534;
}

</style>
