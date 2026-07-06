<script setup>
import { computed } from 'vue'

import RutRequiredModal from '@/components/auth/RutRequiredModal.vue'
import { useAuthStore } from '@/stores/auth'

import HeaderBar from './HeaderBar.vue'

const authStore = useAuthStore()

const shouldRequestRut = computed(() => {
  return (
    authStore.user &&
    authStore.user.isAdmin !== true &&
    !authStore.user.rut
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

  background: #f8fafc;

  display: flex;
  flex-direction: column;
}

/* CONTENT */
.content {
  flex: 1;

  padding: 24px;
}

/* MOBILE */
@media (max-width: 768px) {
  .content {
    padding: 16px;
  }
}
</style>
