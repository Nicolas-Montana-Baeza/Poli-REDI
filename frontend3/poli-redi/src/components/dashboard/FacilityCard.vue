<script setup>
import { computed } from 'vue'

const props = defineProps({
  name: String,
  type: String,
  status: {
    type: String,
    default: 'available'
  },
  image: String
})

const emit = defineEmits(['select'])

const handleClick = () => {
  emit('select', props)
}

const statusLabel = computed(() => {
  switch (props.status) {
    case 'available': return 'Disponible'
    case 'busy': return 'Ocupado'
    case 'maintenance': return 'Mantenimiento'
    default: return 'Desconocido'
  }
})
</script>

<template>
  <div class="card" @click="handleClick">

    <div class="image" v-if="image">
      <img :src="image" :alt="name" />
    </div>

    <div class="content">
      <h3>{{ name }}</h3>
      <p class="type">{{ type }}</p>

      <span class="status" :class="status">
        {{ statusLabel }}
      </span>
    </div>

  </div>
</template>

<style scoped>
/* (mismo CSS que arriba) */
.card {
  background: white;
  border-radius: 14px;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(0,0,0,0.08);
  transition: all 0.2s ease;
}

.card:hover {
  transform: translateY(-4px);
}

.image img {
  width: 100%;
  height: 140px;
  object-fit: cover;
}

.content {
  padding: 15px;
}

h3 {
  margin: 0;
  font-size: 16px;
}

.type {
  font-size: 12px;
  color: #777;
  margin: 5px 0 10px;
}

.status {
  font-size: 12px;
  padding: 5px 10px;
  border-radius: 20px;
}

.available {
  background: #e0f2fe;
  color: #1e3a8a;
}

.busy {
  background: #ffe4e6;
  color: #b91c1c;
}

.maintenance {
  background: #fff7ed;
  color: #f97316;
}
</style>