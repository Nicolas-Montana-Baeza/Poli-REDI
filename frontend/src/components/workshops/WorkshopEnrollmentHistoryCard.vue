<script setup>
import { computed } from 'vue'
import { CalendarCheck, Clock, MapPin, UserRound } from 'lucide-vue-next'

import StatusBadge from '@/components/ui/StatusBadge.vue'

const props = defineProps({
  enrollment: {
    type: Object,
    required: true
  },
  selectable: { type: Boolean, default: false }
})
const emit = defineEmits(['open-detail'])

const status = computed(() => {
  if (props.enrollment.status === 'CANCELLED') {
    return { value: 'CANCELLED', label: 'Inscripción cancelada' }
  }
  if (!props.enrollment.isActive) {
    return { value: 'INACTIVE', label: 'Taller no vigente' }
  }
  return { value: 'ACTIVE', label: 'Inscrito' }
})

const enrolledAt = computed(() => {
  const value = new Date(props.enrollment.enrolledAt)
  if (Number.isNaN(value.getTime())) {
    return 'Fecha no disponible'
  }
  return new Intl.DateTimeFormat('es-CL', {
    day: '2-digit',
    month: 'long',
    year: 'numeric'
  }).format(value)
})
</script>

<template>
  <component
    :is="selectable ? 'button' : 'article'"
    class="workshop-history-card"
    :class="{ interactive: selectable }"
    :type="selectable ? 'button' : undefined"
    :data-history-id="`workshop-enrollment-${enrollment.id}`"
    :aria-label="selectable ? `Ver detalle del taller ${enrollment.title}` : undefined"
    @click="selectable ? emit('open-detail', enrollment, $event) : null"
  >
    <div class="card-heading">
      <div>
        <span class="type-badge">Taller</span>
        <StatusBadge :status="status.value" :label="status.label" />
        <h2>{{ enrollment.title }}</h2>
        <p v-if="enrollment.description">{{ enrollment.description }}</p>
      </div>
    </div>

    <div class="workshop-meta">
      <span>
        <MapPin :size="17" />
        {{ enrollment.location || 'Lugar por confirmar' }}
      </span>
      <span>
        <Clock :size="17" />
        {{ enrollment.dayText }} · {{ enrollment.scheduleText }}
      </span>
      <span v-if="enrollment.instructorName">
        <UserRound :size="17" />
        {{ enrollment.instructorName }}
      </span>
      <span>
        <CalendarCheck :size="17" />
        Inscripción: {{ enrolledAt }}
      </span>
    </div>
  </component>
</template>

<style scoped>
.workshop-history-card {
  width: 100%;
  box-sizing: border-box;
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  color: inherit;
  font: inherit;
  text-align: left;
}
.workshop-history-card.interactive { cursor: pointer; transition: border-color .2s, box-shadow .2s; }
.workshop-history-card.interactive:hover,
.workshop-history-card.interactive:focus-visible {
  border-color: var(--color-primary, #2563eb);
  box-shadow: 0 0 0 3px var(--color-primary-soft, #dbeafe), var(--shadow-card);
  outline: none;
}

.card-heading > div {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.type-badge {
  padding: 5px 9px;
  border-radius: var(--radius-pill);
  background: #ede9fe;
  color: #6d28d9;
  font-size: 12px;
  font-weight: 800;
}

h2 {
  flex-basis: 100%;
  margin: 4px 0 0;
  color: var(--color-text);
  font-size: 18px;
}

p {
  flex-basis: 100%;
  margin: 0;
  color: var(--color-text-muted);
}

.workshop-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.workshop-meta span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: var(--radius-pill);
  background: var(--color-surface-muted);
  color: #475569;
  font-size: 13px;
  font-weight: 650;
}

@media (max-width: 520px) {
  .workshop-meta span {
    width: 100%;
    justify-content: center;
  }
}
</style>
