<script setup>
import { computed } from 'vue'

const props = defineProps({
  resources: { type: Array, default: () => [] },
  selectedId: { type: [Number, String], default: null },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
  isAdmin: { type: Boolean, default: false }
})

const emit = defineEmits(['select'])
const eligibleResources = computed(() => props.resources.filter(resource => {
  if (resource.isActive === false || ['inactive', 'maintenance'].includes(resource.status)) return false
  if (resource.reservationMode === 'INFORMATIVE') return false
  if (resource.reservationMode === 'ADMIN_ONLY' && !props.isAdmin) return false
  return true
}))

const selectedValue = computed(() => (
  props.selectedId === null || props.selectedId === undefined
    ? ''
    : String(props.selectedId)
))

const handleChange = (event) => {
  const resource = eligibleResources.value.find(
    item => String(item.id) === event.target.value
  )

  emit('select', resource || null)
}
</script>

<template>
  <div class="resource-field">
    <label for="reservation-resource">Instalación</label>

    <select
      id="reservation-resource"
      :value="selectedValue"
      :disabled="disabled || loading || !eligibleResources.length"
      :aria-invalid="Boolean(error)"
      :aria-describedby="error ? 'reservation-resource-error' : undefined"
      @change="handleChange"
    >
      <option value="" disabled>
        {{ loading ? 'Cargando instalaciones...' : 'Selecciona una instalación' }}
      </option>

      <option
        v-for="resource in eligibleResources"
        :key="resource.id"
        :value="String(resource.id)"
      >
        {{ resource.name }}
      </option>
    </select>

    <p
      v-if="error"
      id="reservation-resource-error"
      class="field-error"
      role="alert"
    >
      {{ error }}
    </p>

    <p v-else-if="!loading && !eligibleResources.length" class="empty-message">
      No hay instalaciones disponibles.
    </p>
  </div>
</template>

<style scoped>
.resource-field {
  display: flex;
  width: 100%;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.resource-field label {
  color: #334155;
  font-size: 14px;
  font-weight: 700;
}

.resource-field select {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  height: 50px;
  padding: 0 44px 0 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background-color: var(--color-surface);
  color: var(--color-text);
  font: inherit;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.resource-field select:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}

.resource-field select[aria-invalid='true'] {
  border-color: var(--color-error-strong);
  box-shadow: var(--shadow-danger-focus);
}

.resource-field select:disabled {
  cursor: not-allowed;
  background-color: var(--color-surface-muted);
  color: var(--color-text-soft);
}

.field-error,
.empty-message {
  margin: 0;
  font-size: 13px;
}

.field-error {
  color: var(--color-error);
  font-weight: 700;
}

.empty-message {
  color: var(--color-text-muted);
}
</style>
