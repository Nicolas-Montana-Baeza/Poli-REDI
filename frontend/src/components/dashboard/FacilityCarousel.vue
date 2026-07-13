<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import FacilityCard from './FacilityCard.vue'

const props = defineProps({
  facilities: {
    type: Array,
    default: () => []
  }
})

const currentIndex = ref(0)
const isPaused = ref(false)
const transitionEnabled = ref(true)
let autoTimer = null

const loopedFacilities = computed(() => {
  if (props.facilities.length <= 1) {
    return props.facilities
  }

  return [
    ...props.facilities,
    ...props.facilities
  ]
})

const canMove = computed(() => {
  return props.facilities.length > 1
})

const trackStyle = computed(() => {
  return {
    transform: `translateX(calc(${-currentIndex.value} * (var(--carousel-card-width) + var(--carousel-gap))))`,
    transition: transitionEnabled.value
      ? 'transform 520ms cubic-bezier(0.22, 1, 0.36, 1)'
      : 'none'
  }
})

const moveNext = () => {
  if (!canMove.value) {
    return
  }

  transitionEnabled.value = true
  currentIndex.value += 1
}

const movePrevious = () => {
  if (!canMove.value) {
    return
  }

  if (currentIndex.value === 0) {
    transitionEnabled.value = false
    currentIndex.value = props.facilities.length

    window.requestAnimationFrame(() => {
      window.requestAnimationFrame(() => {
        transitionEnabled.value = true
        currentIndex.value -= 1
      })
    })

    return
  }

  transitionEnabled.value = true
  currentIndex.value -= 1
}

const handleTransitionEnd = () => {
  if (currentIndex.value < props.facilities.length) {
    return
  }

  transitionEnabled.value = false
  currentIndex.value = 0

  window.requestAnimationFrame(() => {
    transitionEnabled.value = true
  })
}

const advance = () => {
  if (!isPaused.value) {
    moveNext()
  }
}

onMounted(() => {
  autoTimer = window.setInterval(advance, 3200)
})

onBeforeUnmount(() => {
  if (autoTimer) {
    window.clearInterval(autoTimer)
  }
})
</script>

<template>
  <div class="facility-carousel">
    <div class="carousel-header">
      <h2>
        Instalaciones disponibles
      </h2>

      <div
        v-if="canMove"
        class="carousel-controls"
      >
        <button
          type="button"
          aria-label="Ver instalaciones anteriores"
          @click="movePrevious"
        >
          <ChevronLeft :size="18" />
        </button>

        <button
          type="button"
          aria-label="Ver más instalaciones"
          @click="moveNext"
        >
          <ChevronRight :size="18" />
        </button>
      </div>
    </div>

    <div
      class="carousel-window"
      tabindex="0"
      aria-label="Carrusel de instalaciones"
      @mouseenter="isPaused = true"
      @mouseleave="isPaused = false"
      @focusin="isPaused = true"
      @focusout="isPaused = false"
    >
      <div
        class="carousel-track"
        :style="trackStyle"
        @transitionend="handleTransitionEnd"
      >
        <div
          v-for="(facility, index) in loopedFacilities"
          :key="`${facility.id}-${index}`"
          class="carousel-item"
        >
          <FacilityCard
            :name="facility.name"
            :type="facility.type"
            :status="facility.status"
            :image="facility.image"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.facility-carousel {
  --carousel-card-width: clamp(260px, 24vw, 315px);
  --carousel-gap: 16px;
  width: 100%;
  min-width: 0;
}

.carousel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.carousel-header h2 {
  margin: 0;
  color: var(--color-text);
  font-size: 22px;
  font-weight: 800;
}

.carousel-controls {
  display: flex;
  gap: 8px;
}

.carousel-controls button {
  width: 38px;
  height: 38px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-primary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.carousel-controls button:hover {
  background: var(--color-primary-soft);
  border-color: #bfd3ff;
}

.carousel-window {
  width: 100%;
  overflow: hidden;
  padding: 2px 2px 14px;
}

.carousel-track {
  display: flex;
  gap: var(--carousel-gap);
  will-change: transform;
}

.carousel-item {
  flex: 0 0 var(--carousel-card-width);
  min-width: var(--carousel-card-width);
}

@media (max-width: 768px) {
  .facility-carousel {
    --carousel-card-width: min(82vw, 330px);
  }
}
</style>
