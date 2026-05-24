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
  },

  initialDuration: {
    type: Number,
    default: 60
  }
})

const emit = defineEmits([
  'update'
])

const selectedDate = ref('')
const selectedHour = ref('')
const durationMinutes = ref(60)

const durationOptions = [
  {
    label: '30 minutos',
    value: 30
  },
  {
    label: '45 minutos',
    value: 45
  },
  {
    label: '1 hora',
    value: 60
  },
  {
    label: '1 hora 30 min',
    value: 90
  },
  {
    label: '2 horas',
    value: 120
  },
  {
    label: '3 horas',
    value: 180
  }
]

watch(
  () => [
    props.initialDate,
    props.initialHour,
    props.initialDuration
  ],
  () => {
    selectedDate.value =
      props.initialDate || ''

    selectedHour.value =
      props.initialHour || ''

    durationMinutes.value =
      props.initialDuration || 60
  },
  {
    immediate: true
  }
)

const updateValues = () => {
  emit('update', {
    date: selectedDate.value,
    hour: selectedHour.value,
    durationMinutes: Number(durationMinutes.value)
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
        Ajusta la fecha, hora exacta y duración.
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
        Hora de inicio
      </label>

      <input
        v-model="selectedHour"
        type="time"
        step="60"
        @change="updateValues"
      />

      <small>
        Puedes ajustar por minuto si es necesario.
      </small>

    </div>

    <!-- DURATION -->
    <div class="field">

      <label>
        Duración
      </label>

      <select
        v-model="durationMinutes"
        @change="updateValues"
      >

        <option
          v-for="option in durationOptions"
          :key="option.value"
          :value="option.value"
        >
          {{ option.label }}
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
  font-weight: 700;

  color: #334155;
}

.field small {
  font-size: 12px;

  color: #64748b;
}

/* INPUTS */
.field input,
.field select {
  width: 100%;

  height: 50px;

  border: 1px solid #dbe2ea;

  border-radius: 16px;

  padding: 0 16px;

  font-size: 14px;

  background: white;

  transition: 0.2s;

  outline: none;

  box-sizing: border-box;
}

.field input:focus,
.field select:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}
</style>