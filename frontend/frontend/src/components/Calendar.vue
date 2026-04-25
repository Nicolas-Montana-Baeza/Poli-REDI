<template>
  <div class="calendar">
    <header class="calendar-header">
      <button type="button" @click="prevMonth">‹</button>
      <div>
        <strong>{{ monthName }}</strong>
        <span>{{ displayedYear }}</span>
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
        }"
      >
        {{ dateObj.date }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Calendar',
  data() {
    const today = new Date();
    return {
      currentYear: today.getFullYear(),
      currentMonth: today.getMonth(),
      today,
      weekdays: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
    };
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
        'January',
        'February',
        'March',
        'April',
        'May',
        'June',
        'July',
        'August',
        'September',
        'October',
        'November',
        'December',
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
}

.calendar-day--muted {
  color: #9ca3af;
}

.calendar-day--today {
  background: #2563eb;
  color: white;
}
</style>
