<script setup>
import { computed, ref } from 'vue'
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
  if (!canMove.value) {
    return {}
  }

  const duration = Math.max(props.facilities.length * 5, 26)

  return {
    '--carousel-duration': `${duration}s`,
    '--carousel-offset': String(currentIndex.value),
    '--carousel-items': String(props.facilities.length)
  }
})

const moveNext = () => {
  if (!canMove.value) {
    return
  }

  currentIndex.value =
    (currentIndex.value + 1) % props.facilities.length
}

const movePrevious = () => {
  if (!canMove.value) {
    return
  }

  currentIndex.value =
    (currentIndex.value - 1 + props.facilities.length) %
    props.facilities.length
}
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
        :class="{ paused: isPaused || !canMove }"
        :style="trackStyle"
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
  --carousel-card-width: clamp(240px, 26vw, 320px);
  --carousel-gap: 18px;
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
  padding: 2px 2px 16px;
  mask-image: linear-gradient(
    90deg,
    transparent 0,
    #000 32px,
    #000 calc(100% - 32px),
    transparent 100%
  );
}

.carousel-track {
  display: flex;
  gap: var(--carousel-gap);
  width: max-content;
  transform:
    translateX(calc(
      var(--carousel-offset, 0) *
      -1 *
      (var(--carousel-card-width) + var(--carousel-gap))
    ));
  animation: carousel-marquee var(--carousel-duration, 30s) linear infinite;
  will-change: transform;
}

.carousel-track.paused {
  animation-play-state: paused;
}

.carousel-item {
  flex: 0 0 var(--carousel-card-width);
  min-width: var(--carousel-card-width);
}

@keyframes carousel-marquee {
  from {
    transform:
      translateX(calc(
        var(--carousel-offset, 0) *
        -1 *
        (var(--carousel-card-width) + var(--carousel-gap))
      ));
  }

  to {
    transform:
      translateX(calc(
        -1 *
        var(--carousel-items) *
        (var(--carousel-card-width) + var(--carousel-gap)) -
        (
          var(--carousel-offset, 0) *
          (var(--carousel-card-width) + var(--carousel-gap))
        )
      ));
  }
}

@media (max-width: 768px) {
  .facility-carousel {
    --carousel-card-width: min(78vw, 330px);
    --carousel-gap: 14px;
  }

  .carousel-header {
    align-items: flex-start;
  }

  .carousel-header h2 {
    font-size: 20px;
  }

  .carousel-window {
    mask-image: none;
  }
}
</style>
