<script setup>
import { computed, onMounted, ref } from 'vue'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useResourcesStore } from '@/stores/resources'

const resourcesStore = useResourcesStore()
const search = ref('')
const typeFilter = ref('ALL')
const modeFilter = ref('ALL')
const statusFilter = ref('ALL')

onMounted(() => {
  resourcesStore.fetchResources()
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

    <div
      v-else-if="!resourcesStore.resources.length"
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

      </article>

    </section>

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

  color: #0f172a;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.state-card {
  background: white;

  border-radius: 18px;

  padding: 22px;

  border: 1px solid #e2e8f0;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.filters {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

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

  border: 1px solid #dbe2ea;
  border-radius: 12px;

  padding: 0 12px;

  box-sizing: border-box;
  outline: none;
}

.filters input:focus,
.filters select:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}

.resources-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(260px, 1fr));

  gap: 18px;
}

.resource-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.resource-card h2 {
  margin: 0;

  font-size: 18px;
  font-weight: 800;

  color: #0f172a;
}

.resource-card p {
  margin: 6px 0 0;

  color: #64748b;
}

.meta {
  display: flex;
  flex-wrap: wrap;

  gap: 8px;
}

.meta span {
  padding: 7px 10px;

  border-radius: 999px;

  background: #eff6ff;
  color: #1d4ed8;

  font-size: 12px;
  font-weight: 800;
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
