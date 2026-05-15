<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  initialDate: {
    type: String,
    default: ''
  },

  initialHour: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'update'
])

/* STATE */
const selectedDate = ref('')
const selectedHour = ref('')
const duration = ref(1)

/* HOURS */
const hours = [
  '08:00',
  '09:00',
  '10:00',
  '11:00',
  '12:00',
  '13:00',
  '14:00',
  '15:00',
  '16:00',
  '17:00',
  '18:00',
  '19:00',
  '20:00',
  '21:00'
]

/* INITIAL VALUES */
watch(
  () => [
    props.initialDate,
    props.initialHour
  ],
  () => {

    selectedDate.value =
      props.initialDate || ''

    selectedHour.value =
      props.initialHour || ''
  },
  {
    immediate: true
  }
)

/* UPDATE */
const updateValues = () => {

  emit('update', {
    date: selectedDate.value,

    hour: selectedHour.value,

    duration: duration.value
  })
}
</script>


<template>
  <div class="picker">

    <!-- HEADER -->
    <div class="header">

      <h3>
        Fecha y horario
      </h3>

      <p>
        Define cuándo deseas reservar.
      </p>

    </div>

    <!-- DATE -->
    <div class="field">

      <label>
        Fecha
      </label>

      <input
        v-model="selectedDate"
        type="date"
        @change="updateValues"
      />

    </div>

    <!-- HOUR -->
    <div class="field">

      <label>
        Hora
      </label>

      <select
        v-model="selectedHour"
        @change="updateValues"
      >

        <option
          disabled
          value=""
        >
          Selecciona horario
        </option>

        <option
          v-for="hour in hours"
          :key="hour"
          :value="hour"
        >
          {{ hour }}
        </option>

      </select>

    </div>

    <!-- DURATION -->
    <div class="field">

      <label>
        Duración
      </label>

      <select
        v-model="duration"
        @change="updateValues"
      >

        <option :value="1">
          1 hora
        </option>

        <option :value="2">
          2 horas
        </option>

        <option :value="3">
          3 horas
        </option>

      </select>

    </div>

  </div>
</template>

<style scoped>
.picker {
  display: flex;
  flex-direction: column;

  gap: 20px;
}

/* HEADER */
.header h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 700;

  color: #0f172a;
}

.header p {
  margin-top: 4px;

  font-size: 14px;

  color: #64748b;
}

/* FIELD */
.field {
  display: flex;
  flex-direction: column;

  gap: 8px;
}

.field label {
  font-size: 14px;
  font-weight: 600;

  color: #334155;
}

/* INPUTS */
.field input,
.field select {
  width: 100%;

  height: 48px;

  border: 1px solid #dbe2ea;

  border-radius: 14px;

  padding: 0 14px;

  font-size: 14px;

  background: white;

  transition: 0.2s;

  outline: none;
}

.field input:focus,
.field select:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}
</style>