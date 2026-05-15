<script setup>
import { ref } from 'vue'

import TimeSlot from './TimeSlot.vue'

const props = defineProps({
  resource: {
    type: Object,
    required: true
  },

  slots: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits([
  'slot-selected'
])

const selectedSlot = ref(null)

/* SELECT */
const handleSelect = (time) => {

  selectedSlot.value = time

  emit('slot-selected', time)
}
</script>

<template>
  <div class="column">

    <!-- HEADER -->
    <div class="resource-header">

      <div class="resource-info">

        <h3>
          {{ resource.name }}
        </h3>

        <p>
          {{ resource.type }}
        </p>

      </div>

      <!-- STATUS -->
      <div
        class="resource-status"
        :class="resource.status"
      >
        {{ resource.status }}
      </div>

    </div>

    <!-- SLOTS -->
    <div class="slots">

      <TimeSlot
        v-for="slot in slots"
        :key="slot.time"

        :time="slot.time"

        :available="slot.available"

        :reserved="slot.reserved"

        :selected="
          selectedSlot === slot.time
        "

        @select="
          handleSelect
        "
      />

    </div>

  </div>
</template>

<style scoped>
.column {
  min-width: 280px;

  background: white;

  border-radius: 24px;

  padding: 18px;

  border: 1px solid #e2e8f0;

  display: flex;
  flex-direction: column;

  gap: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

/* HEADER */
.resource-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 12px;
}

/* INFO */
.resource-info h3 {
  margin: 0;

  font-size: 18px;
  font-weight: 700;

  color: #0f172a;
}

.resource-info p {
  margin-top: 4px;

  font-size: 13px;

  color: #64748b;
}

/* STATUS */
.resource-status {
  padding: 6px 12px;

  border-radius: 999px;

  font-size: 12px;
  font-weight: 700;

  text-transform: capitalize;
}

/* COLORS */
.available {
  background: #dcfce7;
  color: #15803d;
}

.busy {
  background: #fee2e2;
  color: #dc2626;
}

.maintenance {
  background: #fef3c7;
  color: #b45309;
}

/* SLOTS */
.slots {
  display: flex;
  flex-direction: column;

  gap: 12px;
}

/* MOBILE */
@media (max-width: 768px) {
  .column {
    min-width: 85vw;
  }
}
</style>