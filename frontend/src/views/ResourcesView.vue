<script setup>
import { computed, onMounted, ref } from 'vue'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useResourcesStore } from '@/stores/resources'

const resourcesStore = useResourcesStore()
const authStore = useAuthStore()
const search = ref('')
const typeFilter = ref('ALL')
const modeFilter = ref('ALL')
const statusFilter = ref('ALL')
const editingImageId = ref(null)
const imageDraft = ref('')

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

const startImageEdit = (resource) => {
  editingImageId.value = resource.id
  imageDraft.value = resource.imageUrl || ''
  resourcesStore.actionError = null
  resourcesStore.actionSuccess = null
}

const cancelImageEdit = () => {
  editingImageId.value = null
  imageDraft.value = ''
}

const saveImageEdit = async (resource) => {
  try {
    await resourcesStore.updateImage(
      resource.id,
      imageDraft.value.trim()
    )

    cancelImageEdit()
  } catch {
    // El store deja el mensaje visible para el usuario.
  }
}
</script>

<template>
  <main class="resources-view">

    <header class="page-header">

      <h1>
        Instalaciones
      </h1>

      <p>
        Recursos deportivos disponibles desde la base de datos.
      </p>

    </header>

    <section class="filters">

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

          <span>
            {{ modeLabel(resource.reservationMode) }}
          </span>

          <span>
            {{ statusLabel(resource) }}
          </span>

          <span v-if="resource.capacity">
            Capacidad: {{ resource.capacity }}
          </span>

        </div>

        <div
          v-if="authStore.isAdmin"
          class="image-editor"
        >
          <template v-if="editingImageId === resource.id">
            <input
              v-model="imageDraft"
              type="url"
              placeholder="https://... o /images/..."
            />

            <div class="editor-actions">
              <button
                type="button"
                class="save"
                :disabled="resourcesStore.updatingImageId === resource.id"
                @click="saveImageEdit(resource)"
              >
                Guardar
              </button>

              <button
                type="button"
                class="cancel"
                :disabled="resourcesStore.updatingImageId === resource.id"
                @click="cancelImageEdit"
              >
                Cancelar
              </button>
            </div>
          </template>

          <button
            v-else
            type="button"
            class="edit-image"
            @click="startImageEdit(resource)"
          >
            Cambiar imagen
          </button>
        </div>

      </article>

      </section>
    </template>

  </main>
</template>

<style scoped>
.resources-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: var(--color-text);
}

.page-header p {
  margin-top: 8px;

  color: var(--color-text-muted);
}

.state-card {
  border-radius: var(--radius-lg);
}

.state-card.error {
  background: var(--color-error-soft);
  color: var(--color-error);
  border-color: var(--color-error-border);
}

.state-card.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border-color: var(--color-success-border);
}

.filters {
  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  padding: 18px;

  display: grid;
  grid-template-columns:
    minmax(220px, 1.4fr)
    repeat(3, minmax(150px, 1fr));

  gap: 14px;
}

.filters label {
  color: #334155;

  display: flex;
  flex-direction: column;

  gap: 7px;

  font-size: 13px;
  font-weight: 800;
}

.filters input,
.filters select {
  width: 100%;
  height: 42px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  padding: 0 12px;

  box-sizing: border-box;
  outline: none;
}

.resources-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(260px, 1fr));

  gap: 18px;
}

.resource-card {
  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow: var(--shadow-card);
}

.resource-image {
  height: 150px;
  margin: -20px -20px 0;
  overflow: hidden;
  background: var(--color-surface-soft);
}

.resource-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.resource-image.fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-primary);
  font-size: 30px;
  font-weight: 900;
  background: var(--color-primary-soft);
}

.resource-card h2 {
  margin: 0;

  font-size: 18px;
  font-weight: 800;

  color: var(--color-text);
}

.resource-card p {
  margin: 6px 0 0;

  color: var(--color-text-muted);
}

.meta {
  display: flex;
  flex-wrap: wrap;

  gap: 8px;
}

.meta span {
  padding: 7px 10px;

  border-radius: 999px;

  background: var(--color-primary-soft);
  color: var(--color-primary-strong);

  font-size: 12px;
  font-weight: 800;
}

.image-editor {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.image-editor input {
  height: 42px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 0 12px;
  font: inherit;
}

.editor-actions {
  display: flex;
  gap: 8px;
}

.edit-image,
.editor-actions button {
  min-height: 38px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 0 12px;
  font-weight: 800;
  cursor: pointer;
}

.edit-image,
.editor-actions .save {
  background: var(--color-primary);
  color: var(--color-primary-contrast);
  border-color: var(--color-primary);
}

.editor-actions .cancel {
  background: var(--color-surface);
  color: var(--color-text);
}

@media (max-width: 900px) {
  .filters {
    grid-template-columns: 1fr 1fr;
  }

  .search-field {
    grid-column: 1 / -1;
  }
}

@media (max-width: 600px) {
  .filters {
    grid-template-columns: 1fr;
  }

  .search-field {
    grid-column: auto;
  }
}
</style>
