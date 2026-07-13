<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  name: String,
  type: String,
  status: {
    type: String,
    default: 'available'
  },
  image: String
})

const statusLabel = computed(() => {
  switch (props.status) {
    case 'available': return 'Disponible'
    case 'busy': return 'Ocupado'
    case 'maintenance': return 'Mantenimiento'
    default: return 'Desconocido'
  }
})

const imageFailed = ref(false)

watch(
  () => props.image,
  () => {
    imageFailed.value = false
  }
)
</script>

<template>
  <RouterLink
    class="card"
    to="/availability"
  >

    <div
      v-if="image && !imageFailed"
      class="image"
    >
      <img
        :src="image"
        :alt="name"
        @error="imageFailed = true"
      />
    </div>

    <div
      v-else
      class="image fallback"
    >
      <span>
        {{ name?.slice(0, 1) || 'R' }}
      </span>
    </div>

    <div class="content">
      <h3>{{ name }}</h3>
      <p class="type">{{ type }}</p>

      <span class="status" :class="status">
        {{ statusLabel }}
      </span>
    </div>

  </RouterLink>
</template>

<style scoped>
/* (mismo CSS que arriba) */
.card {
  width: 100%;

  background: white;
  border-radius: 14px;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(0,0,0,0.08);
  transition: all 0.2s ease;
  border: none;
  padding: 0;
  text-align: left;
  text-decoration: none;
}

.card:hover {
  transform: translateY(-4px);
}

.image img {
  width: 100%;
  height: 140px;
  object-fit: cover;
}

.image.fallback {
  height: 140px;

  background:
    linear-gradient(
      135deg,
      #dbeafe,
      #f8fafc
    );

  display: flex;
  align-items: center;
  justify-content: center;
}

.image.fallback span {
  width: 56px;
  height: 56px;

  border-radius: 18px;

  background: white;
  color: #1e3a8a;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 24px;
  font-weight: 800;
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
