<script setup>
import { computed, onMounted, ref } from 'vue'

import { useResourcesStore } from '@/stores/resources'
import { useVenuesStore } from '@/stores/venues'

const emit = defineEmits(['created', 'cancel'])

const resourcesStore = useResourcesStore()
const venuesStore = useVenuesStore()

const venues = computed(() => venuesStore.venues)

const venueId = ref('')
const name = ref('')
const resourceType = ref('')
const reservationMode = ref('RESERVABLE')
const capacity = ref('')
const imageUrl = ref('')
const isActive = ref(true)
const validationError = ref('')

onMounted(() => {
  if (!venuesStore.loading && !venuesStore.venues.length) {
    venuesStore.fetchVenues().catch(() => {})
  }
})

const createResource = async () => {
  validationError.value = ''

  const selectedVenueId = Number(venueId.value)
  const normalizedName = name.value.trim()
  const normalizedType = resourceType.value.trim()
  const normalizedImage = imageUrl.value.trim()

  if (!Number.isSafeInteger(selectedVenueId) ||
      selectedVenueId <= 0) {
    validationError.value = 'Selecciona una sede.'
    return
  }

  if (!normalizedName || [...normalizedName].length > 120) {
    validationError.value =
      'El nombre debe tener entre 1 y 120 caracteres.'
    return
  }

  if (!normalizedType || [...normalizedType].length > 80) {
    validationError.value =
      'El tipo debe tener entre 1 y 80 caracteres.'
    return
  }

  let parsedCapacity = null

  if (capacity.value !== '') {
    parsedCapacity = Number(capacity.value)

    if (!Number.isSafeInteger(parsedCapacity) ||
        parsedCapacity <= 0) {
      validationError.value =
        'La capacidad debe ser un entero positivo.'
      return
    }
  }

  if (normalizedImage.length > 500 ||
      (normalizedImage &&
       !normalizedImage.startsWith('https://') &&
       !normalizedImage.startsWith('http://') &&
       !normalizedImage.startsWith('/'))) {
    validationError.value = 'La URL de imagen no es válida.'
    return
  }

  try {
    await resourcesStore.createResource({
      venueId: selectedVenueId,
      name: normalizedName,
      type: normalizedType,
      reservationMode: reservationMode.value,
      capacity: parsedCapacity,
      imageUrl: normalizedImage,
      isActive: isActive.value
    })

    name.value = ''
    resourceType.value = ''
    capacity.value = ''
    imageUrl.value = ''

    // Actualizar los contadores de recursos en las sedes.
    venuesStore.fetchVenues().catch(() => {})
    emit('created')
  } catch {
    // El store mantiene visible el error del backend.
  }
}
</script>

<template>
  <section class="resource-create-panel">

    <form class="resource-form" @submit.prevent="createResource">
      <h2>Nueva instalación</h2>

      <p>
        Registra el recurso en su sede correspondiente.
        No se incorpora automáticamente a las políticas
        de reserva existentes.
      </p>

      <p v-if="validationError" role="alert" class="form-error">
        {{ validationError }}
      </p>

      <p
        v-if="resourcesStore.actionError"
        role="alert"
        class="form-error"
      >
        {{ resourcesStore.actionError }}
      </p>

      <p
        v-if="resourcesStore.actionSuccess"
        role="status"
        class="state-card success"
      >
        {{ resourcesStore.actionSuccess }}
      </p>

      <p v-if="venuesStore.error" role="alert" class="form-error">
        {{ venuesStore.error }}
      </p>

      <div class="resource-form-grid">
        <label>
          Sede
          <select v-model="venueId" required>
            <option value="">Seleccionar sede</option>
            <option
              v-for="venue in venues"
              :key="venue.id"
              :value="venue.id"
            >
              {{ venue.name }}
              {{ venue.isActive ? '' : '(inactiva)' }}
            </option>
          </select>
        </label>

        <label>
          Nombre
          <input
            v-model="name"
            type="text"
            maxlength="120"
            required
          />
        </label>

        <label>
          Tipo
          <input
            v-model="resourceType"
            type="text"
            maxlength="80"
            placeholder="Cancha, Sala, Gimnasio..."
            required
          />
        </label>

        <label>
          Modo de reserva
          <select v-model="reservationMode">
            <option value="RESERVABLE">Reservable</option>
            <option value="OPEN_USE">Uso libre</option>
            <option value="INFORMATIVE">Informativo</option>
            <option value="ADMIN_ONLY">Solo administrador</option>
          </select>
        </label>

        <label>
          Capacidad
          <input
            v-model="capacity"
            type="number"
            min="1"
            step="1"
            placeholder="Opcional"
          />
        </label>

      </div>

      <details class="advanced-options">
        <summary>Opciones adicionales (imagen y estado)</summary>
        <label class="advanced-field">
          URL de imagen (opcional)
          <input
            v-model="imageUrl"
            type="text"
            maxlength="500"
            placeholder="https://... o /images/..."
          />
        </label>
        <label class="active-label">
          <input v-model="isActive" type="checkbox" />
          Crear como instalación activa
        </label>
        <p class="helper">
          Una sede inactiva no puede recibir recursos activos.
        </p>
      </details>

      <div class="form-actions">
        <button
          type="submit"
          class="app-button primary"
          :disabled="resourcesStore.creatingResource || venuesStore.loading || !venues.length"
        >
          {{ resourcesStore.creatingResource ? 'Creando...' : 'Crear instalación' }}
        </button>
        <button
          type="button"
          class="app-button secondary"
          :disabled="resourcesStore.creatingResource"
          @click="emit('cancel')"
        >Cancelar</button>
      </div>
    </form>
  </section>
</template>

<style scoped>
.resource-create-panel {
  min-width: 0;
  padding: var(--space-5);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.resource-form {
  display: grid;
  gap: var(--space-4);
}

.resource-form h2 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-section-title);
}

.resource-form > p:not(.form-error):not(.state-card) {
  margin: 0;
  color: var(--color-text-muted);
  line-height: 1.5;
}

.resource-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.resource-form-grid label,
.advanced-field {
  display: grid;
  min-width: 0;
  gap: var(--space-2);
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.resource-form-grid input,
.resource-form-grid select,
.advanced-field input {
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

.advanced-options {
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface-muted);
}

.advanced-options summary {
  cursor: pointer;
  color: var(--color-primary-strong);
  font-weight: 700;
}

.advanced-options summary:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 3px;
}

.advanced-field { margin-top: var(--space-3); }

.active-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  font-size: var(--text-help);
  font-weight: 700;
}

.active-label input {
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.helper {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--text-help);
  line-height: 1.5;
}

.form-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

@media (max-width: 600px) {
  .resource-form-grid { grid-template-columns: 1fr; }
}

@media (max-width: 480px) {
  .resource-create-panel { padding: var(--space-4); }
  .form-actions .app-button { flex: 1 1 140px; }
}
</style>
