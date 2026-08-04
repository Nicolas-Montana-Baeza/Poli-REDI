<script setup>
import { computed } from 'vue'
import { CheckCircle2 } from 'lucide-vue-next'
import {
  getWorkshopEnrollmentLabel,
  isWorkshopAvailabilityItem
} from '@/utils/workshopEnrollment'

const props = defineProps({
  item: {
    type: Object,
    required: true
  },
  compact: {
    type: Boolean,
    default: false
  },
  ariaHidden: {
    type: Boolean,
    default: false
  }
})

const visible = computed(() => (
  isWorkshopAvailabilityItem(props.item) &&
  props.item?.workshop?.isEnrolled === true
))
const label = computed(() => getWorkshopEnrollmentLabel(props.item))
const title = computed(() => 'Ya estás inscrito en este taller')
</script>

<template>
  <span
    v-if="visible"
    class="workshop-enrollment-badge enrolled"
    :class="{ compact }"
    data-workshop-enrollment="enrolled"
    :title="title"
    :aria-label="ariaHidden ? undefined : title"
    :aria-hidden="ariaHidden ? 'true' : undefined"
  >
    <CheckCircle2 :size="compact ? 11 : 13" aria-hidden="true" />
    <span>{{ label }}</span>
  </span>
</template>

<style scoped>
.workshop-enrollment-badge {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 4px;
  min-height: 20px;
  padding: 2px 7px;
  border: 1px solid #cbd5e1;
  border-radius: var(--radius-pill);
  background: #f8fafc;
  color: #475569;
  font-size: 10px;
  font-weight: 850;
  line-height: 1.1;
  white-space: nowrap;
}

.workshop-enrollment-badge.enrolled {
  border-color: #86efac;
  background: #f0fdf4;
  color: #15803d;
}

.workshop-enrollment-badge.compact {
  min-height: 17px;
  gap: 3px;
  padding: 1px 5px;
  font-size: 9px;
}

.workshop-enrollment-badge svg {
  flex: 0 0 auto;
}
</style>
