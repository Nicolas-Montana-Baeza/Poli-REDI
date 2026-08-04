<script setup>
import { computed } from 'vue'
import AvailabilityTypeChip from './AvailabilityTypeChip.vue'
import {
  AVAILABILITY_TYPE_ORDER,
  getAvailabilityType
} from '@/utils/availabilityType'
import { getReservationDateKey } from '@/utils/reservationTime'

const props = defineProps({
  resources: {
    type: Array,
    default: () => []
  },
  reservations: {
    type: Array,
    default: () => []
  },
  selectedDate: {
    type: String,
    default: ''
  }
})

const visibleTypes = computed(() => {
  const resourcesById = new Map(
    props.resources.map(resource => [
      String(resource.id),
      resource
    ])
  )
  const typesByKey = new Map()

  props.resources
    .filter(resource => resource.reservationMode === 'OPEN_USE')
    .forEach((resource) => {
      const meta = getAvailabilityType(null, resource)
      typesByKey.set(meta.key, meta)
    })

  props.reservations
    .filter((reservation) => {
      const isVisibleStatus = reservation.status !== 'CANCELLED'
      const isVisibleDate = !props.selectedDate ||
        getReservationDateKey(reservation.startTime) === props.selectedDate

      return isVisibleStatus && isVisibleDate
    })
    .forEach((reservation) => {
      const resource = resourcesById.get(String(reservation.resourceId))
      const meta = getAvailabilityType(reservation, resource)

      typesByKey.set(meta.key, meta)
    })

  return AVAILABILITY_TYPE_ORDER
    .filter(key => typesByKey.has(key))
    .map(key => typesByKey.get(key))
})
</script>

<template>
  <div
    v-if="visibleTypes.length"
    class="availability-type-legend"
    aria-label="Leyenda de tipos de bloque"
  >
    <span class="legend-title">
      Tipos de bloque
    </span>

    <div
      class="legend-list"
      role="list"
    >
      <span
        v-for="typeMeta in visibleTypes"
        :key="typeMeta.key"
        class="legend-item"
        role="listitem"
      >
        <AvailabilityTypeChip :meta="typeMeta" />
      </span>
    </div>
  </div>
</template>

<style scoped>
.availability-type-legend {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.legend-title {
  flex: 0 0 auto;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 800;
  white-space: nowrap;
}

.legend-list {
  display: flex;
  min-width: 0;
  gap: 6px;
  overflow-x: auto;
  padding: 2px;
  scrollbar-width: thin;
  white-space: nowrap;
}

.legend-item {
  display: inline-flex;
}

@media (max-width: 768px) {
  .availability-type-legend {
    width: 100%;
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }

  .legend-list {
    width: 100%;
    padding-bottom: 5px;
  }
}
</style>
