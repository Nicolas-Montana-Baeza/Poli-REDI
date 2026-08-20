<script setup>
import { computed, ref } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import FacilityCard from './FacilityCard.vue'

const props = defineProps({
  facilities: { type: Array, default: () => [] }
})

const carousel = ref(null)
const canMove = computed(() => props.facilities.length > 1)

const move = (direction) => {
  carousel.value?.scrollBy({
    left: direction * 320,
    behavior: 'smooth'
  })
}
</script>

<template>
  <div class="facility-carousel">
    <div class="carousel-header">
      <h2>Instalaciones disponibles</h2>

      <div v-if="canMove" class="carousel-controls">
        <button type="button" @click="move(-1)" aria-label="Anterior">
          <ChevronLeft :size="18" />
        </button>

        <button type="button" @click="move(1)" aria-label="Siguiente">
          <ChevronRight :size="18" />
        </button>
      </div>
    </div>

    <div ref="carousel" class="carousel-window">
      <div class="carousel-track">
        <div
          v-for="facility in facilities"
          :key="facility.id"
          class="carousel-item"
        >
          <FacilityCard
            :resource-id="facility.id"
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
  width: 100%;
  min-width: 0;
}

.carousel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.carousel-header h2 {
  margin: 0;
  font-size: 22px;
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
}

.carousel-window {
  overflow-x: auto;
  scroll-behavior: smooth;
  scroll-snap-type: x proximity;
  padding-bottom: 12px;
}

.carousel-track {
  display: flex;
  width: max-content;
  gap: 18px;
}

.carousel-item {
  width: clamp(240px, 26vw, 320px);
  flex: 0 0 auto;
  scroll-snap-align: start;
}

@media (max-width: 768px) {
  .carousel-item {
    width: min(78vw, 330px);
  }
}
</style>
