<script setup>
import { computed } from 'vue'

const props = defineProps({
  time: {
    type: String,
    default: '18:00'
  },

  available: {
    type: Boolean,
    default: true
  },

  selected: {
    type: Boolean,
    default: false
  },

  reserved: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'select'
])

const slotClass = computed(() => {
  if (props.selected) return 'selected'

  if (props.reserved) return 'reserved'

  if (!props.available) return 'disabled'

  return 'available'
})

const handleClick = () => {
  if (!props.available || props.reserved) return

  emit('select', props.time)
}
</script>

<template>
  <button
    class="slot"
    :class="slotClass"
    @click="handleClick"
  >

    <!-- TIME -->
    <span class="time">
      {{ time }}
    </span>

    <!-- STATUS -->
    <span class="status">

      <template v-if="selected">
        Seleccionado
      </template>

      <template v-else-if="reserved">
        Reservado
      </template>

      <template v-else-if="!available">
        No disponible
      </template>

      <template v-else>
        Disponible
      </template>

    </span>

  </button>
</template>

<style scoped>
.slot {
  width: 100%;

  min-height: 72px;

  border: none;

  border-radius: 16px;

  padding: 14px;

  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;

  gap: 6px;

  cursor: pointer;

  transition: 0.2s ease;

  text-align: left;

  font-family: inherit;
}

/* Hover */
.slot:hover {
  transform: translateY(-2px);
}

/* TIME */
.time {
  font-size: 15px;
  font-weight: 700;
}

/* STATUS */
.status {
  font-size: 12px;
  font-weight: 500;
}

/* AVAILABLE */
.available {
  background: white;

  border: 1px solid #dbeafe;

  color: #1e3a8a;
}

.available .status {
  color: #64748b;
}

.available:hover {
  background: #eff6ff;
}

/* SELECTED */
.selected {
  background: linear-gradient(
    135deg,
    #2563eb,
    #1d4ed8
  );

  color: white;

  box-shadow:
    0 8px 20px rgba(37,99,235,0.25);
}

.selected .status {
  color: rgba(255,255,255,0.82);
}

/* RESERVED */
.reserved {
  background: #fee2e2;

  color: #b91c1c;

  cursor: not-allowed;

  border: 1px solid #fecaca;
}

.reserved .status {
  color: #dc2626;
}

/* DISABLED */
.disabled {
  background: #f1f5f9;

  color: #94a3b8;

  cursor: not-allowed;

  border: 1px solid #e2e8f0;
}

.disabled .status {
  color: #94a3b8;
}

/* Mobile */
@media (max-width: 768px) {
  .slot {
    min-height: 64px;

    padding: 12px;
  }

  .time {
    font-size: 14px;
  }
}
</style>