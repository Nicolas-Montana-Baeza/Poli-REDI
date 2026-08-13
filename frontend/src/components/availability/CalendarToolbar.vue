<script setup>
import {
  ChevronLeft,
  ChevronRight,
  CalendarDays
} from '@lucide/vue'

const props = defineProps({
  currentDate: {
    type: String,
    default: 'Lunes 12 Mayo'
  }
})

const emit = defineEmits([
  'prev-day',
  'next-day',
  'today'
])
</script>

<template>
  <section class="toolbar">

    <!-- LEFT -->
    <div class="left">

      <!-- DATE NAV -->
      <div class="date-navigation">

        <button
          class="toolbar-button icon-button"
          type="button"
          @click="$emit('prev-day')"
        >
          <ChevronLeft :size="18" />
        </button>

        <div class="date-label">

          <CalendarDays :size="18" />

          <span>
            {{ currentDate }}
          </span>

        </div>

        <button
          class="toolbar-button icon-button"
          type="button"
          @click="$emit('next-day')"
        >
          <ChevronRight :size="18" />
        </button>

      </div>

    </div>

    <!-- RIGHT -->
    <div class="right">

      <!-- TODAY -->
      <button
        class="today-button app-button primary"
        type="button"
        @click="$emit('today')"
      >
        Hoy
      </button>

    </div>

  </section>
</template>

<style scoped>
.toolbar {
  width: 100%;

  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 20px;

  flex-wrap: wrap;
}

/* LEFT */
.left {
  display: flex;
  align-items: center;
  gap: 14px;
}

/* DATE NAV */
.date-navigation {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-button {
  height: 42px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: #334155;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  font-weight: 700;
  transition: background-color 0.18s ease, border-color 0.18s ease, color 0.18s ease;
}

.toolbar-button:hover {
  background: var(--color-primary-soft);
  border-color: #bfd3ff;
  color: var(--color-primary-strong);
}

/* ICON BUTTON */
.icon-button {
  width: 42px;
}

/* DATE BUTTON */
.date-label {
  height: 42px;
  padding: 0 18px;
  gap: 10px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
}

/* RIGHT */
.right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* TODAY */
.today-button {
  padding: 0 22px;
}

/* MOBILE */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .left,
  .right {
    width: 100%;
  }

  .date-navigation {
    width: 100%;
  }

  .date-label {
    flex: 1;
    justify-content: center;
  }

  .today-button {
    flex: 1;
    justify-content: center;
  }
}
</style>
