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

const scrollAmount = 320

const scrollLeft = () => {
  scrollRef.value?.scrollBy({
    left: -scrollAmount,
    behavior: 'smooth'
  })
}

const scrollRight = () => {
  scrollRef.value?.scrollBy({
    left: scrollAmount,
    behavior: 'smooth'
  })
}

const handleSelect = (facility) => {
  emit('select', facility)
}
</script>

<template>
  <div class="wrapper">

    <!-- HEADER -->
    <div class="header">
      <h2>Instalaciones disponibles</h2>

      <div class="controls">
        <button @click="scrollLeft">
          ‹
        </button>

        <button @click="scrollRight">
          ›
        </button>
      </div>
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
  position: relative;
}

/* HEADER */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  margin-bottom: 18px;
}

.header h2 {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
}

/* CONTROLES */
.controls {
  display: flex;
  gap: 10px;
}

.controls button {
  width: 42px;
  height: 42px;

  border: none;
  border-radius: 999px;

  background: #1e3a8a;
  color: white;

  font-size: 22px;
  cursor: pointer;

  transition: 0.2s;

  display: flex;
  align-items: center;
  justify-content: center;
}

.controls button:hover {
  background: #f97316;
  transform: scale(1.05);
}

/* CAROUSEL */
.carousel {
  display: flex;
  gap: 18px;

  overflow-x: auto;
  overflow-y: hidden;

  scroll-behavior: smooth;
  scroll-snap-type: x mandatory;

  padding-bottom: 12px;

  width: 100%;
}

/* SCROLLBAR */
.carousel::-webkit-scrollbar {
  height: 8px;
}

.carousel::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 999px;
}

/* ITEMS */
.item {
  flex: 0 0 auto;

  width: 280px;

  scroll-snap-align: start;
}

/* TABLET */
@media (max-width: 1024px) {
  .item {
    width: 240px;
  }
}

/* MOBILE */
@media (max-width: 768px) {
  .item {
    width: 85%;
  }

  .header h2 {
    font-size: 18px;
  }

  .controls button {
    width: 36px;
    height: 36px;
  }
}
</style>