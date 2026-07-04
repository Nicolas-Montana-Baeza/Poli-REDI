<script setup>
defineProps({
  variant: {
    type: String,
    default: 'card'
  },

  items: {
    type: Number,
    default: 3
  }
})
</script>

<template>
  <div
    class="skeleton"
    :class="variant"
    aria-hidden="true"
  >

    <template v-if="variant === 'availability'">

      <div class="toolbar">
        <span class="line medium" />
        <span class="button" />
        <span class="button" />
        <span class="button" />
      </div>

      <div class="availability-layout">

        <div class="calendar">
          <span class="line short" />

          <div class="calendar-grid">
            <span
              v-for="day in 35"
              :key="day"
              class="dot"
            />
          </div>
        </div>

        <div class="schedule">
          <div
            v-for="row in 5"
            :key="row"
            class="schedule-row"
          >
            <span class="avatar" />

            <div class="schedule-lines">
              <span class="line wide" />
              <span class="line narrow" />
            </div>

            <span class="block" />
          </div>
        </div>

      </div>

    </template>

    <template v-else-if="variant === 'resources'">

      <div class="grid">
        <div
          v-for="item in items"
          :key="item"
          class="card-shell"
        >
          <span class="line wide" />
          <span class="line narrow" />

          <div class="chips">
            <span class="chip" />
            <span class="chip" />
            <span class="chip short-chip" />
          </div>
        </div>
      </div>

    </template>

    <template v-else-if="variant === 'reservations'">

      <div class="list">
        <div
          v-for="item in items"
          :key="item"
          class="card-shell"
        >
          <div class="row between">
            <div class="stack">
              <span class="chip short-chip" />
              <span class="line wide" />
              <span class="line narrow" />
            </div>

            <span class="button" />
          </div>

          <div class="chips">
            <span class="chip" />
            <span class="chip" />
            <span class="chip" />
          </div>
        </div>
      </div>

    </template>

    <template v-else-if="variant === 'dashboard'">

      <div class="grid dashboard-grid">
        <div
          v-for="item in items"
          :key="item"
          class="media-card"
        >
          <span class="media" />
          <span class="line wide" />
          <span class="line narrow" />
        </div>
      </div>

    </template>

    <template v-else>

      <div
        v-for="item in items"
        :key="item"
        class="card-shell"
      >
        <span class="line wide" />
        <span class="line medium" />
        <span class="line narrow" />
      </div>

    </template>

  </div>
</template>

<style scoped>
.skeleton {
  width: 100%;
}

.card-shell,
.calendar,
.schedule,
.toolbar,
.media-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.card-shell {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 14px;
}

.list {
  display: flex;
  flex-direction: column;

  gap: 14px;
}

.grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(240px, 1fr));

  gap: 18px;
}

.dashboard-grid {
  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));
}

.toolbar {
  min-height: 70px;

  padding: 16px;

  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 12px;

  margin-bottom: 24px;
}

.availability-layout {
  display: grid;

  grid-template-columns: 340px 1fr;

  gap: 24px;
}

.calendar {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 20px;
}

.calendar-grid {
  display: grid;

  grid-template-columns: repeat(7, 1fr);

  gap: 10px;
}

.schedule {
  padding: 18px;

  display: flex;
  flex-direction: column;

  gap: 14px;
}

.schedule-row {
  display: grid;

  grid-template-columns: 44px 1fr 28%;

  gap: 14px;
  align-items: center;

  min-height: 64px;
}

.schedule-lines,
.stack {
  display: flex;
  flex-direction: column;

  gap: 10px;
}

.row {
  display: flex;
  align-items: flex-start;

  gap: 16px;
}

.between {
  justify-content: space-between;
}

.chips {
  display: flex;
  flex-wrap: wrap;

  gap: 8px;
}

.media-card {
  padding: 16px;

  display: flex;
  flex-direction: column;

  gap: 12px;
}

.line,
.button,
.chip,
.dot,
.avatar,
.block,
.media {
  display: block;

  overflow: hidden;

  background:
    linear-gradient(
      90deg,
      #e2e8f0 0%,
      #f8fafc 45%,
      #e2e8f0 90%
    );

  background-size: 220% 100%;

  animation: shimmer 1.35s ease-in-out infinite;
}

.line {
  height: 14px;

  border-radius: 999px;
}

.wide {
  width: 78%;
}

.medium {
  width: 54%;
}

.narrow {
  width: 36%;
}

.short {
  width: 42%;
}

.button {
  width: 92px;
  height: 40px;

  border-radius: 14px;

  flex-shrink: 0;
}

.chip {
  width: 104px;
  height: 30px;

  border-radius: 999px;
}

.short-chip {
  width: 72px;
}

.dot {
  width: 100%;

  aspect-ratio: 1;

  border-radius: 12px;
}

.avatar {
  width: 44px;
  height: 44px;

  border-radius: 14px;
}

.block {
  width: 100%;
  height: 42px;

  border-radius: 14px;
}

.media {
  width: 100%;

  aspect-ratio: 16 / 9;

  border-radius: 14px;
}

@keyframes shimmer {
  0% {
    background-position: 120% 0;
  }

  100% {
    background-position: -120% 0;
  }
}

@media (max-width: 1024px) {
  .availability-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .button {
    width: 100%;
  }

  .schedule-row {
    grid-template-columns: 44px 1fr;
  }

  .block {
    grid-column: 1 / -1;
  }
}
</style>
