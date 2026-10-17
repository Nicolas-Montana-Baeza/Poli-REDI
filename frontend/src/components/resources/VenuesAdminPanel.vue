<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useVenuesStore } from '@/stores/venues'

const venuesStore = useVenuesStore()

const showCreateVenue = ref(false)
const newName = ref('')
const newCountry = ref('Chile')
const editingVenueId = ref(null)
const nameDraft = ref('')
const activeDraft = ref(true)
const validationError = ref('')
const editCountry = ref('Chile')

const emptyLocation = () => ({
  addressLine: '',
  commune: '',
  city: '',
  region: '',
  latitude: '',
  longitude: ''
})

const createLocation = reactive(emptyLocation())
const editLocation = reactive(emptyLocation())

const locationFields = [
  { key: 'addressLine', label: 'Dirección', maxLength: 200 },
  { key: 'commune', label: 'Comuna', maxLength: 100 },
  { key: 'city', label: 'Ciudad', maxLength: 100 },
  { key: 'region', label: 'Región', maxLength: 100 }
]

const buildLocationPayload = (form) => {
  const payload = {}

  for (const field of locationFields) {
    payload[field.key] = String(form[field.key] ?? '').trim()
  }

  const latitudeText = String(form.latitude ?? '').trim().replace(',', '.')
  const longitudeText = String(form.longitude ?? '').trim().replace(',', '.')

  if (Boolean(latitudeText) !== Boolean(longitudeText)) {
    return {
      error: 'Debes ingresar ambas coordenadas o dejar ambas vacías.'
    }
  }

  if (!latitudeText) {
    return {
      payload: {
        ...payload,
        latitude: null,
        longitude: null
      }
    }
  }

  const latitude = Number(latitudeText)
  const longitude = Number(longitudeText)

  if (
    !Number.isFinite(latitude) ||
    !Number.isFinite(longitude) ||
    latitude < -90 ||
    latitude > 90 ||
    longitude < -180 ||
    longitude > 180
  ) {
    return {
      error: 'Las coordenadas ingresadas no son válidas.'
    }
  }

  return {
    payload: { ...payload, latitude, longitude }
  }
}

const loadVenues = async () => {
  try {
    await venuesStore.fetchVenues()
  } catch {
    // El store conserva el error visible.
  }
}

onMounted(loadVenues)

const createVenue = async () => {
  validationError.value = ''

  const name = newName.value.trim()
  const country = newCountry.value.trim()

  if (!name || !country) {
    validationError.value = 'Nombre y país son obligatorios.'
    return
  }

  const location = buildLocationPayload(createLocation)

  if (location.error) {
    validationError.value = location.error
    return
  }

  try {
    await venuesStore.createVenue({
      name,
      country,
      ...location.payload
    })

    newName.value = ''
    Object.assign(createLocation, emptyLocation())
    showCreateVenue.value = false
  } catch {
    // El store conserva el error visible.
  }
}

const startEdit = (venue) => {
  showCreateVenue.value = false
  editingVenueId.value = venue.id
  nameDraft.value = venue.name
  activeDraft.value = venue.isActive
  editCountry.value = venue.country ?? 'Chile'

  Object.assign(editLocation, {
    addressLine: venue.addressLine ?? '',
    commune: venue.commune ?? '',
    city: venue.city ?? '',
    region: venue.region ?? '',
    latitude: venue.latitude == null ? '' : String(venue.latitude),
    longitude: venue.longitude == null ? '' : String(venue.longitude)
  })

  validationError.value = ''
  venuesStore.actionError = null
  venuesStore.actionSuccess = null
}

const cancelEdit = () => {
  editingVenueId.value = null
  validationError.value = ''
}

const saveVenue = async (venueId) => {
  validationError.value = ''

  const name = nameDraft.value.trim()
  const country = editCountry.value.trim()

  if (!name || !country) {
    validationError.value = 'Nombre y país son obligatorios.'
    return
  }

  const location = buildLocationPayload(editLocation)

  if (location.error) {
    validationError.value = location.error
    return
  }

  try {
    await venuesStore.updateVenue(venueId, {
      name,
      country,
      isActive: activeDraft.value,
      ...location.payload
    })

    cancelEdit()
  } catch {
    // El store conserva el error visible.
  }
}
</script>

<template>
  <section class="venues-admin-panel">
    <div class="venues-content">
      <header class="venues-heading">
        <div>
          <h2>Sedes del polideportivo</h2>
          <p>Organiza las sedes y sus instalaciones deportivas.</p>
        </div>
        <button
          type="button"
          class="app-button primary"
          :aria-expanded="showCreateVenue"
          @click="showCreateVenue = !showCreateVenue; editingVenueId = null; validationError = ''"
        >{{ showCreateVenue ? 'Cerrar formulario' : '+ Nueva sede' }}</button>
      </header>

      <p v-if="venuesStore.loading" role="status" class="state-card">
        Cargando sedes...
      </p>

      <p v-if="venuesStore.error" role="alert" class="state-card error">
        {{ venuesStore.error }}
      </p>

      <p v-if="validationError" role="alert" class="state-card error">
        {{ validationError }}
      </p>

      <p v-if="venuesStore.actionError" role="alert" class="state-card error">
        {{ venuesStore.actionError }}
      </p>

      <p v-if="venuesStore.actionSuccess" role="status" class="state-card success">
        {{ venuesStore.actionSuccess }}
      </p>

      <div class="venues-toolbar">
        <h3>Sedes registradas</h3>
        <button
          type="button"
          class="app-button secondary"
          :disabled="venuesStore.loading"
          @click="loadVenues"
        >Actualizar sedes</button>
      </div>

      <form v-if="showCreateVenue" class="venue-form" @submit.prevent="createVenue">
        <h3>Nueva sede</h3>

        <label>
          Nombre
          <input
            v-model="newName"
            type="text"
            maxlength="150"
            required
          />
        </label>

        <label>
          País
          <input
            v-model="newCountry"
            type="text"
            maxlength="100"
            required
          />
        </label>

        <div class="location-grid">
          <label v-for="field in locationFields" :key="field.key">
            {{ field.label }}
            <input
              v-model="createLocation[field.key]"
              type="text"
              :maxlength="field.maxLength"
            />
          </label>

          <details class="coordinate-fields">
            <summary>Coordenadas (opcional)</summary>
            <div class="coordinate-grid">
              <label>
                Latitud
                <input v-model="createLocation.latitude" type="text" inputmode="decimal" placeholder="-33.4489" />
              </label>
              <label>
                Longitud
                <input v-model="createLocation.longitude" type="text" inputmode="decimal" placeholder="-70.6693" />
              </label>
            </div>
          </details>
        </div>

        <div class="venue-actions">
          <button type="submit" class="app-button primary" :disabled="venuesStore.saving">
            {{ venuesStore.saving ? 'Guardando...' : 'Crear sede' }}
          </button>
          <button type="button" class="app-button secondary" :disabled="venuesStore.saving" @click="showCreateVenue = false; validationError = ''">Cancelar</button>
        </div>
      </form>

      <p
        v-if="!venuesStore.loading &&
          !venuesStore.error &&
          !venuesStore.venues.length"
        class="state-card"
      >
        No hay sedes registradas.
      </p>

      <div
        v-for="venue in venuesStore.venues"
        :key="venue.id"
        class="venue-item"
      >
        <form
          v-if="editingVenueId === venue.id"
          class="venue-form"
          @submit.prevent="saveVenue(venue.id)"
        >
          <label>
            Nombre
            <input
              v-model="nameDraft"
              type="text"
              maxlength="150"
              required
            />
          </label>

          <label>
            País
            <input
              v-model="editCountry"
              type="text"
              maxlength="100"
              required
            />
          </label>


        <div class="location-grid">
          <label v-for="field in locationFields" :key="field.key">
            {{ field.label }}
            <input
              v-model="editLocation[field.key]"
              type="text"
              :maxlength="field.maxLength"
            />
          </label>

          <details class="coordinate-fields">
            <summary>Coordenadas (opcional)</summary>
            <div class="coordinate-grid">
              <label>
                Latitud
                <input v-model="editLocation.latitude" type="text" inputmode="decimal" placeholder="-33.4489" />
              </label>
              <label>
                Longitud
                <input v-model="editLocation.longitude" type="text" inputmode="decimal" placeholder="-70.6693" />
              </label>
            </div>
          </details>
        </div>

          <label class="checkbox-label">
            <input v-model="activeDraft" type="checkbox" />
            Sede activa
          </label>

          <div class="venue-actions">
            <button type="submit" class="app-button primary" :disabled="venuesStore.saving">
              Guardar cambios
            </button>

            <button
              type="button"
              class="app-button secondary"
              :disabled="venuesStore.saving"
              @click="cancelEdit"
            >
              Cancelar
            </button>
          </div>
        </form>

        <template v-else>
          <div class="venue-summary">
            <strong>{{ venue.name }}</strong>

            <p class="venue-metadata">
              <span class="venue-status" :class="venue.isActive ? 'is-active' : 'is-inactive'">
                {{ venue.isActive ? 'Activa' : 'Inactiva' }}
              </span>
              <span>{{ venue.resourceCount ?? 0 }} recursos</span>
              <span>{{ venue.activeResourceCount ?? 0 }} activos</span>
            </p>
            <p v-if="venue.commune || venue.city || venue.region || venue.country">
              {{ [
                venue.commune,
                venue.city,
                venue.region,
                venue.country
              ].filter(Boolean).join(', ') }}
            </p>
            <p v-if="venue.addressLine">
              {{ venue.addressLine }}
            </p>
          </div>

          <button type="button" class="app-button secondary" @click="startEdit(venue)">
            Editar sede
          </button>
        </template>
      </div>
    </div>
  </section>
</template>

<style scoped>
.venues-admin-panel {
  padding: var(--space-5);
  min-width: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.venues-content {
  display: grid;
  gap: var(--space-4);
}

.venues-heading,
.venues-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.venues-heading > div { min-width: 0; }

.venues-heading h2 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-section-title);
  line-height: 1.25;
}

.venues-toolbar h3 {
  margin: 0;
  color: var(--color-text);
  font-size: var(--text-body);
}

.venues-heading p {
  margin: var(--space-2) 0 0;
  color: var(--color-text-muted);
  line-height: 1.5;
}

.venue-form {
  display: grid;
  min-width: 0;
  gap: var(--space-4);
  width: 100%;
  padding: var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface-muted);
}

.venue-form h3 {
  margin: 0;
  font-size: var(--text-section-title);
  color: var(--color-text);
}

.venue-form label,
.coordinate-grid label {
  display: grid;
  min-width: 0;
  gap: var(--space-2);
  color: var(--color-text);
  font-size: var(--text-help);
  font-weight: 700;
}

.venue-form input:not([type='checkbox']) {
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

.venue-form .checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.checkbox-label input {
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.venue-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  min-width: 0;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
}

.venue-item:has(.venue-form) { display: block; }
.venue-summary { flex: 1 1 220px; min-width: 0; }
.venue-summary strong { color: var(--color-text); overflow-wrap: anywhere; }

.venue-item p {
  margin: var(--space-2) 0 0;
  color: var(--color-text-muted);
  font-size: var(--text-help);
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.venue-metadata {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-3);
}

.venue-metadata span:not(.venue-status) { color: var(--color-text-muted); }

.venue-status {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 3px var(--space-3);
  border-radius: var(--radius-pill);
  font-size: 12px;
  font-weight: 750;
}

.venue-status.is-active {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.venue-status.is-inactive {
  background: var(--color-surface-soft);
  color: var(--color-text-muted);
}

.venue-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.location-grid,
.coordinate-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.coordinate-fields {
  grid-column: 1 / -1;
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
}

.coordinate-fields summary {
  cursor: pointer;
  font-weight: 700;
  color: var(--color-primary-strong);
}

.coordinate-fields summary:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 3px;
}

.coordinate-grid { margin-top: var(--space-3); }

@media (max-width: 600px) {
  .location-grid,
  .coordinate-grid { grid-template-columns: 1fr; }
  .venues-heading > .app-button { width: 100%; }
  .venue-item > .app-button { width: 100%; }
}

@media (max-width: 480px) {
  .venues-admin-panel { padding: var(--space-4); }
  .venue-form { padding: var(--space-3); }
  .venue-actions .app-button { flex: 1 1 140px; }
}
</style>
