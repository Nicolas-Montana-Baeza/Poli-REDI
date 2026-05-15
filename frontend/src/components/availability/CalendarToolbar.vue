<script setup>
import {
  ChevronLeft,
  ChevronRight,
  CalendarDays,
  Filter
} from 'lucide-vue-next'

const props = defineProps({
  currentDate: {
    type: String,
    default: 'Lunes 12 Mayo'
  }
})

const emit = defineEmits([
  'prev-day',
  'next-day',
  'today',
  'open-calendar'
])
</script>

<template>
  <section class="toolbar">

    <!-- LEFT -->
    <div class="left">

      <!-- DATE NAV -->
      <div class="date-navigation">

        <button
          class="icon-button"
          @click="$emit('prev-day')"
        >
          <ChevronLeft :size="18" />
        </button>

        <button
          class="date-button"
          @click="$emit('open-calendar')"
        >

          <CalendarDays :size="18" />

          <span>
            {{ currentDate }}
          </span>

        </button>

        <button
          class="icon-button"
          @click="$emit('next-day')"
        >
          <ChevronRight :size="18" />
        </button>

      </div>

    </div>

    <!-- RIGHT -->
    <div class="right">

      <!-- FILTER -->
      <button class="filter-button">

        <Filter :size="18" />

        <span>
          Filtros
        </span>

      </button>

      <!-- TODAY -->
      <button
        class="today-button"
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

/* ICON BUTTON */
.icon-button {
  width: 44px;
  height: 44px;

  border: none;

  border-radius: 14px;

  background: white;

  border: 1px solid #e2e8f0;

  display: flex;
  align-items: center;
  justify-content: center;

  cursor: pointer;

  transition: 0.2s;
}

.icon-button:hover {
  background: #eff6ff;

  border-color: #bfdbfe;
}

/* DATE BUTTON */
.date-button {
  height: 44px;

  border: none;

  border-radius: 14px;

  background: white;

  border: 1px solid #e2e8f0;

  padding: 0 18px;

  display: flex;
  align-items: center;
  gap: 10px;

  cursor: pointer;

  font-size: 14px;
  font-weight: 600;

  color: #0f172a;

  transition: 0.2s;
}

.date-button:hover {
  background: #eff6ff;

  border-color: #bfdbfe;
}

/* RIGHT */
.right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* FILTER */
.filter-button {
  height: 44px;

  border: none;

  border-radius: 14px;

  background: white;

  border: 1px solid #e2e8f0;

  padding: 0 16px;

  display: flex;
  align-items: center;
  gap: 10px;

  cursor: pointer;

  font-size: 14px;
  font-weight: 600;

  color: #334155;

  transition: 0.2s;
}

.filter-button:hover {
  background: #eff6ff;

  border-color: #bfdbfe;
}

/* TODAY */
.today-button {
  height: 44px;

  border: none;

  border-radius: 14px;

  padding: 0 18px;

  background: linear-gradient(
    135deg,
    #2563eb,
    #1d4ed8
  );

  color: white;

  font-size: 14px;
  font-weight: 700;

  cursor: pointer;

  transition: 0.2s;
}

.today-button:hover {
  transform: translateY(-1px);

  box-shadow:
    0 10px 20px rgba(37,99,235,0.25);
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

  .date-button {
    flex: 1;
    justify-content: center;
  }

  .filter-button,
  .today-button {
    flex: 1;
    justify-content: center;
  }
}
</style>