<script setup>
import { RouterLink } from 'vue-router'

defineProps({
  label: {
    type: String,
    default: ''
  },

  icon: {
    type: [Object, Function],
    default: null
  },

  to: {
    type: String,
    default: '/'
  }
})
</script>

<template>

  <RouterLink
    :to="to"
    class="sidebar-item"
  >

    <component
      v-if="icon"
      :is="icon"
      :size="20"
      class="icon"
    />

    <span class="label">
      {{ label }}
    </span>

  </RouterLink>

</template>

<style scoped>
.sidebar-item {
  width: 100%;

  display: flex;
  align-items: center;

  gap: 14px;

  padding: 14px 16px;

  border-radius: var(--radius-lg);

  color: var(--color-sidebar-text);

  text-decoration: none;

  transition: 0.2s;

  font-size: 15px;
  font-weight: 500;

  box-sizing: border-box;

  cursor: pointer;
}

.sidebar-item:hover {
  background: rgba(255, 255, 255, 0.08);

  color: white;
}

.router-link-active {
  background: var(--color-primary);

  color: white;

  font-weight: 700;

  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.32);
}

.icon {
  flex-shrink: 0;
}

.label {
  flex: 1;
}
</style>
