<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  selectedSport: String,
  selectedResource: Object
})

const emit = defineEmits(['sport-selected'])

const sportsMap = {
  'Cancha 1': ['Fútbol', 'Vóley', 'Básquet', 'Tenis'],
  'Cancha 2': ['Fútbol', 'Vóley', 'Básquet', 'Tenis'],
  'Cancha 3': ['Fútbol', 'Vóley', 'Básquet', 'Tenis'],
  'Gimnasio': ['Musculación', 'Cardio', 'Libre'],
  'Piscina': ['Natación', 'Recreativo'],
  'Sala Multiuso': ['Baile', 'Yoga', 'Evento']
}

const localSelected = ref(props.selectedSport || null)

watch(localSelected, (newValue) => {
  emit('sport-selected', newValue)
})
</script>

<template>
  <div class="sport-picker" v-if="selectedResource">
    <h2>Selecciona actividad</h2>

    <div class="sport-list">
      <button
        v-for="sport in sportsMap[selectedResource.name] || []"
        :key="sport"
        class="sport-card"
        :class="{ active: localSelected === sport }"
        @click="localSelected = sport"
      >
        {{ sport }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.sport-picker {
  background: white;
  border-radius: 16px;
  padding: 20px;
  margin: 20px 0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.sport-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.sport-card {
  border: 2px solid #eee;
  border-radius: 10px;
  padding: 10px 16px;
  background: white;
  cursor: pointer;
  transition: 0.2s;
}

.sport-card:hover {
  border-color: #4caf50;
}

.sport-card.active {
  border-color: #4caf50;
  background: #f3fff5;
}
</style>