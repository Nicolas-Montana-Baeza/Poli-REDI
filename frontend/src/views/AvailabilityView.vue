<script setup>
import AvailabilitySection from
'../components/availability/AvailabilitySection.vue'
</script>

<template>
  <main class="availability-view">

    <AvailabilitySection />

  </main>
</template>

<style scoped>
.availability-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}
</style>