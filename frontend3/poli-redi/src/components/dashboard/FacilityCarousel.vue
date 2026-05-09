<script setup>
import { ref } from 'vue'
import FacilityCard from './FacilityCard.vue'

const props = defineProps({
  facilities: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['select'])

const scrollRef = ref(null)

const handleSelect = (facility) => {
  emit('select', facility)
}
</script>

<template>
  <div class="wrapper">

    <div class="header">
      <h2>Instalaciones disponibles</h2>
    </div>

    <!-- CAROUSEL -->
    <div
      class="carousel"
      ref="scrollRef"
    >

      <div
        v-for="(facility, index) in facilities"
        :key="index"
        class="item"
      >
        <FacilityCard
          :name="facility.name"
          :type="facility.type"
          :status="facility.status"
          :image="facility.image"
          @select="handleSelect(facility)"
        />
      </div>

    </div>

  </div>
</template>

<style scoped>
.wrapper {
  width: 100%;
  overflow: hidden;
}

/* HEADER */
.header {
  margin-bottom: 18px;
}

.header h2 {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
}

/* CAROUSEL */
.carousel {
  display: flex;

  gap: 16px;

  overflow-x: auto;
  overflow-y: hidden;

  scroll-snap-type: x mandatory;
  scroll-behavior: smooth;

  -webkit-overflow-scrolling: touch;

  padding-bottom: 12px;
}

/* Scrollbar */
.carousel::-webkit-scrollbar {
  height: 6px;
}

.carousel::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 999px;
}

/* ITEM */
.item {
  flex: 0 0 100%;

  scroll-snap-align: start;
}

/* TABLET */
@media (min-width: 768px) {
  .item {
    flex: 0 0 70%;
  }
}

/* DESKTOP */
@media (min-width: 1200px) {
  .item {
    flex: 0 0 45%;
  }
}
</style>