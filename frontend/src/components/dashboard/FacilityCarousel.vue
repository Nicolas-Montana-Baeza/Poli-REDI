<script setup>
import FacilityCard from './FacilityCard.vue'

defineProps({
  facilities: {
    type: Array,
    default: () => []
  }
})
</script>

<template>
  <div
    class="facilities-grid"
    aria-live="polite"
    aria-label="Instalaciones disponibles"
    tabindex="0"
  >
    <FacilityCard
      v-for="facility in facilities"
      :key="facility.id"
      :name="facility.name"
      :type="facility.type"
      :status="facility.status"
      :image="facility.image"
    />
  </div>
</template>

<style scoped>
.facilities-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 900px) {
  .facilities-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 600px) {
  .facilities-grid {
    width: 100%;
    max-width: 100%;
    display: flex;
    gap: 12px;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 2px 2px 12px;
    scroll-padding-inline: 2px;
    scroll-snap-type: x mandatory;
    overscroll-behavior-inline: contain;
    scrollbar-width: thin;
    -webkit-overflow-scrolling: touch;
  }

  .facilities-grid > * {
    flex: 0 0 min(82vw, 280px);
    width: min(82vw, 280px);
    scroll-snap-align: start;
  }

  .facilities-grid:focus-visible {
    outline: 3px solid var(--color-primary-soft);
    outline-offset: 3px;
  }
}
</style>
