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
const fallbackFailed = ref(false)

const fallbackImage = computed(() => {
  const name = props.name?.toLowerCase() || ''
  const type = props.type?.toLowerCase() || ''

  if (name.includes('piscina') || type.includes('piscina')) {
    return 'https://images.unsplash.com/photo-1575429198097-0414ec08e8cd?auto=format&fit=crop&w=900&q=80'
  }

  if (name.includes('gimnasio') || type.includes('gimnasio')) {
    return 'https://images.unsplash.com/photo-1534438327276-14e5300c3a48?auto=format&fit=crop&w=900&q=80'
  }

  if (name.includes('muro') || type.includes('escalada')) {
    return 'https://images.unsplash.com/photo-1522163182402-834f871fd851?auto=format&fit=crop&w=900&q=80'
  }

  if (name.includes('spinning')) {
    return 'https://images.unsplash.com/photo-1518611012118-696072aa579a?auto=format&fit=crop&w=900&q=80'
  }

  if (type.includes('sala')) {
    return 'https://images.unsplash.com/photo-1517457373958-b7bdd4587205?auto=format&fit=crop&w=900&q=80'
  }

  return 'https://images.unsplash.com/photo-1522778119026-d647f0596c20?auto=format&fit=crop&w=900&q=80'
})

const displayImage = computed(() => {
  const isGenericStadiumImage =
    String(props.image || '').includes('1522778119026')

  if (props.image && !imageFailed.value && !isGenericStadiumImage) {
    return props.image
  }

  if (!fallbackFailed.value) {
    return fallbackImage.value
  }

  return ''
})

const handleImageError = () => {
  if (props.image && !imageFailed.value) {
    imageFailed.value = true
    return
  }

  fallbackFailed.value = true
}

watch(
  () => props.image,
  () => {
    imageFailed.value = false
    fallbackFailed.value = false
  }
)
</script>

<template>
  <RouterLink
    class="card"
    to="/availability"
  >

    <div
      v-if="displayImage"
      class="image"
    >
      <img
        :src="displayImage"
        :alt="name"
        @error="handleImageError"
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
.card {
  width: 100%;
  height: 100%;

  background: var(--color-surface);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
  border: 1px solid var(--color-border);
  padding: 0;
  text-align: left;
  text-decoration: none;
  display: flex;
  flex-direction: column;
}

.card:hover {
  transform: translateY(-2px);
  border-color: #bfd3ff;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.image img {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 9;
  height: auto;
  object-fit: cover;
  object-position: center;
}

.image.fallback {
  aspect-ratio: 16 / 9;

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

  border-radius: var(--radius-md);

  background: white;
  color: #1e3a8a;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 24px;
  font-weight: 800;
}

.content {
  flex: 1;
  padding: 14px 15px 16px;
}

h3 {
  margin: 0;
  font-size: 16px;
  line-height: 1.25;
  color: var(--color-text);
}

.type {
  font-size: 12px;
  color: var(--color-text-muted);
  margin: 5px 0 10px;
}

.status {
  font-size: 12px;
  padding: 5px 10px;
  border-radius: var(--radius-pill);
  font-weight: 800;
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
