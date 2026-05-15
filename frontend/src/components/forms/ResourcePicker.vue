<script setup>
const props = defineProps({
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

    <!-- DEBUG -->
    <pre class="debug">
{{ resources }}
    </pre>

    <!-- EMPTY -->
    <div
      v-if="
        !resources ||
        resources.length === 0
      "
      class="empty"
    >
      No hay instalaciones.
    </div>

    <!-- LIST -->
    <div
      v-else
      class="resources"
    >

      <div
        v-for="resource in resources"
        :key="resource.id"

        class="resource"

        :class="{
          selected:
            selectedId === resource.id
        }"

        @click="
          handleSelect(resource)
        "
      >

        <!-- LEFT -->
        <div class="left">

          <!-- STATUS -->
          <div
            class="status"
            :class="
              resource.status ||
              'available'
            "
          />

          <!-- INFO -->
          <div class="info">

            <div class="name">
              {{ resource.name }}
            </div>

            <div class="type">
              {{ resource.type }}
            </div>

          </div>

        </div>

        <!-- RIGHT -->
        <div
          v-if="
            selectedId === resource.id
          "
          class="badge"
        >
          Seleccionado
        </div>

      </div>

    </div>

  </div>
</template>

<style scoped>
/* ROOT */
.picker {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 16px;
}

/* HEADER */
.header {
  display: flex;
  flex-direction: column;

  gap: 4px;
}

.header h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 700;

  color: #111827;
}

.header p {
  margin: 0;

  font-size: 14px;

  color: #6b7280;
}

/* DEBUG */
.debug {
  padding: 12px;

  background: #f3f4f6;

  border-radius: 12px;

  font-size: 12px;

  overflow-x: auto;
}

/* EMPTY */
.empty {
  padding: 18px;

  border-radius: 16px;

  border: 1px dashed #d1d5db;

  background: #f9fafb;

  color: #6b7280;

  font-size: 14px;
}

/* LIST */
.resources {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 12px;
}

/* CARD */
.resource {
  width: 100%;

  min-height: 72px;

  background: white;

  border: 2px solid #e5e7eb;

  border-radius: 18px;

  padding: 16px;

  box-sizing: border-box;

  display: flex !important;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  cursor: pointer;

  transition: 0.2s;

  user-select: none;
}

/* FORCE */
.resource * {
  box-sizing: border-box;
}

/* HOVER */
.resource:hover {
  border-color: #93c5fd;

  background: #eff6ff;
}

/* SELECTED */
.selected {
  border-color: #2563eb !important;

  background: #dbeafe !important;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.1);
}

/* LEFT */
.left {
  display: flex !important;
  align-items: center;

  gap: 14px;

  min-width: 0;
}

/* STATUS */
.status {
  width: 14px;
  height: 14px;

  border-radius: 999px;

  flex-shrink: 0;
}

/* STATUS COLORS */
.available {
  background: #22c55e;
}

.busy {
  background: #ef4444;
}

.maintenance {
  background: #f59e0b;
}

/* INFO */
.info {
  display: flex;
  flex-direction: column;

  min-width: 0;
}

.name {
  font-size: 15px;
  font-weight: 700;

  color: #111827;
}

.type {
  margin-top: 2px;

  font-size: 13px;

  color: #6b7280;
}

/* BADGE */
.badge {
  flex-shrink: 0;

  padding: 6px 12px;

  border-radius: 999px;

  background: #2563eb;

  color: white;

  font-size: 12px;
  font-weight: 700;
}

/* MOBILE */
@media (max-width: 768px) {
  .resource {
    padding: 14px;
  }

  .name {
    font-size: 14px;
  }
}
</style>