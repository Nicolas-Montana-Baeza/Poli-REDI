<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  selectedResource: Object
})

const emit = defineEmits(['resource-selected'])

const resources = ref([
  { id: 1, name: 'Cancha 1' },
  { id: 2, name: 'Cancha 2' },
  { id: 3, name: 'Cancha 3' },
  { id: 4, name: 'Gimnasio' },
  { id: 5, name: 'Piscina' },
  { id: 6, name: 'Sala Multiuso' }
])

const localSelected = ref(props.selectedResource || null)

watch(localSelected, (newValue) => {
  emit('resource-selected', newValue)
})
</script>

<template>
  <div class="resource-picker">
    <h2>Selecciona un espacio</h2>

    <div class="resource-list">
      <button
        v-for="resource in resources"
        :key="resource.id"
        class="resource-card"
        :class="{ active: localSelected?.id === resource.id }"
        @click="localSelected = resource"
      >
        {{ resource.name }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.resource-picker {
  background: white;
  border-radius: 16px;
  padding: 20px;
  margin: 20px 0;
}

.resource-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 10px;
}

.resource-card {
  border: 2px solid #eee;
  border-radius: 10px;
  padding: 12px;
  background: white;
  cursor: pointer;
}

.resource-card.active {
  border-color: #4caf50;
  background: #f3fff5;
}
</style>