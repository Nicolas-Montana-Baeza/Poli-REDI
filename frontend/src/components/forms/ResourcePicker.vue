<script setup>
defineProps({
  resources: {
    type: Array,
    default: () => []
  },

  selectedId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits([
  'select'
])

const handleSelect = (resource) => {
  emit('select', resource)
}
</script>

<template>
  <div class="picker">

    <!-- HEADER -->
    <div class="header">

      <h3>
        Instalación
      </h3>

      <p>
        Selecciona un recurso.
      </p>

    </div>

    <!-- EMPTY -->
    <div
      v-if="!resources || resources.length === 0"
      class="empty"
    >
      No hay instalaciones disponibles.
    </div>

    <!-- LIST -->
    <div
      v-else
      class="resources"
    >

      <div
        v-for="resource in resources"
        :key="resource.id"

        class="resource-card"

        :class="{
          selected:
            selectedId === resource.id
        }"

        @click="handleSelect(resource)"
      >

        <!-- LEFT -->
        <div class="left">

          <div
            class="status-dot"
            :class="resource.status"
          />

          <div class="info">

            <strong>
              {{ resource.name }}
            </strong>

            <span>
              {{ resource.type }}
            </span>

          </div>

        </div>

        <!-- RIGHT -->
        <div
          v-if="selectedId === resource.id"
          class="badge"
        >
          Seleccionado
        </div>

      </div>

    </div>

  </div>
</template>

<style scoped>
.picker {
  display: flex;
  flex-direction: column;

  gap: 16px;
}

/* HEADER */
.header h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 800;

  color: #0f172a;
}

.header p {
  margin-top: 4px;

  font-size: 14px;

  color: #64748b;
}

/* EMPTY */
.empty {
  padding: 18px;

  border-radius: 18px;

  background: #f8fafc;

  border: 1px dashed #cbd5e1;

  color: #64748b;

  font-size: 14px;
}

/* LIST */
.resources {
  display: flex;
  flex-direction: column;

  gap: 12px;
}

/* CARD */
.resource-card {
  width: 100%;

  min-height: 72px;

  padding: 16px;

  border-radius: 18px;

  border: 1px solid #e2e8f0;

  background: white;

  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 14px;

  cursor: pointer;

  transition: 0.2s;

  box-sizing: border-box;
}

.resource-card:hover {
  background: #eff6ff;

  border-color: #bfdbfe;
}

/* SELECTED */
.resource-card.selected {
  background: #dbeafe;

  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}

/* LEFT */
.left {
  display: flex;
  align-items: center;

  gap: 14px;

  min-width: 0;
}

/* STATUS */
.status-dot {
  width: 12px;
  height: 12px;

  border-radius: 999px;

  flex-shrink: 0;
}

.status-dot.available {
  background: #22c55e;
}

.status-dot.busy {
  background: #ef4444;
}

.status-dot.maintenance {
  background: #f59e0b;
}

/* INFO */
.info {
  display: flex;
  flex-direction: column;

  min-width: 0;
}

.info strong {
  font-size: 15px;
  font-weight: 800;

  color: #0f172a;
}

.info span {
  margin-top: 2px;

  font-size: 13px;

  color: #64748b;
}

/* BADGE */
.badge {
  padding: 6px 12px;

  border-radius: 999px;

  background: #2563eb;

  color: white;

  font-size: 12px;
  font-weight: 700;

  white-space: nowrap;
}

/* MOBILE */
@media (max-width: 768px) {
  .resource-card {
    padding: 14px;
  }

  .badge {
    display: none;
  }
}
</style>