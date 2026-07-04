<script setup>
import { onMounted } from 'vue'

import { useResourcesStore } from '@/stores/resources'

const resourcesStore = useResourcesStore()

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

    <div
      v-if="resourcesStore.loading"
      class="state-card"
    >
      Cargando instalaciones...
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

    <section
      v-else
      class="resources-grid"
    >

      <article
        v-for="resource in resourcesStore.resources"
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
</style>
