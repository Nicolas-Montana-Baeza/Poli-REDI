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

  padding: 12px 14px;

  border-radius: var(--radius-lg);

  color: var(--color-sidebar-text);

  text-decoration: none;

  transition: 0.2s;

  font-size: 15px;
  font-weight: 600;

  box-sizing: border-box;

  cursor: pointer;
}

.sidebar-item:hover {
  background: var(--color-primary-soft);

  color: var(--color-primary-strong);
}

.router-link-active {
  background: var(--color-primary);

  color: white;

  font-weight: 750;

  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.icon {
  flex-shrink: 0;
}

.label {
  flex: 1;
}
</style>
