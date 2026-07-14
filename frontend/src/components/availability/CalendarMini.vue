<script setup>
import { computed, ref, watch } from 'vue'
import { getBusinessDateKey } from '@/utils/reservationTime'

const props = defineProps({
  month: {
    type: Number,
    default: new Date().getMonth()
  },

  year: {
    type: Number,
    default: new Date().getFullYear()
  },

  selectedDate: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'select-date'
])

const currentMonth = ref(props.month)
const currentYear = ref(props.year)

const syncVisibleMonth = (dateKey) => {
  if (!dateKey) {
    return
  }

  const date = new Date(`${dateKey}T00:00:00`)

  if (Number.isNaN(date.getTime())) {
    return
  }

  currentMonth.value = date.getMonth()
  currentYear.value = date.getFullYear()
}

watch(
  () => props.selectedDate,
  syncVisibleMonth,
  {
    immediate: true
  }
)

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
const toDateKey = (year, month, day) => {
  return [
    year,
    String(month + 1).padStart(2, '0'),
    String(day).padStart(2, '0')
  ].join('-')
}

const todayKey = computed(() => {
  return getBusinessDateKey()
})

const getDayKey = (day) => {
  if (!day) {
    return ''
  }

  return toDateKey(
    currentYear.value,
    currentMonth.value,
    day
  )
}

const isToday = (day) => {
  return getDayKey(day) === todayKey.value
}

const isSelected = (day) => {
  return getDayKey(day) === props.selectedDate
}

const isPast = (day) => {
  const key = getDayKey(day)

  return Boolean(key) && key < todayKey.value
}

/* SELECT */
const selectDate = (day) => {
  if (!day) return

  if (isPast(day)) return

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
        type="button"
        aria-label="Mes anterior"
        @click="prevMonth"
      >
        ‹
      </button>

      <h3>
        {{ monthNames[currentMonth] }}
        {{ currentYear }}
      </h3>

      <button
        class="nav-button"
        type="button"
        aria-label="Mes siguiente"
        @click="nextMonth"
      >
        ›
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
          today: isToday(day),
          selected: isSelected(day),
          past: isPast(day)
        }"

        :disabled="!day || isPast(day)"
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

  background: var(--color-surface);

  border-radius: var(--radius-xl);

  padding: var(--space-4);

  border: 1px solid var(--color-border);

  box-shadow: var(--shadow-card);

  display: flex;
  flex-direction: column;

  gap: var(--space-4);
}

/* HEADER */
.calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.calendar-header h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 800;

  color: var(--color-text);
}

/* BUTTON */
.nav-button {
  width: 34px;
  height: 34px;

  border: 1px solid var(--color-border);

  border-radius: var(--radius-md);

  background: var(--color-surface);

  color: var(--color-primary);

  cursor: pointer;

  font-size: 22px;
  line-height: 1;

  transition: 0.2s;
}

.nav-button:hover {
  background: var(--color-primary-soft);
  border-color: #bfd3ff;
}

/* WEEK */
.weekdays {
  display: grid;

  grid-template-columns:
    repeat(7, 1fr);

  gap: 6px;
}

.weekday {
  text-align: center;

  font-size: 12px;
  font-weight: 800;

  color: var(--color-text-muted);
}

/* DAYS */
.days-grid {
  display: grid;

  grid-template-columns:
    repeat(7, 1fr);

  gap: 6px;
}

/* DAY */
.day {
  aspect-ratio: 1;

  border: 1px solid transparent;

  border-radius: var(--radius-md);

  background: var(--color-surface);

  cursor: pointer;

  font-size: 14px;
  font-weight: 700;

  color: #334155;

  transition: 0.2s;
}

.day:hover {
  background: var(--color-primary-soft);
  border-color: #bfd3ff;
  color: var(--color-primary-strong);
}

/* TODAY */
.today {
  border-color: var(--color-primary);
  color: var(--color-primary-strong);
}

/* SELECTED */
.selected {
  background: var(--color-primary);
  border-color: var(--color-primary);

  color: white;

  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.2);
}

.selected:hover {
  background: var(--color-primary-strong);
  color: white;
}

/* PAST */
.past {
  color: var(--color-text-soft);
  cursor: not-allowed;
  opacity: 0.42;
}

.past:hover {
  background: var(--color-surface);
  border-color: transparent;
  color: var(--color-text-soft);
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
