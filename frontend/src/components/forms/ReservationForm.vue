<script setup>
import { ref, watch } from 'vue'

import ResourcePicker from './ResourcePicker.vue'
import DateTimePicker from './DateTimePicker.vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },

  slot: {
    type: Object,
    default: null
  },

  resources: {
    type: Array,
    default: () => []
  },

  errorMessage: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'close',
  'submit'
])

/* FORM */
const form = ref({
  resource: null,

  date: '',

  hour: '',

  durationMinutes: 60,

  sport: '',

  participants: 10
})

/* AUTOFILL FROM SLOT */
watch(
  () => props.slot,
  (slot) => {
    if (!slot) return

    form.value.resource =
      slot.resource || null

    form.value.hour =
      slot.hour || ''

    form.value.date =
      slot.date ||
      new Date().toISOString().slice(0, 10)

    form.value.durationMinutes = 60
  },
  {
    immediate: true
  }
)

/* RESOURCE */
const handleResourceSelect = (resource) => {
  form.value.resource = resource
}

/* DATETIME */
const handleDateTimeUpdate = (data) => {
  form.value.date =
    data.date

  form.value.hour =
    data.hour

  form.value.durationMinutes =
    data.durationMinutes
}

/* SUBMIT */
const handleSubmit = () => {
  if (!form.value.resource) {
    return
  }

  if (!form.value.date) {
    return
  }

  if (!form.value.hour) {
    return
  }

  emit('submit', {
    ...form.value
  })
}

/* CLOSE */
const handleClose = () => {
  emit('close')
}
</script>

<template>
  <Teleport to="body">

    <div
      v-if="visible"
      class="overlay"
      @click.self="handleClose"
    >

      <div class="modal">

        <!-- HEADER -->
        <div class="modal-header">

          <div>

            <h2>
              Crear Reserva
            </h2>

            <p>
              Completa la información de la reserva.
            </p>

          </div>

          <button
            class="close-btn"
            @click="handleClose"
          >
            ✕
          </button>

        </div>

        <!-- SUMMARY -->
        <div
          v-if="form.resource"
          class="summary"
        >

          <div>

            <span>
              Instalación
            </span>

            <strong>
              {{ form.resource.name }}
            </strong>

          </div>

          <div>

            <span>
              Inicio
            </span>

            <strong>
              {{ form.date }} · {{ form.hour }}
            </strong>

          </div>

        </div>

        <!-- RESOURCE -->
        <ResourcePicker
          :resources="resources"
          :selected-id="form.resource?.id"
          @select="handleResourceSelect"
        />

        <!-- DATE TIME -->
        <DateTimePicker
          :initial-date="form.date"
          :initial-hour="form.hour"
          :initial-duration="form.durationMinutes"
          @update="handleDateTimeUpdate"
        />

        <!-- SPORT -->
        <div class="field">

          <label>
            Actividad / deporte
          </label>

          <input
            v-model="form.sport"
            type="text"
            placeholder="Ej: Fútbol, vóleibol, entrenamiento"
          />

        </div>

        <!-- PARTICIPANTS -->
        <div class="field">

          <label>
            Participantes estimados
          </label>

          <input
            v-model="form.participants"
            type="number"
            min="1"
          />

        </div>

        <!-- FORM ERROR -->
        <div
          v-if="errorMessage"
          class="form-error"
        >
          {{ errorMessage }}
        </div>

        <!-- FOOTER -->
        <div class="actions">

          <button
            class="cancel-btn"
            @click="handleClose"
          >
            Cancelar
          </button>

          <button
            class="submit-btn"
            @click="handleSubmit"
          >
            Confirmar Reserva
          </button>

        </div>

      </div>

    </div>

  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15,23,42,0.55);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 20px;

  z-index: 9999;

  backdrop-filter: blur(4px);
}

/* MODAL */
.modal {
  width: 100%;
  max-width: 720px;

  max-height: 90vh;

  overflow-y: auto;

  background: white;

  border-radius: 28px;

  padding: 28px;

  display: flex;
  flex-direction: column;

  gap: 24px;

  box-shadow:
    0 24px 60px rgba(0,0,0,0.2);
}

/* HEADER */
.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;
}

.modal-header h2 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: #0f172a;
}

.modal-header p {
  margin-top: 6px;

  color: #64748b;

  font-size: 14px;
}

/* CLOSE */
.close-btn {
  width: 44px;
  height: 44px;

  border: none;

  border-radius: 14px;

  background: #f1f5f9;

  color: #334155;

  cursor: pointer;

  font-size: 18px;

  transition: 0.2s;
}

.close-btn:hover {
  background: #e2e8f0;
}

/* SUMMARY */
.summary {
  display: grid;

  grid-template-columns:
    repeat(2, 1fr);

  gap: 14px;

  background: #eff6ff;

  border: 1px solid #bfdbfe;

  border-radius: 20px;

  padding: 16px;
}

.summary div {
  display: flex;
  flex-direction: column;

  gap: 4px;
}

.summary span {
  font-size: 12px;
  font-weight: 700;

  color: #64748b;

  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.summary strong {
  font-size: 15px;

  color: #1e3a8a;
}

/* FIELD */
.field {
  display: flex;
  flex-direction: column;

  gap: 8px;
}

.field label {
  font-size: 14px;
  font-weight: 700;

  color: #334155;
}

.field input {
  width: 100%;

  height: 50px;

  border: 1px solid #dbe2ea;

  border-radius: 16px;

  padding: 0 16px;

  font-size: 14px;

  outline: none;

  box-sizing: border-box;

  transition: 0.2s;
}

.field input:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}

/* ERROR */
.form-error {
  padding: 14px 16px;

  border-radius: 16px;

  background: #fee2e2;

  border: 1px solid #fecaca;

  color: #b91c1c;

  font-size: 14px;
  font-weight: 700;
}

/* ACTIONS */
.actions {
  display: flex;
  justify-content: flex-end;

  gap: 14px;

  padding-top: 8px;
}

.actions button {
  height: 50px;

  border: none;

  border-radius: 16px;

  padding: 0 22px;

  font-size: 14px;
  font-weight: 700;

  cursor: pointer;

  transition: 0.2s;
}

.cancel-btn {
  background: #f1f5f9;

  color: #334155;
}

.cancel-btn:hover {
  background: #e2e8f0;
}

.submit-btn {
  background: linear-gradient(
    135deg,
    #2563eb,
    #1d4ed8
  );

  color: white;
}

.submit-btn:hover {
  transform: translateY(-1px);

  box-shadow:
    0 10px 20px rgba(37,99,235,0.25);
}

/* MOBILE */
@media (max-width: 768px) {
  .modal {
    padding: 22px;

    border-radius: 24px;
  }

  .summary {
    grid-template-columns: 1fr;
  }

  .actions {
    flex-direction: column;
  }

  .actions button {
    width: 100%;
  }
}
</style>