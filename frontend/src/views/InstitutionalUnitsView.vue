<script setup>
import {
  computed,
  onMounted,
  reactive,
  ref
} from 'vue'

import { useInstitutionalUnitsStore } from '../stores/institutionalUnits'

const unitsStore =
  useInstitutionalUnitsStore()

const search = ref('')
const showCreateForm = ref(false)
const successMessage = ref('')

const form = reactive({
  name: '',
  code: '',
  unitType: 'ACADEMIC_PROGRAM'
})

// ============================================================================
// CATÁLOGOS DE PRESENTACIÓN
// ============================================================================
//
// Los valores enviados al backend conservan exactamente el dominio
// institucional definido por MVP2. Las etiquetas son únicamente una
// representación amigable para el administrador.

const unitTypes = [
  {
    value: 'ACADEMIC_PROGRAM',
    label: 'Programa académico'
  },
  {
    value: 'POSTGRADUATE_PROGRAM',
    label: 'Programa de postgrado'
  },
  {
    value: 'SPORTS_UNIT',
    label: 'Unidad deportiva'
  },
  {
    value: 'ADMINISTRATIVE_UNIT',
    label: 'Unidad administrativa'
  },
  {
    value: 'OTHER',
    label: 'Otra'
  }
]

const unitTypeLabels = Object.fromEntries(
  unitTypes.map(
    (type) => [
      type.value,
      type.label
    ]
  )
)

const filteredUnits = computed(() => {
  const query = search.value
    .trim()
    .toLocaleLowerCase('es')

  const units =
    unitsStore.activeUnits || []

  if (!query) {
    return units
  }

  return units.filter((unit) => {
    const typeLabel =
      unitTypeLabels[unit.unitType] ||
      unit.unitType ||
      ''

    return [
      unit.name,
      unit.code,
      typeLabel
    ].some((value) => (
      String(value || '')
        .toLocaleLowerCase('es')
        .includes(query)
    ))
  })
})

const resetForm = () => {
  form.name = ''
  form.code = ''
  form.unitType = 'ACADEMIC_PROGRAM'

  unitsStore.clearError()
}

const openCreateForm = () => {
  successMessage.value = ''
  resetForm()

  showCreateForm.value = true
}

const closeCreateForm = () => {
  showCreateForm.value = false
  resetForm()
}

const submitUnit = async () => {
  successMessage.value = ''
  unitsStore.clearError()

  const name =
    form.name.trim()

  const code =
    form.code.trim().toUpperCase()

  if (!name || !code) {
    unitsStore.error =
      'Nombre y código son obligatorios.'

    return
  }

  try {
    const unit =
      await unitsStore.createUnit({
        name,
        code,
        unitType: form.unitType
      })

    successMessage.value =
      `Unidad "${unit.name}" creada correctamente.`

    showCreateForm.value = false
    resetForm()
  } catch {
    // El store ya mantiene un mensaje de error apto para la interfaz.
  }
}

const reload = async () => {
  successMessage.value = ''

  try {
    await unitsStore.loadUnits()
  } catch {
    // El error se presenta desde el store.
  }
}

onMounted(reload)
</script>

<template>
  <main class="institutional-units-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">
          Programación institucional
        </p>

        <h1>
          Unidades institucionales
        </h1>

        <p class="page-description">
          Administra las estructuras institucionales que pueden
          programar actividades en los recursos de Poli-REDI.
        </p>
      </div>

      <button
        class="primary-button"
        type="button"
        @click="openCreateForm"
      >
        Nueva unidad
      </button>
    </header>

    <section
      v-if="successMessage"
      class="message success-message"
      role="status"
    >
      {{ successMessage }}
    </section>

    <section
      v-if="unitsStore.error"
      class="message error-message"
      role="alert"
    >
      <span>
        {{ unitsStore.error }}
      </span>

      <button
        type="button"
        class="text-button"
        @click="unitsStore.clearError"
      >
        Cerrar
      </button>
    </section>

    <section class="toolbar">
      <label class="search-field">
        <span>
          Buscar unidad
        </span>

        <input
          v-model="search"
          type="search"
          placeholder="Nombre, código o tipo"
        >
      </label>

      <div class="unit-counter">
        {{ filteredUnits.length }}
        {{
          filteredUnits.length === 1
            ? 'unidad'
            : 'unidades'
        }}
      </div>
    </section>

    <section
      v-if="unitsStore.loading"
      class="state-card"
    >
      Cargando unidades institucionales...
    </section>

    <section
      v-else-if="filteredUnits.length === 0"
      class="state-card"
    >
      <template v-if="search">
        No hay unidades que coincidan con la búsqueda.
      </template>

      <template v-else>
        Todavía no existen unidades institucionales disponibles.
      </template>
    </section>

    <section
      v-else
      class="units-grid"
      aria-label="Unidades institucionales"
    >
      <article
        v-for="unit in filteredUnits"
        :key="unit.id"
        class="unit-card"
      >
        <div class="unit-card-header">
          <div>
            <h2>
              {{ unit.name }}
            </h2>

            <span class="unit-code">
              {{ unit.code }}
            </span>
          </div>

          <span
            class="status-badge"
            :class="{
              active: unit.isActive
            }"
          >
            {{
              unit.isActive
                ? 'Activa'
                : 'Inactiva'
            }}
          </span>
        </div>

        <dl class="unit-details">
          <div>
            <dt>
              Tipo
            </dt>

            <dd>
              {{
                unitTypeLabels[unit.unitType] ||
                unit.unitType
              }}
            </dd>
          </div>

          <div>
            <dt>
              Identificador
            </dt>

            <dd>
              #{{ unit.id }}
            </dd>
          </div>
        </dl>

        <p class="next-step-note">
          La administración de miembros y gestores se habilitará
          en el siguiente bloque de programación institucional.
        </p>
      </article>
    </section>

    <Teleport to="body">
      <div
        v-if="showCreateForm"
        class="modal-backdrop"
        @click.self="closeCreateForm"
      >
        <section
          class="modal-card"
          role="dialog"
          aria-modal="true"
          aria-labelledby="create-unit-title"
        >
          <header class="modal-header">
            <div>
              <p class="eyebrow">
                Administración
              </p>

              <h2 id="create-unit-title">
                Nueva unidad institucional
              </h2>
            </div>

            <button
              class="close-button"
              type="button"
              aria-label="Cerrar"
              @click="closeCreateForm"
            >
              ×
            </button>
          </header>

          <form
            class="unit-form"
            @submit.prevent="submitUnit"
          >
            <label>
              <span>
                Nombre
              </span>

              <input
                v-model="form.name"
                type="text"
                autocomplete="off"
                placeholder="Ej. Escuela de Informática"
                required
              >
            </label>

            <label>
              <span>
                Código institucional
              </span>

              <input
                v-model="form.code"
                type="text"
                autocomplete="off"
                placeholder="Ej. INF"
                required
              >

              <small>
                El backend normaliza el código a mayúsculas.
              </small>
            </label>

            <label>
              <span>
                Tipo de unidad
              </span>

              <select
                v-model="form.unitType"
                required
              >
                <option
                  v-for="type in unitTypes"
                  :key="type.value"
                  :value="type.value"
                >
                  {{ type.label }}
                </option>
              </select>
            </label>

            <div
              v-if="unitsStore.error"
              class="form-error"
              role="alert"
            >
              {{ unitsStore.error }}
            </div>

            <footer class="modal-actions">
              <button
                class="secondary-button"
                type="button"
                :disabled="unitsStore.creating"
                @click="closeCreateForm"
              >
                Cancelar
              </button>

              <button
                class="primary-button"
                type="submit"
                :disabled="unitsStore.creating"
              >
                {{
                  unitsStore.creating
                    ? 'Creando...'
                    : 'Crear unidad'
                }}
              </button>
            </footer>
          </form>
        </section>
      </div>
    </Teleport>
  </main>
</template>

<style scoped>
.institutional-units-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.5rem;
}

.page-header h1,
.modal-header h2 {
  margin: 0;
}

.eyebrow {
  margin: 0 0 0.35rem;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  opacity: 0.65;
}

.page-description {
  max-width: 46rem;
  margin: 0.5rem 0 0;
  line-height: 1.5;
  opacity: 0.75;
}

.toolbar {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1rem;
}

.search-field {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.4rem;
  max-width: 32rem;
  font-weight: 600;
}

.search-field input,
.unit-form input,
.unit-form select {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(127, 127, 127, 0.35);
  border-radius: 0.7rem;
  padding: 0.75rem 0.85rem;
  background: inherit;
  color: inherit;
  font: inherit;
}

.unit-counter {
  opacity: 0.7;
}

.units-grid {
  display: grid;
  grid-template-columns: repeat(
    auto-fit,
    minmax(260px, 1fr)
  );
  gap: 1rem;
}

.unit-card,
.state-card {
  border: 1px solid rgba(127, 127, 127, 0.25);
  border-radius: 1rem;
  padding: 1.2rem;
  background: rgba(127, 127, 127, 0.04);
}

.unit-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.unit-card h2 {
  margin: 0;
  font-size: 1.05rem;
}

.unit-code {
  display: inline-block;
  margin-top: 0.35rem;
  font-family: monospace;
  opacity: 0.65;
}

.status-badge {
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font-size: 0.72rem;
  font-weight: 700;
  background: rgba(127, 127, 127, 0.15);
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.15);
}

.unit-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.8rem;
  margin: 1.2rem 0 0;
}

.unit-details div {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.unit-details dt {
  font-size: 0.75rem;
  font-weight: 700;
  opacity: 0.55;
}

.unit-details dd {
  margin: 0;
}

.next-step-note {
  margin: 1rem 0 0;
  font-size: 0.82rem;
  line-height: 1.4;
  opacity: 0.6;
}

.message {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-radius: 0.8rem;
  padding: 0.85rem 1rem;
}

.success-message {
  background: rgba(34, 197, 94, 0.12);
}

.error-message,
.form-error {
  background: rgba(239, 68, 68, 0.12);
}

.form-error {
  border-radius: 0.7rem;
  padding: 0.75rem;
}

.primary-button,
.secondary-button,
.text-button,
.close-button {
  border: 0;
  cursor: pointer;
  font: inherit;
}

.primary-button,
.secondary-button {
  border-radius: 0.7rem;
  padding: 0.7rem 1rem;
  font-weight: 700;
}

.primary-button {
  background: #2563eb;
  color: white;
}

.secondary-button {
  background: rgba(127, 127, 127, 0.15);
  color: inherit;
}

.primary-button:disabled,
.secondary-button:disabled {
  cursor: wait;
  opacity: 0.55;
}

.text-button,
.close-button {
  background: transparent;
  color: inherit;
}

.text-button {
  font-weight: 700;
}

.modal-backdrop {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.55);
}

.modal-card {
  width: min(100%, 34rem);
  border-radius: 1rem;
  padding: 1.3rem;
  background: var(--color-background, #ffffff);
  color: var(--color-text, #1f2937);
  box-shadow: 0 1.5rem 4rem rgba(0, 0, 0, 0.25);
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.close-button {
  font-size: 1.6rem;
  line-height: 1;
}

.unit-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1.3rem;
}

.unit-form label {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  font-weight: 600;
}

.unit-form small {
  font-weight: 400;
  opacity: 0.6;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.7rem;
  margin-top: 0.5rem;
}

@media (max-width: 700px) {
  .page-header,
  .toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .search-field {
    max-width: none;
  }

  .unit-details {
    grid-template-columns: 1fr;
  }
}
</style>
