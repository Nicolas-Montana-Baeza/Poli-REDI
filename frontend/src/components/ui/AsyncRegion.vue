<script setup>
import SkeletonLoader from './SkeletonLoader.vue'

defineProps({
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: [String, Object, Boolean],
    default: ''
  },
  empty: {
    type: Boolean,
    default: false
  },
  loadingLabel: {
    type: String,
    default: 'Cargando contenido'
  },
  skeletonVariant: {
    type: String,
    default: 'list'
  },
  skeletonItems: {
    type: Number,
    default: 3
  },
  skeletonColumns: {
    type: Number,
    default: null
  },
  mobileCarousel: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <section
    class="async-region"
    :aria-busy="loading"
  >
    <template v-if="loading">
      <p class="sr-only" role="status" aria-live="polite">{{ loadingLabel }}</p>
      <slot name="loading">
        <SkeletonLoader
          :variant="skeletonVariant"
          :items="skeletonItems"
          :columns="skeletonColumns"
          :mobile-carousel="mobileCarousel"
        />
      </slot>
    </template>

    <template v-else-if="error">
      <slot name="error" :error="error">
        <div class="state-card error" role="alert">
          {{ typeof error === 'string' ? error : 'No se pudo cargar el contenido.' }}
        </div>
      </slot>
    </template>

    <template v-else-if="empty">
      <slot name="empty">
        <div class="state-card">No hay información disponible.</div>
      </slot>
    </template>

    <slot v-else />
  </section>
</template>

<style scoped>
.async-region {
  width: 100%;
  min-width: 0;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
