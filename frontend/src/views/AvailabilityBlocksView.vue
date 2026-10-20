<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { CalendarOff, Plus, RefreshCcw, ShieldAlert } from 'lucide-vue-next'
import { useResourcesStore } from '@/stores/resources'
import { useAvailabilityBlocksStore } from '@/stores/availabilityBlocks'
import {
  chileDateKey, shiftDateKey, santiagoLocalToISOString,
  formatChileBlockTime
} from '@/utils/availabilityBlockTime'

const resourcesStore = useResourcesStore()
const blocksStore = useAvailabilityBlocksStore()

const initialDate = chileDateKey()
const startDate = ref(initialDate)
const endDate = ref(shiftDateKey(initialDate, 14))
const resourceFilter = ref('ALL')
const statusFilter = ref('ACTIVE')
const showCreate = ref(false)
const formError = ref('')
const filterError = ref('')
const confirmDeactivateId = ref(null)

const types = [
  { value: 'MAINTENANCE', label: 'Mantención' },
  { value: 'ADMINISTRATIVE', label: 'Administrativo' },
  { value: 'EVENT', label: 'Evento' },
  { value: 'CLOSED', label: 'Cierre' },
  { value: 'OTHER', label: 'Otro' }
]
const typeLabel = (value) => types.find(type => type.value === value)?.label || value

const form = reactive({
  resourceId: '', blockType: 'MAINTENANCE', reason: '',
  startTime: `${initialDate}T10:00`, endTime: `${initialDate}T11:00`
})

const activeResources = computed(() => resourcesStore.resources.filter(
  resource => resource.isActive !== false
))

const visibleBlocks = computed(() => blocksStore.blocks.filter(block => {
  if (resourceFilter.value !== 'ALL' && String(block.resourceId) !== resourceFilter.value) return false
  if (statusFilter.value === 'ACTIVE' && !block.isActive) return false
  if (statusFilter.value === 'INACTIVE' && block.isActive) return false
  return true
}).sort((a, b) => new Date(a.startTime) - new Date(b.startTime)))

const range = () => {
  // Vista inclusiva en pantalla; backend usa limite superior exclusivo.
  const afterEnd = shiftDateKey(endDate.value, 1)
  shiftDateKey(startDate.value, 0)
  const fromMs = Date.parse(`${startDate.value}T12:00:00Z`)
  const toMs = Date.parse(`${afterEnd}T12:00:00Z`)
  if (!Number.isFinite(fromMs) || !Number.isFinite(toMs) ||
      toMs <= fromMs || (toMs - fromMs) > 31 * 86400000) {
    throw new Error('Selecciona un rango válido de 1 a 31 días.')
  }
  return { from: startDate.value, to: afterEnd }
}

const loadBlocks = async () => {
  filterError.value = ''
  try {
    await blocksStore.fetchBlocks(range())
  } catch (error) {
    filterError.value = error.message
  }
}

onMounted(() => {
  resourcesStore.fetchResources()
  loadBlocks()
})

const openCreate = () => {
  blocksStore.actionError = null
  blocksStore.actionSuccess = null
  formError.value = ''
  form.resourceId = ''
  form.blockType = 'MAINTENANCE'
  form.reason = ''
  form.startTime = `${startDate.value}T10:00`
  form.endTime = `${startDate.value}T11:00`
  showCreate.value = true
}

const createBlock = async () => {
  if (blocksStore.saving) return
  formError.value = ''
  blocksStore.actionError = null
  const resourceId = Number(form.resourceId)
  if (!Number.isSafeInteger(resourceId) || resourceId <= 0 ||
      !activeResources.value.some(resource => resource.id === resourceId)) {
    formError.value = 'Selecciona una instalación activa.'
    return
  }
  if ([...form.reason.trim()].length > 500) {
    formError.value = 'El motivo admite un máximo de 500 caracteres.'
    return
  }
  try {
    const startTime = santiagoLocalToISOString(form.startTime)
    const endTime = santiagoLocalToISOString(form.endTime)
    const duration = new Date(endTime) - new Date(startTime)
    if (duration <= 0 || duration > 31 * 86400000) {
      throw new Error('La hora de término debe ser posterior al inicio (máximo 31 días).')
    }
    await blocksStore.createBlock({
      resourceId, blockType: form.blockType, reason: form.reason.trim(),
      startTime, endTime
    })
    showCreate.value = false
    // La creación ya está confirmada. Una falla al refrescar no la revierte.
    await loadBlocks()
  } catch (error) {
    formError.value = error.message
  }
}

const deactivate = async (id) => {
  try {
    await blocksStore.deactivateBlock(id)
    confirmDeactivateId.value = null
  } catch {
    // El store proporciona el mensaje de error visible.
  }
}
</script>

<template>
  <main class="availability-blocks-view">
    <header class="page-header">
      <div>
        <h1>Bloqueos de disponibilidad</h1>
        <p>Gestiona cierres, mantenciones y otras indisponibilidades de instalaciones.</p>
      </div>
      <button
        type="button"
        class="app-button primary"
        :aria-expanded="showCreate"
        :disabled="blocksStore.saving"
        @click="showCreate ? showCreate = false : openCreate()"
      >
        <Plus :size="18" /> {{ showCreate ? 'Cerrar formulario' : 'Nuevo bloqueo' }}
      </button>
    </header>

    <section v-if="showCreate" class="app-card create-panel" aria-label="Nuevo bloqueo">
      <div class="panel-heading">
        <h2>Nuevo bloqueo</h2>
        <p>Los horarios se interpretan en America/Santiago. No se podrán reservar mientras el bloqueo esté activo.</p>
      </div>
      <form class="block-form" @submit.prevent="createBlock">
        <label class="form-field">
          Instalación
          <select v-model="form.resourceId" required :disabled="blocksStore.saving || resourcesStore.loading">
            <option value="" disabled>Selecciona una instalación</option>
            <option v-for="resource in activeResources" :key="resource.id" :value="String(resource.id)">
              {{ resource.name }}
            </option>
          </select>
        </label>
        <label class="form-field">
          Motivo del bloqueo
          <select v-model="form.blockType" required :disabled="blocksStore.saving">
            <option v-for="type in types" :key="type.value" :value="type.value">{{ type.label }}</option>
          </select>
        </label>
        <label class="form-field">
          Inicio (Santiago)
          <input v-model="form.startTime" type="datetime-local" required :disabled="blocksStore.saving" />
        </label>
        <label class="form-field">
          Término (Santiago)
          <input v-model="form.endTime" type="datetime-local" required :disabled="blocksStore.saving" />
        </label>
        <label class="form-field full-width">
          Descripción interna (opcional)
          <textarea v-model="form.reason" rows="3" maxlength="500" :disabled="blocksStore.saving" />
          <small>Solo visible en administración. Máximo 500 caracteres.</small>
        </label>
        <p v-if="formError || blocksStore.actionError" role="alert" class="form-error full-width">
          {{ formError || blocksStore.actionError }}
        </p>
        <div class="form-actions full-width">
          <button type="button" class="app-button secondary" :disabled="blocksStore.saving" @click="showCreate = false">
            Cancelar
          </button>
          <button type="submit" class="app-button primary" :disabled="blocksStore.saving || resourcesStore.loading">
            {{ blocksStore.saving ? 'Guardando...' : 'Crear bloqueo' }}
          </button>
        </div>
      </form>
    </section>

    <p v-if="blocksStore.actionSuccess" class="state-card success" role="status">
      {{ blocksStore.actionSuccess }}
    </p>
    <p v-if="blocksStore.actionError && !showCreate" class="state-card error" role="alert">
      {{ blocksStore.actionError }}
    </p>

    <section class="app-card listing-panel" aria-label="Bloqueos registrados">
      <div class="panel-heading">
        <h2>Bloqueos registrados</h2>
        <p>Consulta un período de hasta 31 días. El último día seleccionado está incluido.</p>
      </div>
      <form class="filters" @submit.prevent="loadBlocks">
        <label class="form-field">Desde
          <input v-model="startDate" type="date" required />
        </label>
        <label class="form-field">Hasta
          <input v-model="endDate" type="date" required />
        </label>
        <label class="form-field">Instalación
          <select v-model="resourceFilter">
            <option value="ALL">Todas</option>
            <option v-for="resource in resourcesStore.resources" :key="resource.id" :value="String(resource.id)">
              {{ resource.name }}
            </option>
          </select>
        </label>
        <label class="form-field">Estado
          <select v-model="statusFilter">
            <option value="ACTIVE">Activos</option>
            <option value="INACTIVE">Inactivos</option>
            <option value="ALL">Todos</option>
          </select>
        </label>
        <button type="submit" class="app-button secondary" :disabled="blocksStore.loading">
          <RefreshCcw :size="16" /> Actualizar
        </button>
      </form>
      <p v-if="filterError || blocksStore.error" class="state-card error" role="alert">
        {{ filterError || blocksStore.error }}
      </p>
      <p v-if="resourcesStore.error" class="state-card warning" role="alert">
        No se pudieron cargar las instalaciones. Intenta actualizar la página.
      </p>
      <p v-if="blocksStore.loading" class="state-card" role="status">Cargando bloqueos...</p>
      <p v-else-if="!visibleBlocks.length && !blocksStore.error && !filterError" class="empty-state">
        No hay bloqueos que coincidan con los filtros del período.
      </p>
      <div v-else-if="!blocksStore.error && !filterError" class="blocks-list">
        <article v-for="block in visibleBlocks" :key="block.id" class="block-row">
          <div class="block-icon"><CalendarOff :size="20" aria-hidden="true" /></div>
          <div class="block-details">
            <div class="block-title">
              <strong>{{ block.resourceName || 'Instalación' }}</strong>
              <span :class="['status', block.isActive ? 'active' : 'inactive']">
                {{ block.isActive ? 'Activo' : 'Inactivo' }}
              </span>
            </div>
            <p>{{ typeLabel(block.blockType) }}<span v-if="block.reason"> · {{ block.reason }}</span></p>
            <small>{{ formatChileBlockTime(block.startTime) }} — {{ formatChileBlockTime(block.endTime) }}</small>
          </div>
          <div v-if="block.isActive" class="row-actions">
            <button
              v-if="confirmDeactivateId !== block.id"
              type="button" class="app-button secondary deactivate-trigger"
              :disabled="blocksStore.updatingId !== null"
              @click="confirmDeactivateId = block.id"
            >
              Desactivar
            </button>
            <div v-else class="confirm-actions">
              <div class="confirm-label"><ShieldAlert :size="16" aria-hidden="true" /> ¿Desactivar este bloqueo?</div>
              <button type="button" class="app-button danger confirm-deactivate" :disabled="blocksStore.updatingId !== null" @click="deactivate(block.id)">
                {{ blocksStore.updatingId === block.id ? 'Desactivando...' : 'Confirmar' }}
              </button>
              <button type="button" class="app-button secondary" :disabled="blocksStore.updatingId !== null" @click="confirmDeactivateId = null">Cancelar</button>
            </div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>

<style scoped>
.availability-blocks-view { display: flex; flex-direction: column; gap: var(--space-5); min-width: 0; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: var(--space-4); }
.page-header h1 { margin: 0; font-size: var(--text-page-title); font-weight: 800; color: var(--color-text); }
.page-header p { margin: var(--space-2) 0 0; color: var(--color-text-muted); }
.create-panel, .listing-panel { padding: var(--space-5); display: flex; flex-direction: column; gap: var(--space-4); }
.panel-heading h2 { margin: 0; font-size: var(--text-section-title); color: var(--color-text); }
.panel-heading p { margin: var(--space-2) 0 0; color: var(--color-text-muted); font-size: var(--text-help); }
.block-form { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-4); }
.form-field { min-width: 0; gap: var(--space-2); }
.form-field input, .form-field select, .form-field textarea { width: 100%; min-height: 42px; padding: var(--space-2) var(--space-3); background: var(--color-surface); }
.form-field textarea { resize: vertical; }
.form-field small { font-size: 12px; font-weight: 400; color: var(--color-text-muted); }
.full-width { grid-column: 1 / -1; }
.form-actions { display: flex; justify-content: flex-end; flex-wrap: wrap; gap: var(--space-3); }
.filters { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)) auto; align-items: end; gap: var(--space-3); }
.filters .app-button { white-space: nowrap; }
.blocks-list { display: flex; flex-direction: column; gap: var(--space-3); }
.block-row { display: flex; align-items: flex-start; gap: var(--space-3); padding: var(--space-4); border: 1px solid var(--color-border-soft); border-radius: var(--radius-lg); min-width: 0; }
.block-icon { display: flex; align-items: center; justify-content: center; padding: var(--space-3); background: var(--color-primary-soft); color: var(--color-primary-strong); border-radius: var(--radius-md); }
.block-details { flex: 1; min-width: 0; overflow-wrap: anywhere; }
.block-title { display: flex; align-items: center; flex-wrap: wrap; gap: var(--space-2); }
.block-details p { margin: var(--space-2) 0; color: var(--color-text-muted); font-size: var(--text-help); }
.block-details small { color: var(--color-text-muted); }
.status { border-radius: var(--radius-pill); padding: 4px var(--space-3); font-size: 12px; font-weight: 750; }
.status.active { color: var(--color-success); background: var(--color-success-soft); }
.status.inactive { color: var(--color-text-muted); background: var(--color-surface-muted); }
.row-actions { margin-left: auto; }
.confirm-actions { display: flex; align-items: flex-end; flex-direction: column; gap: var(--space-2); }
.confirm-label { display: flex; align-items: center; gap: var(--space-2); color: var(--color-error); font-size: var(--text-help); font-weight: 700; }
.empty-state { color: var(--color-text-muted); padding: var(--space-4); }
@media (max-width: 1000px) { .filters { grid-template-columns: repeat(2, minmax(0, 1fr)); } .filters button { width: 100%; } }
@media (max-width: 640px) { .page-header > button { width: 100%; } .block-form, .filters { grid-template-columns: 1fr; } .create-panel, .listing-panel { padding: var(--space-4); } .block-row { flex-wrap: wrap; } .row-actions { width: 100%; } .row-actions > button { width: 100%; } .confirm-actions { align-items: stretch; } }
</style>
