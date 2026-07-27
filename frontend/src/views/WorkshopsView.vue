<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import {
  CalendarDays,
  CheckCircle2,
  MapPin,
  Search,
  Users
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useWorkshopsStore } from '@/stores/workshops'

const workshopsStore = useWorkshopsStore()
const search = ref('')
const actionErrorElement = ref(null)

onMounted(async () => {
  workshopsStore.clearMessages()
  await workshopsStore.fetchWorkshops()
})

const filteredWorkshops = computed(() => {
  const term =
    search.value.trim().toLowerCase()

  if (!term) {
    return workshopsStore.workshops
  }

  return workshopsStore.workshops.filter((workshop) => {
    return [
      workshop.title,
      workshop.dayText,
      workshop.scheduleText,
      workshop.location
    ].some(value =>
      String(value || '').toLowerCase().includes(term)
    )
  })
})

const availableSlots = (workshop) => {
  return Math.max(
    0,
    Number(workshop.capacity || 0) -
      Number(workshop.enrolledCount || 0)
  )
}

const canEnroll = (workshop) => {
  return (
    workshopsStore.enrollingId === null &&
    !workshop.isEnrolled &&
    availableSlots(workshop) > 0
  )
}

const enroll = async (workshop) => {
  if (!canEnroll(workshop)) {
    return
  }

  try {
    await workshopsStore.enroll(workshop.id)
  } catch {
    await nextTick()
    actionErrorElement.value?.focus()
  }
}
</script>

<template>
  <main class="workshops-view">

    <header class="page-header">
      <div>
        <h1>
          Talleres deportivos
        </h1>

        <p>
          Revisa la oferta disponible e inscríbete en los talleres abiertos.
        </p>
      </div>
    </header>

    <section class="toolbar">
      <div class="search-box">
        <Search :size="18" />

        <input
          v-model="search"
          type="search"
          placeholder="Buscar por taller, día o lugar"
        >
      </div>
    </section>

    <div
      v-if="workshopsStore.actionError"
      ref="actionErrorElement"
      class="state-card error"
      role="alert"
      tabindex="-1"
    >
      <strong>{{ workshopsStore.actionError.message }}</strong>
      <span v-if="workshopsStore.actionError.conflict">
        {{ workshopsStore.actionError.conflict.title }}:
        {{ workshopsStore.actionError.conflict.dayText }},
        {{ workshopsStore.actionError.conflict.scheduleText }}.
      </span>
    </div>

    <div
      v-if="workshopsStore.actionSuccess"
      class="state-card success"
    >
      {{ workshopsStore.actionSuccess }}
    </div>

    <div
      v-if="workshopsStore.loading"
      aria-label="Cargando talleres"
    >
      <SkeletonLoader
        variant="reservations"
        :items="4"
      />
    </div>

    <div
      v-else-if="workshopsStore.loadingError"
      class="state-card error"
    >
      {{ workshopsStore.loadingError }}
    </div>

    <section
      v-else-if="!filteredWorkshops.length"
      class="state-card"
    >
      No hay talleres disponibles con esos filtros.
    </section>

    <section
      v-else
      class="workshops-grid"
    >
      <article
        v-for="workshop in filteredWorkshops"
        :key="workshop.id"
        class="workshop-card"
      >
        <div class="card-top">
          <div>
            <h2>
              {{ workshop.title }}
            </h2>

            <p>
              {{ workshop.description || 'Taller deportivo institucional.' }}
            </p>
          </div>

          <span
            class="status-pill"
            :class="{ enrolled: workshop.isEnrolled }"
          >
            {{ workshop.isEnrolled ? 'Inscrito' : 'Disponible' }}
          </span>
        </div>

        <div class="details">
          <div class="detail-item">
            <CalendarDays :size="18" />

            <span>
              {{ workshop.dayText }} · {{ workshop.scheduleText }}
            </span>
          </div>

          <div class="detail-item">
            <MapPin :size="18" />

            <span>
              {{ workshop.location || 'Lugar por confirmar' }}
            </span>
          </div>

          <div class="detail-item">
            <Users :size="18" />

            <span>
              {{ workshop.enrolledCount }} / {{ workshop.capacity }} inscritos
            </span>
          </div>
        </div>

        <button
          type="button"
          class="enroll-button"
          :disabled="
            !canEnroll(workshop)
          "
          @click="enroll(workshop)"
        >
          <CheckCircle2 :size="18" />

          <span>
            {{
              workshop.isEnrolled
                ? 'Ya estás inscrito'
                : availableSlots(workshop) === 0
                  ? 'Sin cupos'
                  : workshopsStore.enrollingId === workshop.id
                    ? 'Inscribiendo...'
                    : 'Inscribirme'
            }}
          </span>
        </button>
      </article>
    </section>

  </main>
</template>

<style scoped>
.workshops-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-header h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 30px;
  font-weight: 900;
}

.page-header p {
  margin-top: 8px;
  color: var(--color-text-muted);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  width: min(100%, 480px);
  padding: 0 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  color: var(--color-text-muted);
  box-shadow: var(--shadow-card);
}

.search-box input {
  width: 100%;
  min-height: 44px;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--color-text);
  font: inherit;
}

.state-card {
  border-radius: var(--radius-lg);
}

.state-card.error {
  background: var(--color-error-soft);
  color: var(--color-error);
  border-color: var(--color-error-border);
}

.state-card.success {
  background: var(--color-success-soft);
  color: var(--color-success);
  border-color: var(--color-success-border);
}

.workshops-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
  gap: 16px;
}

.workshop-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  background: var(--color-surface);
  box-shadow: var(--shadow-card);
}

.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.card-top h2 {
  margin: 0;
  color: var(--color-text);
  font-size: 19px;
  font-weight: 900;
}

.card-top p {
  margin-top: 6px;
  color: var(--color-text-muted);
  font-size: 14px;
}

.status-pill {
  flex: 0 0 auto;
  padding: 6px 10px;
  border-radius: var(--radius-pill);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-size: 12px;
  font-weight: 900;
}

.status-pill.enrolled {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text-muted);
  font-size: 14px;
  font-weight: 700;
}

.detail-item svg {
  flex: 0 0 auto;
  color: var(--color-primary);
}

.enroll-button {
  margin-top: auto;
  min-height: 44px;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
  background: var(--color-primary);
  color: var(--color-primary-contrast);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 900;
  cursor: pointer;
}

.enroll-button:disabled {
  border-color: var(--color-border);
  background: var(--color-surface-soft);
  color: var(--color-text-muted);
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }
}
</style>
