<script setup>
import { computed } from 'vue'
import { getAvailabilityType } from '@/utils/availabilityType'

const props = defineProps({
  item: {
    type: Object,
    default: null
  },
  resource: {
    type: Object,
    default: null
  },
  meta: {
    type: Object,
    default: null
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

const typeMeta = computed(() =>
  props.meta || getAvailabilityType(props.item, props.resource)
)
</script>

<template>
  <span
    class="availability-type-chip"
    :class="[
      `tone-${typeMeta.tone}`,
      { compact }
    ]"
    :data-availability-type="typeMeta.key"
    :aria-hidden="ariaHidden ? 'true' : undefined"
  >
    {{ typeMeta.label }}
  </span>
</template>

<style scoped>
.availability-type-chip {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  min-height: 22px;
  max-width: 100%;
  padding: 3px 9px;
  border: 1px solid transparent;
  border-radius: var(--radius-pill);
  font-size: 11px;
  font-weight: 850;
  line-height: 1.15;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.availability-type-chip.compact {
  min-height: 18px;
  padding: 2px 6px;
  font-size: 9px;
}

.tone-reservation {
  border-color: #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

.tone-group {
  border-color: #c7d2fe;
  background: #eef2ff;
  color: #4338ca;
}

.tone-open-use {
  border-color: #99f6e4;
  background: #f0fdfa;
  color: #0f766e;
}

.tone-workshop {
  border-color: var(--color-status-workshop-border);
  background: var(--color-status-workshop-bg);
  color: var(--color-status-workshop-text);
}

.tone-institutional {
  border-color: #fed7aa;
  background: #fff7ed;
  color: #9a3412;
}

@media (max-width: 480px) {
  .availability-type-chip {
    min-height: 20px;
    padding-inline: 7px;
    font-size: 10px;
  }
}
</style>
