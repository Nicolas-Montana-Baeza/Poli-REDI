<script setup>
import {
  computed,
  onMounted,
  ref
} from 'vue'

import {
  CalendarDays,
  CheckCircle2,
  LogOut,
  MapPin,
  Search,
  Users
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useWorkshopsStore } from '@/stores/workshops'

const workshopsStore =
  useWorkshopsStore()

const search = ref('')

// ============================================================================
// CARGA
// ============================================================================

onMounted(async () => {
  workshopsStore.clearMessages()

  await workshopsStore.fetchWorkshops()
})

// ============================================================================
// CALENDARIO
// ============================================================================

const weekdayNames = {
  1: 'Lunes',
  2: 'Martes',
  3: 'Miércoles',
  4: 'Jueves',
  5: 'Viernes',
  6: 'Sábado',
  7: 'Domingo'
}

// Evitamos Date("YYYY-MM-DD") porque JavaScript interpreta esa forma como UTC
// y puede desplazar visualmente la fecha según la zona horaria del navegador.
const formatInstitutionalDate = (value) => {
  if (!value) {
    return ''
  }

  const match =
    /^(\d{4})-(\d{2})-(\d{2})$/.exec(
      value
    )

  if (!match) {
    return value
  }

  const [, year, month, day] = match

  const date = new Date(
    Number(year),
    Number(month) - 1,
    Number(day)
  )

  return new Intl.DateTimeFormat(
    'es-CL',
    {
      day: '2-digit',
      month: 'short',
      year: 'numeric'
    }
  ).format(date)
}

const scheduleLabel = (schedule) => {
  if (
    schedule.scheduleType === 'SINGLE'
  ) {
    return [
      formatInstitutionalDate(
        schedule.specificDate
      ),
      `${schedule.startTime}–${schedule.endTime}`
    ]
      .filter(Boolean)
      .join(' · ')
  }

  if (
    schedule.scheduleType === 'WEEKLY'
  ) {
    const weekday =
      weekdayNames[schedule.dayOfWeek] ||
      'Día semanal'

    const range =
      schedule.validFrom &&
      schedule.validTo
        ? `${formatInstitutionalDate(
            schedule.validFrom
          )} – ${formatInstitutionalDate(
            schedule.validTo
          )}`
        : ''

    return [
      weekday,
      `${schedule.startTime}–${schedule.endTime}`,
      range
    ]
      .filter(Boolean)
      .join(' · ')
  }

  return `${schedule.startTime}–${schedule.endTime}`
}

// ============================================================================
// CUPOS
// ============================================================================

const availableSlots = (workshop) => {
  if (
    workshop.availableSpots !==
      undefined &&
    workshop.availableSpots !== null
  ) {
    return Math.max(
      0,
      Number(workshop.availableSpots)
    )
  }

  // Fallback defensivo para respuestas antiguas durante desarrollo.
  return Math.max(
    0,
    Number(workshop.capacity || 0) -
      Number(
        workshop.enrollmentCount || 0
      )
  )
}

const workshopStatus = (workshop) => {
  if (workshop.isEnrolled) {
    return 'Inscrito'
  }

  if (availableSlots(workshop) === 0) {
    return 'Sin cupos'
  }

  return 'Disponible'
}

const canEnroll = (workshop) => {
  return (
    !workshop.isEnrolled &&
    availableSlots(workshop) > 0
  )
}

const isMutating = (workshop) => {
  return (
    workshopsStore.mutatingId ===
    workshop.id
  )
}

// ============================================================================
// FILTRO
// ============================================================================

const filteredWorkshops = computed(
  () => {
    const term =
      search.value
        .trim()
        .toLowerCase()

    if (!term) {
      return workshopsStore.workshops
    }

    return workshopsStore.workshops.filter(
      workshop => {
        const schedules =
          workshop.schedules
            ?.map(scheduleLabel)
            .join(' ') || ''

        return [
          workshop.title,
          workshop.description,
          workshop.unitName,
          workshop.resourceName,
          schedules
        ].some(value =>
          String(value || '')
            .toLowerCase()
            .includes(term)
        )
      }
    )
  }
)

// ============================================================================
// ACCIONES
// ============================================================================

const enroll = async (workshop) => {
  if (
    !canEnroll(workshop) ||
    isMutating(workshop)
  ) {
    return
  }

  try {
    await workshopsStore.enroll(
      workshop.id
    )
  } catch {
    // El store mantiene el mensaje que debe mostrar la interfaz.
  }
}

const leave = async (workshop) => {
  if (
    !workshop.isEnrolled ||
    isMutating(workshop)
  ) {
    return
  }

  try {
    await workshopsStore.leave(
      workshop.id
    )
  } catch {
    // El store mantiene el mensaje que debe mostrar la interfaz.
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
          Revisa la programación institucional,
          los cupos disponibles y administra tu inscripción.
        </p>
      </div>
    </header>

    <section class="toolbar">
      <div class="search-box">
        <Search :size="18" />

        <input
          v-model="search"
          type="search"
          placeholder="Buscar por taller, unidad, recurso o fecha"
        >
      </div>
    </section>

    <div
      v-if="workshopsStore.actionError"
      class="state-card error"
    >
      {{ workshopsStore.actionError }}
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
            <div class="unit-name">
              {{ workshop.unitName }}
            </div>

            <h2>
              {{ workshop.title }}
            </h2>

            <p>
              {{
                workshop.description ||
                'Taller deportivo institucional.'
              }}
            </p>
          </div>

          <span
            class="status-pill"
            :class="{
              enrolled: workshop.isEnrolled,
              full:
                !workshop.isEnrolled &&
                availableSlots(workshop) === 0
            }"
          >
            {{ workshopStatus(workshop) }}
          </span>
        </div>

        <div class="details">

          <div class="detail-item">
            <MapPin :size="18" />

            <span>
              {{
                workshop.resourceName ||
                'Recurso por confirmar'
              }}
            </span>
          </div>

          <div
            v-for="schedule in workshop.schedules || []"
            :key="schedule.id"
            class="detail-item"
          >
            <CalendarDays :size="18" />

            <span>
              {{ scheduleLabel(schedule) }}
            </span>
          </div>

          <div class="detail-item">
            <Users :size="18" />

            <span>
              {{ workshop.enrollmentCount }}
              /
              {{ workshop.capacity }}
              inscritos
            </span>
          </div>

          <div class="capacity-row">
            <strong>
              {{ availableSlots(workshop) }}
            </strong>

            <span>
              {{
                availableSlots(workshop) === 1
                  ? 'cupo disponible'
                  : 'cupos disponibles'
              }}
            </span>
          </div>

        </div>

        <button
          v-if="!workshop.isEnrolled"
          type="button"
          class="action-button enroll-button"
          :disabled="
            !canEnroll(workshop) ||
            isMutating(workshop)
          "
          @click="enroll(workshop)"
        >
          <CheckCircle2 :size="18" />

          <span>
            {{
              availableSlots(workshop) === 0
                ? 'Sin cupos'
                : isMutating(workshop)
                  ? 'Inscribiendo...'
                  : 'Inscribirme'
            }}
          </span>
        </button>

        <button
          v-else
          type="button"
          class="action-button leave-button"
          :disabled="isMutating(workshop)"
          @click="leave(workshop)"
        >
          <LogOut :size="18" />

          <span>
            {{
              isMutating(workshop)
                ? 'Cancelando inscripción...'
                : 'Cancelar inscripción'
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

  width: min(100%, 560px);

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

  grid-template-columns:
    repeat(
      auto-fill,
      minmax(310px, 1fr)
    );

  gap: 16px;
}

.workshop-card {
  display: flex;
  flex-direction: column;
  gap: 18px;

  padding: 20px;

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

.unit-name {
  margin-bottom: 5px;

  color: var(--color-primary);

  font-size: 12px;
  font-weight: 900;

  text-transform: uppercase;
  letter-spacing: 0.04em;
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

.status-pill.full {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;

  color: var(--color-text-muted);

  font-size: 14px;
  font-weight: 700;
}

.detail-item svg {
  flex: 0 0 auto;

  margin-top: 1px;

  color: var(--color-primary);
}

.capacity-row {
  display: flex;
  align-items: baseline;
  gap: 6px;

  margin-top: 4px;
  padding-top: 12px;

  border-top: 1px solid var(--color-border);

  color: var(--color-text-muted);

  font-size: 13px;
  font-weight: 700;
}

.capacity-row strong {
  color: var(--color-text);

  font-size: 22px;
  font-weight: 900;
}

.action-button {
  margin-top: auto;

  min-height: 44px;

  border-radius: var(--radius-md);

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  font: inherit;
  font-weight: 900;

  cursor: pointer;
}

.enroll-button {
  border: 1px solid var(--color-primary);

  background: var(--color-primary);
  color: var(--color-primary-contrast);
}

.leave-button {
  border: 1px solid var(--color-error-border);

  background: var(--color-surface);
  color: var(--color-error);
}

.action-button:disabled {
  border-color: var(--color-border);

  background: var(--color-surface-soft);
  color: var(--color-text-muted);

  cursor: not-allowed;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }

  .workshops-grid {
    grid-template-columns: 1fr;
  }
}
</style>
