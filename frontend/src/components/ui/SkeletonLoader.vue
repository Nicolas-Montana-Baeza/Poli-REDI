<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'list'
  },
  items: {
    type: Number,
    default: 3
  },
  columns: {
    type: Number,
    default: null
  },
  mobileCarousel: {
    type: Boolean,
    default: false
  }
})

const aliases = {
  availability: 'availability-timelines',
  dashboard: 'media-grid',
  reservations: 'list',
  resources: 'media-grid',
  card: 'compact-rows'
}

const resolvedVariant = computed(() => aliases[props.variant] || props.variant)
const itemCount = computed(() => Math.min(12, Math.max(1, Number(props.items) || 1)))
const rootStyle = computed(() => (
  props.columns
    ? { '--skeleton-columns': Math.max(1, Number(props.columns) || 1) }
    : undefined
))
</script>

<template>
  <div
    class="skeleton"
    :class="[
      `skeleton-${resolvedVariant}`,
      {
        'fixed-columns': columns,
        'mobile-carousel': mobileCarousel
      }
    ]"
    :style="rootStyle"
    aria-hidden="true"
  >
    <template v-if="resolvedVariant === 'availability-timelines'">
      <div class="availability-toolbar">
        <div class="date-navigation">
          <span class="placeholder icon-control" />
          <span class="placeholder date-control" />
          <span class="placeholder icon-control" />
        </div>
        <span class="placeholder today-control" />
      </div>

      <div class="availability-layout">
        <div class="calendar-card skeleton-surface">
          <div class="calendar-heading">
            <span class="placeholder line line-medium" />
            <div class="calendar-arrows">
              <span class="placeholder small-control" />
              <span class="placeholder small-control" />
            </div>
          </div>
          <div class="weekday-row">
            <span v-for="day in 7" :key="`weekday-${day}`" class="placeholder micro-line" />
          </div>
          <div class="calendar-grid">
            <span v-for="day in 35" :key="day" class="placeholder calendar-day" />
          </div>
        </div>

        <div class="timeline-area">
          <div class="view-switch-placeholder">
            <span class="placeholder switch-option" />
            <span class="placeholder switch-option" />
          </div>
          <span class="placeholder line line-short timeline-title" />
          <span class="placeholder line line-medium timeline-description" />
          <div class="timeline-strip">
            <article v-for="item in 3" :key="item" class="timeline-card skeleton-surface">
              <div class="timeline-card-heading">
                <div class="stack">
                  <span class="placeholder line line-wide" />
                  <span class="placeholder line line-short" />
                </div>
                <span class="placeholder badge" />
              </div>
              <span class="placeholder mode-line" />
              <div class="timeline-body">
                <span v-for="hour in 7" :key="hour" class="placeholder hour-line" />
              </div>
            </article>
          </div>
        </div>
      </div>
    </template>

    <template v-else-if="resolvedVariant === 'media-grid'">
      <div class="media-grid">
        <article v-for="item in itemCount" :key="item" class="media-card skeleton-surface">
          <span class="placeholder media" />
          <div class="media-copy stack">
            <span class="placeholder line line-wide" />
            <span class="placeholder line line-short" />
            <span class="placeholder badge" />
          </div>
        </article>
      </div>
    </template>

    <template v-else-if="resolvedVariant === 'card-grid'">
      <div class="card-grid">
        <article v-for="item in itemCount" :key="item" class="content-card skeleton-surface">
          <div class="row between">
            <div class="stack grow">
              <span class="placeholder line line-wide" />
              <span class="placeholder line line-medium" />
            </div>
            <span class="placeholder badge" />
          </div>
          <div class="stack details-lines">
            <span class="placeholder line line-medium" />
            <span class="placeholder line line-wide" />
            <span class="placeholder line line-short" />
          </div>
          <span class="placeholder neutral-footer" />
        </article>
      </div>
    </template>

    <template v-else-if="resolvedVariant === 'detail'">
      <article class="detail-panel skeleton-surface">
        <div class="detail-heading row between">
          <div class="stack grow">
            <span class="placeholder badge" />
            <span class="placeholder line line-medium detail-title" />
            <span class="placeholder line line-short" />
          </div>
        </div>
        <div class="facts-grid">
          <div v-for="item in 3" :key="item" class="fact-card">
            <span class="placeholder avatar" />
            <div class="stack grow">
              <span class="placeholder line line-short" />
              <span class="placeholder line line-wide" />
            </div>
          </div>
        </div>
        <div class="detail-section">
          <span class="placeholder line line-medium" />
          <span class="placeholder progress-track" />
          <div class="chips">
            <span class="placeholder chip" />
            <span class="placeholder chip" />
          </div>
        </div>
      </article>
    </template>

    <template v-else-if="resolvedVariant === 'metrics-table'">
      <div class="metrics-grid">
        <article v-for="item in Math.min(itemCount, 4)" :key="`metric-${item}`" class="metric-card skeleton-surface">
          <span class="placeholder avatar" />
          <span class="placeholder line line-medium" />
          <span class="placeholder metric-value" />
        </article>
      </div>
      <div class="panels-grid">
        <article v-for="panel in 2" :key="`panel-${panel}`" class="data-panel skeleton-surface">
          <span class="placeholder line line-medium" />
          <span class="placeholder line line-short" />
          <div class="compact-list">
            <div v-for="row in 3" :key="row" class="compact-row">
              <span class="placeholder avatar" />
              <div class="stack grow">
                <span class="placeholder line line-wide" />
                <span class="placeholder line line-short" />
              </div>
            </div>
          </div>
        </article>
      </div>
    </template>

    <template v-else-if="resolvedVariant === 'compact-rows'">
      <div class="compact-list">
        <div v-for="item in itemCount" :key="item" class="compact-row">
          <span class="placeholder avatar" />
          <div class="stack grow">
            <span class="placeholder line line-wide" />
            <span class="placeholder line line-medium" />
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="list">
        <article v-for="item in itemCount" :key="item" class="list-card skeleton-surface">
          <div class="stack">
            <span class="placeholder badge" />
            <span class="placeholder line line-medium" />
            <span class="placeholder line line-short" />
          </div>
          <div class="chips">
            <span class="placeholder chip" />
            <span class="placeholder chip" />
            <span class="placeholder chip" />
          </div>
        </article>
      </div>
    </template>
  </div>
</template>

<style scoped>
.skeleton {
  width: 100%;
  min-width: 0;
}

.skeleton-surface {
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  box-shadow: var(--shadow-card);
}

.placeholder {
  display: block;
  overflow: hidden;
  background: linear-gradient(
    90deg,
    var(--color-border-soft) 0%,
    var(--color-surface-muted) 45%,
    var(--color-border-soft) 90%
  );
  background-size: 220% 100%;
  animation: skeleton-shimmer 1.35s ease-in-out infinite;
}

.stack {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: var(--space-2);
}

.grow {
  flex: 1;
}

.row {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.between {
  justify-content: space-between;
}

.line {
  height: 13px;
  border-radius: var(--radius-pill);
}

.line-wide { width: 78%; }
.line-medium { width: 56%; }
.line-short { width: 36%; }
.micro-line { width: 18px; height: 8px; border-radius: var(--radius-pill); }

.badge,
.chip {
  width: 84px;
  height: 26px;
  border-radius: var(--radius-pill);
}

.chip {
  width: 112px;
  height: 30px;
}

.avatar {
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  border-radius: var(--radius-md);
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.list,
.compact-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.list-card {
  display: flex;
  min-height: 132px;
  flex-direction: column;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
}

.compact-row {
  display: flex;
  min-height: 66px;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border-soft);
}

.compact-row:last-child {
  border-bottom: 0;
}

.media-grid,
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: var(--space-4);
}

.fixed-columns .media-grid {
  grid-template-columns: repeat(var(--skeleton-columns), minmax(0, 1fr));
}

.media-card {
  min-width: 0;
}

.media {
  width: 100%;
  aspect-ratio: 16 / 9;
}

.media-copy {
  padding: var(--space-4);
}

.content-card {
  display: flex;
  min-height: 260px;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-5);
}

.details-lines {
  flex: 1;
  justify-content: center;
}

.neutral-footer {
  width: 100%;
  height: 42px;
  border-radius: var(--radius-md);
}

.availability-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-5);
  margin-bottom: var(--space-6);
}

.date-navigation {
  display: flex;
  align-items: center;
  gap: 10px;
}

.icon-control,
.today-control,
.small-control {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
}

.date-control {
  width: 210px;
  height: 42px;
  border-radius: var(--radius-md);
}

.today-control {
  width: 76px;
}

.small-control {
  width: 34px;
  height: 34px;
}

.availability-layout {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  align-items: start;
  gap: var(--space-6);
}

.calendar-card {
  padding: var(--space-5);
}

.calendar-heading,
.calendar-arrows,
.weekday-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.weekday-row {
  margin: var(--space-5) 0 var(--space-3);
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: var(--space-2);
}

.calendar-day {
  width: 100%;
  aspect-ratio: 1;
  border-radius: var(--radius-md);
}

.timeline-area {
  min-width: 0;
  overflow: hidden;
}

.view-switch-placeholder {
  display: inline-flex;
  gap: 4px;
  margin-bottom: var(--space-4);
  padding: 4px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
}

.switch-option {
  width: 105px;
  height: 34px;
  border-radius: var(--radius-sm);
}

.timeline-title {
  height: 20px;
}

.timeline-description {
  margin: var(--space-2) 0 var(--space-5);
}

.timeline-strip {
  display: flex;
  gap: var(--space-4);
  overflow: hidden;
}

.timeline-card {
  min-width: 250px;
  flex: 1 0 250px;
}

.timeline-card-heading {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: 1px solid var(--color-border-soft);
}

.mode-line {
  height: 32px;
  margin: var(--space-3);
  border-radius: var(--radius-md);
}

.timeline-body {
  display: flex;
  min-height: 440px;
  flex-direction: column;
  justify-content: space-around;
  padding: var(--space-3);
  background: var(--color-surface-muted);
}

.hour-line {
  width: 100%;
  height: 1px;
}

.detail-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-5);
}

.detail-title {
  height: 24px;
}

.facts-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.fact-card {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface-muted);
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.progress-track {
  width: 100%;
  height: 10px;
  border-radius: var(--radius-pill);
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.metric-card {
  display: grid;
  min-height: 132px;
  gap: var(--space-3);
  padding: var(--space-5);
}

.metric-value {
  width: 74px;
  height: 28px;
  border-radius: var(--radius-sm);
}

.panels-grid {
  display: grid;
  grid-template-columns: minmax(240px, 340px) minmax(0, 1fr);
  gap: var(--space-4);
}

.data-panel {
  min-width: 0;
  padding: var(--space-5);
}

@keyframes skeleton-shimmer {
  from { background-position: 120% 0; }
  to { background-position: -120% 0; }
}

@media (max-width: 1024px) {
  .availability-layout,
  .panels-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .fixed-columns .media-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .availability-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .date-navigation,
  .today-control {
    width: 100%;
  }

  .date-control {
    min-width: 0;
    flex: 1;
  }

  .facts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .media-grid,
  .fixed-columns .media-grid,
  .card-grid {
    grid-template-columns: 1fr;
  }

  .mobile-carousel .media-grid {
    display: flex;
    gap: var(--space-3);
    overflow: hidden;
  }

  .mobile-carousel .media-card {
    width: min(82vw, 280px);
    flex: 0 0 min(82vw, 280px);
  }

  .detail-panel {
    padding: var(--space-4);
  }
}

@media (prefers-reduced-motion: reduce) {
  .placeholder {
    animation: none;
    background: var(--color-border-soft);
  }
}
</style>
