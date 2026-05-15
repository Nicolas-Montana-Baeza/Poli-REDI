<script setup>
import { ref, watch } from 'vue'

import ResourcePicker from './ResourcePicker.vue'
import DateTimePicker from './DateTimePicker.vue'

/* PROPS */
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

    required: true
  }
})

/* EMITS */
const emit = defineEmits([
  'close',
  'submit'
])

/* FORM */
const form = ref({
  resource: null,

  date: '',

  hour: '',

  duration: 1,

  sport: '',

  participants: 10
})

/* SLOT WATCH */
watch(
  () => props.slot,

  (slot) => {

    if (!slot) return

    form.value.resource =
      slot.resource

    form.value.hour =
      slot.hour

    if (!form.value.date) {

      form.value.date =
        new Date()
          .toISOString()
          .split('T')[0]
    }
  },

  {
    immediate: true
  }
)

/* RESOURCE */
const handleResourceSelect = (
  resource
) => {

  form.value.resource =
    resource
}

/* DATETIME */
const handleDateTimeUpdate = (
  data
) => {

  form.value.date =
    data.date

  form.value.hour =
    data.hour

  form.value.duration =
    data.duration
}

/* SUBMIT */
const handleSubmit = () => {

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
  <div
    v-if="visible"
    class="overlay"
  >

    <div class="modal">

      <!-- HEADER -->
      <div class="modal-header">

        <div>

          <h2>
            Crear Reserva
          </h2>

          <p>
            Completa la información
            de la reserva.
          </p>

        </div>

        <button
          class="close-btn"
          @click="handleClose"
        >
          ✕
        </button>

      </div>

      <!-- RESOURCE -->
      <ResourcePicker
        :resources="props.resources"

        :selected-id="
          form.resource?.id
        "

        @select="
          handleResourceSelect
        "
      />

      <!-- DATETIME -->
      <DateTimePicker
        :initial-date="
          form.date
        "

        :initial-hour="
          form.hour
        "

        @update="
          handleDateTimeUpdate
        "
      />

      <!-- SPORT -->
      <div class="field">

        <label>
          Deporte
        </label>

        <input
          v-model="form.sport"
          placeholder="Ej: fútbol"
        />

      </div>

      <!-- PARTICIPANTS -->
      <div class="field">

        <label>
          Participantes
        </label>

        <input
          type="number"

          min="1"

          v-model="
            form.participants
          "
        />

      </div>

      <!-- ACTIONS -->
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
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15,23,42,0.45);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 20px;

  z-index: 9999;
}

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
    0 20px 50px rgba(0,0,0,0.15);
}

.modal-header {
  display: flex;
  align-items: start;
  justify-content: space-between;

  gap: 20px;
}

.modal-header h2 {
  margin: 0;

  font-size: 34px;
  font-weight: 800;

  color: #0f172a;
}

.modal-header p {
  margin-top: 6px;

  color: #64748b;
}

.close-btn {
  width: 48px;
  height: 48px;

  border: none;
  border-radius: 16px;

  background: #f1f5f9;

  cursor: pointer;

  font-size: 18px;
}

.field {
  display: flex;
  flex-direction: column;

  gap: 8px;
}

.field label {
  font-size: 14px;
  font-weight: 700;

  color: #0f172a;
}

.field input {
  height: 52px;

  border-radius: 16px;

  border: 1px solid #cbd5e1;

  padding: 0 16px;

  font-size: 15px;
}

.actions {
  display: flex;
  justify-content: flex-end;

  gap: 14px;
}

.cancel-btn,
.submit-btn {
  height: 52px;

  padding: 0 22px;

  border: none;
  border-radius: 16px;

  font-weight: 700;

  cursor: pointer;
}

.cancel-btn {
  background: #e2e8f0;
}

.submit-btn {
  background: #2563eb;

  color: white;
}
</style>