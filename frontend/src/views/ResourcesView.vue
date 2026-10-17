<script setup>
import { computed, onMounted, ref } from 'vue'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useResourcesStore } from '@/stores/resources'
import { useVenuesStore } from '@/stores/venues'
import VenuesAdminPanel from '@/components/resources/VenuesAdminPanel.vue'
import ResourceCreatePanel from '@/components/resources/ResourceCreatePanel.vue'

const resourcesStore = useResourcesStore()
const venuesStore = useVenuesStore()
const authStore = useAuthStore()
const adminTab = ref('resources')
const showCreateResource = ref(false)
const search = ref('')
const typeFilter = ref('ALL')
const modeFilter = ref('ALL')
const statusFilter = ref('ALL')
const venueFilter = ref('ALL')
const editingResourceId = ref(null)
const nameDraft = ref('')
const typeDraft = ref('')
const modeDraft = ref('RESERVABLE')
const capacityDraft = ref('')
const activeDraft = ref(true)

const imageDraft = ref('')
// Ultimos datos generales confirmados durante esta edicion.
const savedOperationalDraft = ref(null)

onMounted(() => {
  resourcesStore.fetchResources()
  authStore.loadAuthUser()
})

const statusLabel = (resource) => {
  if (!resource.isActive) {
    return 'Inactivo'
  }

  return 'Disponible'
}

const modeLabel = (mode) => {
  switch (mode) {
    case 'ADMIN_ONLY':
      return 'Solo admin'

    case 'INFORMATIVE':
      return 'Informativo'

    case 'OPEN_USE':
      return 'Uso libre'

    default:
      return 'Reservable'
  }
}

const typeOptions = computed(() => {
  return [...new Set(
    resourcesStore.resources
      .map((resource) => resource.type)
      .filter(Boolean)
  )].sort((a, b) => a.localeCompare(b))
})

const filteredResources = computed(() => {
  const query = search.value.trim().toLowerCase()

  return resourcesStore.resources.filter((resource) => {
    if (
      query &&
      !`${resource.name} ${resource.type}`.toLowerCase().includes(query)
    ) {
      return false
    }

    if (
      venueFilter.value !== 'ALL' &&
      String(resource.venueId ?? '') !== String(venueFilter.value)
    ) {
      return false
    }

    if (
      typeFilter.value !== 'ALL' &&
      resource.type !== typeFilter.value
    ) {
      return false
    }

    if (
      modeFilter.value !== 'ALL' &&
      resource.reservationMode !== modeFilter.value
    ) {
      return false
    }

    if (statusFilter.value === 'ACTIVE' && !resource.isActive) {
      return false
    }

    if (statusFilter.value === 'INACTIVE' && resource.isActive) {
      return false
    }

    return true
  })
})

const startOperationalEdit = (resource) => {
  savedOperationalDraft.value = null
  editingResourceId.value = resource.id
  nameDraft.value = resource.name
  typeDraft.value = resource.type
  modeDraft.value = resource.reservationMode

  capacityDraft.value =
    resource.capacity === null ||
    resource.capacity === undefined
      ? ''
      : String(resource.capacity)

  activeDraft.value = resource.isActive
  imageDraft.value = resource.imageUrl || ''

  resourcesStore.actionError = null
  resourcesStore.actionSuccess = null
}

const cancelOperationalEdit = () => {
  savedOperationalDraft.value = null
  editingResourceId.value = null
  nameDraft.value = ''
  typeDraft.value = ''
  modeDraft.value = 'RESERVABLE'
  capacityDraft.value = ''
  activeDraft.value = true
  imageDraft.value = ''
}

const saveOperationalEdit = async (resource) => {
  const name = nameDraft.value.trim()
  const resourceType = typeDraft.value.trim()

  if (!name || [...name].length > 120) {
    resourcesStore.actionError =
      'El nombre es obligatorio y admite hasta 120 caracteres.'
    return
  }

  if (!resourceType || [...resourceType].length > 80) {
    resourcesStore.actionError =
      'El tipo es obligatorio y admite hasta 80 caracteres.'
    return
  }

  const rawCapacity = String(capacityDraft.value).trim()
  let capacity = null

  if (rawCapacity !== '') {
    const parsed = Number(rawCapacity)
    if (!Number.isSafeInteger(parsed) || parsed <= 0) {
      resourcesStore.actionError =
        'La capacidad debe ser un entero mayor que cero.'
      return
    }
    capacity = parsed
  }

  const imageUrl = imageDraft.value.trim()
  const imageChanged = imageUrl !== (resource.imageUrl || '')

  if (imageChanged && (imageUrl.length > 500 ||
    (imageUrl && !imageUrl.startsWith('https://') &&
      !imageUrl.startsWith('http://') && !imageUrl.startsWith('/')))) {
    resourcesStore.actionError = 'La URL de imagen no es válida.'
    return
  }

  const operational = {
    name,
    type: resourceType,
    reservationMode: modeDraft.value,
    capacity,
    isActive: activeDraft.value
  }
  // Comparar con los datos ya confirmados, aunque el store
  // todavia no haya actualizado la interfaz.
  const baseline = savedOperationalDraft.value?.id === resource.id
    ? savedOperationalDraft.value
    : resource

  const operationalChanged =
    name !== baseline.name ||
    resourceType !== baseline.type ||
    modeDraft.value !== baseline.reservationMode ||
    capacity !== (baseline.capacity ?? null) ||
    activeDraft.value !== baseline.isActive

  let operationalSaved = false
  try {
    // Dos endpoints existentes: no es una transacción atómica.
    if (operationalChanged) {
      await resourcesStore.updateOperational(resource.id, operational)
      savedOperationalDraft.value = { id: resource.id, ...operational }
      operationalSaved = true
    }

    if (imageChanged) {
      await resourcesStore.updateImage(resource.id, imageUrl)
    }

    cancelOperationalEdit()
    if (operationalChanged || imageChanged) {
      venuesStore.fetchVenues().catch(() => {})
    }
  } catch {
    if (operationalSaved || savedOperationalDraft.value?.id === resource.id) {
      resourcesStore.actionError =
        'Los datos generales se guardaron, pero la imagen no. ' +
        'Corrige la imagen e intenta guardar nuevamente.'
    }
    // Se mantiene abierto el editor. El siguiente intento guarda solo
    // las diferencias con el recurso actualizado del store.
  }
}

</script>

<template>
  <main class="resources-view">

    <header class="page-header">
      <div>
        <h1>Instalaciones deportivas</h1>
        <p>Consulta y administra los espacios del polideportivo.</p>
      </div>

      <button
        v-if="authStore.isAdmin && adminTab === 'resources'"
        type="button"
        class="app-button primary"
        :aria-expanded="showCreateResource"
        @click="showCreateResource = !showCreateResource"
      >
        {{ showCreateResource ? 'Cerrar formulario' : '+ Nueva instalación' }}
      </button>
    </header>

    <nav
      v-if="authStore.isAdmin"
      class="admin-tabs"
      aria-label="Administración de instalaciones"
    >
      <button
        type="button"
        :aria-pressed="adminTab === 'resources'"
        :class="{ active: adminTab === 'resources' }"
        @click="adminTab = 'resources'"
      >
        Instalaciones
      </button>
      <button
        type="button"
        :aria-pressed="adminTab === 'venues'"
        :class="{ active: adminTab === 'venues' }"
        @click="adminTab = 'venues'; showCreateResource = false"
      >
        Sedes
      </button>
    </nav>

    <!-- Se mantiene montado para conservar la carga del store de sedes. -->
    <section
      v-if="authStore.isAdmin"
      v-show="adminTab === 'venues'"
      class="venues-section"
      aria-label="Administración de sedes"
    >
      <VenuesAdminPanel />
    </section>

    <template v-if="!authStore.isAdmin || adminTab === 'resources'">
      <section v-if="authStore.isAdmin && showCreateResource" class="create-section">
        <ResourceCreatePanel @created="showCreateResource = false" @cancel="showCreateResource = false" />
      </section>

    <section class="filters">

      <label v-if="authStore.isAdmin">
        Sede
        <select
          v-model="venueFilter"
          :disabled="venuesStore.loading || !!venuesStore.error"
        >
          <option value="ALL">Todas las sedes</option>
          <option
            v-for="venue in venuesStore.venues"
            :key="venue.id"
            :value="venue.id"
          >
            {{ venue.name }}
          </option>
        </select>
      </label>

      <label class="search-field">
        Buscar
        <input
          v-model="search"
          type="search"
          placeholder="Nombre o tipo"
        />
      </label>

      <label>
        Tipo
        <select v-model="typeFilter">
          <option value="ALL">
            Todos
          </option>
          <option
            v-for="type in typeOptions"
            :key="type"
            :value="type"
          >
            {{ type }}
          </option>
        </select>
      </label>

      <label>
        Modo
        <select v-model="modeFilter">
          <option value="ALL">
            Todos
          </option>
          <option value="RESERVABLE">
            Reservable
          </option>
          <option value="OPEN_USE">
            Uso libre
          </option>
          <option value="ADMIN_ONLY">
            Solo admin
          </option>
          <option value="INFORMATIVE">
            Informativo
          </option>
        </select>
      </label>

      <label>
        Estado
        <select v-model="statusFilter">
          <option value="ALL">
            Todos
          </option>
          <option value="ACTIVE">
            Activos
          </option>
          <option value="INACTIVE">
            Inactivos
          </option>
        </select>
      </label>

    </section>

    <div
      v-if="resourcesStore.loading"
      aria-label="Cargando instalaciones"
    >
      <SkeletonLoader
        variant="resources"
        :items="6"
      />
    </div>

    <div
      v-else-if="resourcesStore.error"
      class="state-card error"
    >
      {{ resourcesStore.error }}
    </div>

    <template v-else>
      <div
        v-if="resourcesStore.actionError"
        class="state-card error"
      >
        {{ resourcesStore.actionError }}
      </div>

      <div
        v-if="resourcesStore.actionSuccess"
        class="state-card success"
      >
        {{ resourcesStore.actionSuccess }}
      </div>

      <div
        v-if="!resourcesStore.resources.length"
        class="state-card"
      >
        No hay instalaciones registradas.
      </div>

      <div
        v-else-if="!filteredResources.length"
        class="state-card"
      >
        No hay instalaciones que coincidan con los filtros.
      </div>

      <section
      v-else
      class="resources-grid"
    >

      <article
        v-for="resource in filteredResources"
        :key="resource.id"
        class="resource-card"
        :class="{ 'is-editing': editingResourceId === resource.id }"
      >

        <div
          v-if="resource.imageUrl"
          class="resource-image"
        >
          <img
            :src="resource.imageUrl"
            :alt="resource.name"
          />
        </div>

        <div
          v-else
          class="resource-image fallback"
        >
          {{ resource.name.slice(0, 1) }}
        </div>

        <div>

          <h2>
            {{ resource.name }}
          </h2>

          <p>
            {{ resource.type }}
          </p>

        </div>

        <div class="meta">

          <span class="mode-badge">
            {{ modeLabel(resource.reservationMode) }}
          </span>

          <span
            class="status-badge"
            :class="resource.isActive ? 'is-active' : 'is-inactive'"
          >
            {{ statusLabel(resource) }}
          </span>

          <span v-if="resource.capacity" class="capacity-badge">
            Capacidad: {{ resource.capacity }}
          </span>

        </div>

        <div v-if="authStore.isAdmin" class="resource-admin">
          <form
            v-if="editingResourceId === resource.id"
            class="resource-editor"
            @submit.prevent="saveOperationalEdit(resource)"
          >
            <h3>Editar instalación</h3>
            <p class="config-help">
              {{ resource.name }} · {{ venuesStore.venues.find(v => v.id === resource.venueId)?.name || 'Sede asignada' }}
            </p>
            <p v-if="resourcesStore.actionError" class="form-error" role="alert">
              {{ resourcesStore.actionError }}
            </p>
            <div class="config-grid">
              <label>
                Nombre
                <input v-model="nameDraft" type="text" maxlength="120" required />
              </label>
              <label>
                Tipo
                <input v-model="typeDraft" type="text" maxlength="80" required />
              </label>
              <label>
                Modo de reserva
                <select v-model="modeDraft">
                  <option value="RESERVABLE">Reservable</option>
                  <option value="OPEN_USE">Uso libre</option>
                  <option value="ADMIN_ONLY">Solo admin</option>
                  <option value="INFORMATIVE">Informativo</option>
                </select>
              </label>
              <label>
                Capacidad
                <input v-model="capacityDraft" type="number" min="1" step="1" placeholder="Sin capacidad" />
              </label>
            </div>
            <label class="active-control">
              <input v-model="activeDraft" type="checkbox" />
              Instalación activa
            </label>
            <p class="config-help">
              Los recursos con confirmación grupal deben mantener modo
              y capacidad compatibles con la política vigente.
            </p>
            <details class="advanced-fields">
              <summary>Imagen de la instalación (opcional)</summary>
              <label class="image-field">
                URL de imagen
                <input v-model="imageDraft" type="text" maxlength="500" placeholder="https://... o /images/..." />
              </label>
              <p class="config-help">
                Los datos y la imagen usan operaciones independientes.
                Si una falla, se informará cuál debes revisar.
              </p>
            </details>
            <div class="editor-actions">
              <button
                type="submit"
                class="app-button primary"
                :disabled="resourcesStore.updatingResourceId === resource.id || resourcesStore.updatingImageId === resource.id"
              >
                {{ resourcesStore.updatingResourceId === resource.id || resourcesStore.updatingImageId === resource.id ? 'Guardando...' : 'Guardar cambios' }}
              </button>
              <button
                type="button"
                class="app-button secondary"
                :disabled="resourcesStore.updatingResourceId === resource.id || resourcesStore.updatingImageId === resource.id"
                @click="cancelOperationalEdit"
              >Cancelar</button>
            </div>
          </form>
          <button
            v-else
            type="button"
            class="app-button secondary edit-resource"
            @click="startOperationalEdit(resource)"
          >Editar instalación</button>
        </div>

      </article>

      </section>
    </template>
    </template>

  </main>
</template>

<style scoped>
.resources-view {
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.page-header > div { min-width: 0; }

.page-header h1 {
  margin: 0;
  font-size: var(--text-page-title);
  font-weight: 800;
  line-height: 1.15;
  color: var(--color-text);
}

.page-header p {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  line-height: 1.45;
}

.admin-tabs {
  display: inline-flex;
  align-items: stretch;
  gap: var(--space-1);
  max-width: 100%;
  width: fit-content;
  padding: var(--space-1);
  background: var(--color-surface-soft);
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-lg);
}

.admin-tabs button {
  min-height: 42px;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-muted);
  padding: 0 var(--space-5);
  cursor: pointer;
  font: inherit;
  font-weight: 700;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.admin-tabs button:hover:not(.active) {
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
}

.admin-tabs button.active {
  background: var(--color-surface);
  color: var(--color-primary-strong);
  box-shadow: var(--shadow-card);
}

.admin-tabs button:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.venues-section,
.create-section { min-width: 0; }

.filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.filters label {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: var(--space-2);
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.filters .search-field { grid-column: span 2; }

.filters input,
.filters select {
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

.filters input:focus-visible,
.filters select:focus-visible {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-focus);
  outline: none;
}

.resources-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 255px), 1fr));
  align-items: stretch;
  gap: var(--space-4);
  min-width: 0;
}

.resource-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.resource-image {
  height: 150px;
  margin: calc(-1 * var(--space-5)) calc(-1 * var(--space-5)) 0;
  overflow: hidden;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  background: var(--color-surface-soft);
}

.resource-image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.resource-image.fallback {
  display: flex;
  justify-content: center;
  align-items: center;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-size: 30px;
  font-weight: 800;
}

.resource-card h2 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-section-title);
  font-weight: 800;
  overflow-wrap: anywhere;
}

.resource-card p {
  margin: var(--space-1) 0 0;
  color: var(--color-text-muted);
  line-height: 1.45;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.meta span {
  display: inline-flex;
  align-items: center;
  min-height: 27px;
  padding: 4px var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--color-primary-soft);
  color: var(--color-primary-strong);
  font-size: 12px;
  font-weight: 750;
}

.meta .status-badge.is-active {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.meta .status-badge.is-inactive {
  background: var(--color-surface-soft);
  color: var(--color-text-muted);
}

.meta .capacity-badge {
  background: var(--color-surface-muted);
  color: var(--color-text);
  border: 1px solid var(--color-border-soft);
}

.resource-admin {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin-top: auto;
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border-soft);
}

.resource-admin > .edit-resource { align-self: flex-start; }

.resource-card.is-editing {
  grid-column: 1 / -1;
  border-color: var(--color-primary);
}

.resource-editor {
  display: grid;
  min-width: 0;
  gap: var(--space-4);
}

.resource-editor h3 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-section-title);
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.config-grid label,
.image-field {
  display: grid;
  min-width: 0;
  gap: var(--space-2);
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.config-grid input,
.config-grid select,
.image-field input {
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

.active-control {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.active-control input {
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.config-help {
  margin: 0;
  color: var(--color-text-muted);
  font-size: var(--text-help);
  line-height: 1.5;
}

.advanced-fields {
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface-muted);
}

.advanced-fields summary {
  cursor: pointer;
  color: var(--color-primary-strong);
  font-weight: 700;
}

.advanced-fields summary:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 3px;
}

.advanced-fields .image-field { margin-top: var(--space-3); }
.advanced-fields .config-help { margin-top: var(--space-2); }

.editor-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

@media (min-width: 1101px) {
  .filters { grid-template-columns: 1fr 1.7fr repeat(3, 1fr); }
  .filters .search-field { grid-column: auto; }
}

@media (max-width: 720px) {
  .filters { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .filters .search-field { grid-column: 1 / -1; }
}

@media (max-width: 600px) {
  .page-header > .app-button { width: 100%; }
  .admin-tabs { width: 100%; }
  .admin-tabs button { flex: 1; padding: 0 var(--space-3); }
  .config-grid { grid-template-columns: 1fr; }
  .editor-actions .app-button { flex: 1 1 140px; }
}

@media (max-width: 480px) {
  .filters { grid-template-columns: 1fr; }
  .filters .search-field { grid-column: auto; }
  .resource-card { padding: var(--space-4); }
  .resource-image {
    margin: calc(-1 * var(--space-4)) calc(-1 * var(--space-4)) 0;
  }
}
</style>
