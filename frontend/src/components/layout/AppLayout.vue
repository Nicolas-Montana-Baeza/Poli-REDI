<script setup>
import { computed } from 'vue'

import RutRequiredModal from '@/components/auth/RutRequiredModal.vue'
import { useAuthStore } from '@/stores/auth'
import { hasRut } from '@/utils/validators'

import HeaderBar from './HeaderBar.vue'

const authStore = useAuthStore()

const shouldRequestRut = computed(() => {
  return (
    authStore.profileReady &&
    !authStore.loading &&
    !authStore.error &&
    authStore.user &&
    authStore.user.isAdmin !== true &&
    !hasRut(authStore.user.rut)
  )
})
</script>

<template>
  <div class="layout">

    <HeaderBar />

    <main class="content">

      <router-view />

    </main>

    <RutRequiredModal :visible="shouldRequestRut" />

  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;

  background: var(--color-bg);
  color: var(--color-text);

  display: flex;
  flex-direction: column;
}

/* CONTENT */
.content {
  flex: 1;

  width: 100%;
  max-width: 1440px;

  margin: 0 auto;
  padding: var(--space-6);
}

/* MOBILE */
@media (max-width: 768px) {
  .content {
    padding: var(--space-4);
  }
}
</style>
