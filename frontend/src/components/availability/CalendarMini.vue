<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  month: {
    type: Number,
    default: new Date().getMonth()
  },

  year: {
    type: Number,
    default: new Date().getFullYear()
  }
})

const emit = defineEmits([
  'select-date'
])

const currentMonth = ref(props.month)
const currentYear = ref(props.year)

const monthNames = [
  'Enero',
  'Febrero',
  'Marzo',
  'Abril',
  'Mayo',
  'Junio',
  'Julio',
  'Agosto',
  'Septiembre',
  'Octubre',
  'Noviembre',
  'Diciembre'
]

const weekDays = [
  'Lu',
  'Ma',
  'Mi',
  'Ju',
  'Vi',
  'Sa',
  'Do'
]

/* DAYS */
const days = computed(() => {
  const firstDay = new Date(
    currentYear.value,
    currentMonth.value,
    1
  )

  const lastDay = new Date(
    currentYear.value,
    currentMonth.value + 1,
    0
  )

  let startingDay = firstDay.getDay()

  if (startingDay === 0) {
    startingDay = 7
  }

  const totalDays = lastDay.getDate()

  const result = []

  /* Empty */
  for (let i = 1; i < startingDay; i++) {
    result.push(null)
  }

  /* Real days */
  for (let day = 1; day <= totalDays; day++) {
    result.push(day)
  }

  return result
})

/* TODAY */
const today = new Date()

const isToday = (day) => {
  return (
    day === today.getDate() &&
    currentMonth.value === today.getMonth() &&
    currentYear.value === today.getFullYear()
  )
}

/* SELECT */
const selectDate = (day) => {
  if (!day) return

  emit('select-date', {
    day,
    month: currentMonth.value,
    year: currentYear.value
  })
}

/* NAVIGATION */
const prevMonth = () => {
  if (currentMonth.value === 0) {
    currentMonth.value = 11
    currentYear.value--
  } else {
    currentMonth.value--
  }
}

const nextMonth = () => {
  if (currentMonth.value === 11) {
    currentMonth.value = 0
    currentYear.value++
  } else {
    currentMonth.value++
  }
}
</script>

<template>
  <div class="calendar">

    <!-- HEADER -->
    <div class="calendar-header">

      <button
        class="nav-button"
        @click="prevMonth"
      >
        ←
      </button>

      <h3>
        {{ monthNames[currentMonth] }}
        {{ currentYear }}
      </h3>

      <button
        class="nav-button"
        @click="nextMonth"
      >
        →
      </button>

    </div>

    <!-- WEEK DAYS -->
    <div class="weekdays">

      <div
        v-for="day in weekDays"
        :key="day"
        class="weekday"
      >
        {{ day }}
      </div>

    </div>

    <!-- DAYS -->
    <div class="days-grid">

      <button
        v-for="(day, index) in days"
        :key="index"

        class="day"

        :class="{
          empty: !day,
          today: isToday(day)
        }"

        @click="selectDate(day)"
      >

        {{ day }}

      </button>

    </div>

  </div>
</template>

<style scoped>
.calendar {
  width: 100%;
  max-width: 340px;

  background: white;

  border-radius: 24px;

  padding: 20px;

  border: 1px solid #e2e8f0;

  box-shadow:
    0 6px 20px rgba(0,0,0,0.05);

  display: flex;
  flex-direction: column;

  gap: 20px;
}

/* HEADER */
.calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.calendar-header h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 700;

  color: #0f172a;
}

/* BUTTON */
.nav-button {
  width: 38px;
  height: 38px;

  border: none;

  border-radius: 12px;

  background: #eff6ff;

  color: #2563eb;

  cursor: pointer;

  transition: 0.2s;
}

.nav-button:hover {
  background: #dbeafe;
}

/* WEEK */
.weekdays {
  display: grid;

  grid-template-columns:
    repeat(7, 1fr);

  gap: 8px;
}

.weekday {
  text-align: center;

  font-size: 13px;
  font-weight: 700;

  color: #64748b;
}

/* DAYS */
.days-grid {
  display: grid;

  grid-template-columns:
    repeat(7, 1fr);

  gap: 8px;
}

/* DAY */
.day {
  aspect-ratio: 1;

  border: none;

  border-radius: 14px;

  background: white;

  cursor: pointer;

  font-size: 14px;
  font-weight: 600;

  color: #334155;

  transition: 0.2s;
}

.day:hover {
  background: #eff6ff;
}

/* TODAY */
.today {
  background: linear-gradient(
    135deg,
    #2563eb,
    #1d4ed8
  );

  color: white;

  box-shadow:
    0 8px 18px rgba(37,99,235,0.25);
}

/* EMPTY */
.empty {
  visibility: hidden;

  pointer-events: none;
}

/* Mobile */
@media (max-width: 768px) {
  .calendar {
    max-width: 100%;
  }
}
</style>