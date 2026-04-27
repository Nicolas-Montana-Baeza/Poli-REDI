
<template>
  <div class="calendar">
    <header class="calendar-header">
      <button type="button" @click="prevMonth">‹</button>
      <div class="calendar-title">
        <div class="calendar-month">{{ monthName }}</div>
        <div class="calendar-year">{{ displayedYear }}</div>
      </div>
      <button type="button" @click="nextMonth">›</button>
    </header>

    <div class="calendar-grid">
      <div class="calendar-weekday" v-for="day in weekdays" :key="day">
        {{ day }}
      </div>
      <div
        class="calendar-day"
        v-for="dateObj in calendarDays"
        :key="dateObj.key"
        :class="{
          'calendar-day--muted': !dateObj.currentMonth,
          'calendar-day--today': dateObj.isToday,
          'calendar-day--disabled': !isDateSelectable(dateObj),
          'calendar-day--selectable': isDateSelectable(dateObj) && dateObj.currentMonth,
          'calendar-day--full': isDateFull(dateObj),
          'calendar-day--selected': isDateSelected(dateObj),
          'calendar-day--locked': isDateLocked(dateObj),
        }"
        @click="selectDate(dateObj)"
      >
        {{ dateObj.date }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Calendar',
  props: {
    fullyBookedDates: {
      type: Array,
      default: () => [],
    },
    selectedDate: {
      type: Object,
      default: null,
    },
  },
  data() {
    const today = new Date();
    return {
      currentYear: today.getFullYear(),
      currentMonth: today.getMonth(),
      today,
      weekdays: ['Dom', 'Lun', 'Mar', 'Mie', 'Jue', 'Vie', 'Sab'],
    };
  },
  mounted() {
    // Always start from today when component mounts
    const today = new Date();
    this.currentYear = today.getFullYear();
    this.currentMonth = today.getMonth();
    this.today = today;
    
    // Set current day as default selection
    if (!this.selectedDate) {
      this.$emit('date-selected', {
        date: today.getDate(),
        month: today.getMonth(),
        year: today.getFullYear(),
      });
    }
  },
  computed: {
    displayedYear() {
      return this.currentYear;
    },
    monthName() {
      return this.today.toLocaleString('default', {
        month: 'long',
      }).replace(this.today.toLocaleDateString('default', { month: 'long' }), this.months[this.currentMonth]);
    },
    months() {
      return [
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
        'Diciembre',
      ];
    },
    calendarDays() {
      const firstDayOfMonth = new Date(this.currentYear, this.currentMonth, 1);
      const startDay = firstDayOfMonth.getDay();
      const daysInMonth = new Date(this.currentYear, this.currentMonth + 1, 0).getDate();

      const previousMonthDays = new Date(this.currentYear, this.currentMonth, 0).getDate();
      const totalCells = 42; // 6 weeks
      const days = [];

      for (let i = 0; i < startDay; i++) {
        days.push({
          key: `prev-${i}`,
          date: previousMonthDays - startDay + 1 + i,
          currentMonth: false,
          isToday: false,
        });
      }

      for (let day = 1; day <= daysInMonth; day++) {
        const isToday =
          day === this.today.getDate() &&
          this.currentMonth === this.today.getMonth() &&
          this.currentYear === this.today.getFullYear();
        days.push({
          key: `current-${day}`,
          date: day,
          currentMonth: true,
          isToday,
        });
      }

      let nextDay = 1;
      while (days.length < totalCells) {
        days.push({
          key: `next-${nextDay}`,
          date: nextDay,
          currentMonth: false,
          isToday: false,
        });
        nextDay += 1;
      }

      return days;
    },
    selectableDateRange() {
      const startDate = new Date(this.today);
      startDate.setHours(0, 0, 0, 0);
      const endDate = new Date(this.today);
      endDate.setHours(0, 0, 0, 0);
      endDate.setDate(endDate.getDate() + 30);
      return { startDate, endDate };
    },
  },
  methods: {
    prevMonth() {
      if (this.currentMonth === 0) {
        this.currentMonth = 11;
        this.currentYear -= 1;
      } else {
        this.currentMonth -= 1;
      }
    },
    nextMonth() {
      if (this.currentMonth === 11) {
        this.currentMonth = 0;
        this.currentYear += 1;
      } else {
        this.currentMonth += 1;
      }
    },
    isDateSelectable(dateObj) {
      if (!dateObj.currentMonth) return false;
      const dateToCheck = new Date(this.currentYear, this.currentMonth, dateObj.date);
      return dateToCheck >= this.selectableDateRange.startDate && dateToCheck <= this.selectableDateRange.endDate;
    },
    isDateFull(dateObj) {
      if (!dateObj.currentMonth) return false;
      const dateString = `${this.currentYear}-${String(this.currentMonth + 1).padStart(2, '0')}-${String(dateObj.date).padStart(2, '0')}`;
      return this.fullyBookedDates.includes(dateString);
    },
    isDateSelected(dateObj) {
      if (!dateObj.currentMonth || !this.selectedDate) return false;
      return (
        this.selectedDate.date === dateObj.date &&
        this.selectedDate.month === this.currentMonth &&
        this.selectedDate.year === this.currentYear
      );
    },
    isDateLocked(dateObj) {
      return this.selectedDate && !this.isDateSelected(dateObj);
    },
    selectDate(dateObj) {
      if (this.isDateSelectable(dateObj)) {
        this.$emit('date-selected', {
          date: dateObj.date,
          month: this.currentMonth,
          year: this.currentYear,
        });
      }
    },
  },
};
</script>

<style scoped>
.calendar {
  max-width: 320px;
  border: 1px solid #d1d5db;
  border-radius: 12px;
  padding: 16px;
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.calendar-header button {
  border: none;
  background: transparent;
  font-size: 1.5rem;
  cursor: pointer;
  color: #111827;
  opacity: 1;
}

.calendar-title {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.calendar-month {
  font-size: 1.1rem;
  font-weight: 600;
  color: #111827;
}

.calendar-year {
  font-size: 0.9rem;
  color: #6b7280;
  margin-top: 2px;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 6px;
}

.calendar-weekday,
.calendar-day {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 40px;
}

.calendar-weekday {
  font-size: 0.8rem;
  font-weight: 600;
  color: #6b7280;
}

.calendar-day {
  border-radius: 8px;
  font-size: 0.95rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.calendar-day:hover:not(.calendar-day--muted) {
  background: #e5e7eb;
}

.calendar-day--muted {
  color: #9ca3af;
  cursor: not-allowed;
}

.calendar-day--disabled {
  color: #d1d5db;
  cursor: not-allowed;
}

.calendar-day--selectable {
  cursor: pointer;
}

.calendar-day--selectable:hover {
  background: #dbeafe;
}

.calendar-day--today {
  background: #10b981;
  color: white;
}

.calendar-day--full {
  background: #ef4444;
  color: white;
  cursor: pointer;
}

.calendar-day--full:hover {
  background: #dc2626;
}

.calendar-day--selected {
  background: transparent !important;
  color: rgba(255, 255, 255, 0) !important;
  border: 2px solid #111827;
  font-weight: 600;
}

.calendar-day--locked {
  opacity: 0.5;
  cursor: not-allowed;
}

.calendar-day--locked:hover {
  background: inherit;
}
</style>
