<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppLayout from './components/layout/AppLayout.vue'
import AuthLoadingScreen from './components/auth/AuthLoadingScreen.vue'

import { useAuthStore } from './stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const isPublicRoute = computed(() => {
  return route.meta.public === true
})

const callbackOwnsLoading = computed(() => {
  return route.path === '/auth/callback'
})

const showAuthBootstrap = computed(() => {
  if (authStore.loggingOut) {
    return true
  }

  return (
    !callbackOwnsLoading.value &&
    !authStore.initialized
  )
})

const authTransitionText = computed(() => {
  if (authStore.loggingOut) {
    return 'Finalizando la sesión de forma segura'
  }

  return 'Verificando tu sesión institucional'
})

const authTransitionTitle = computed(() => {
  if (authStore.loggingOut) {
    return 'Cerrando tu sesión'
  }

  return 'Preparando tu cuenta'
})

const authTransitionDescription = computed(() => {
  if (authStore.loggingOut) {
    return (
      'Estamos cerrando tu sesión y limpiando ' +
      'la información temporal de este dispositivo.'
    )
  }

  return (
    'Estamos cargando tu información institucional ' +
    'y configurando tu acceso.'
  )
})
</script>

<template>

  <AuthLoadingScreen
    v-if="showAuthBootstrap"
    :title="authTransitionTitle"
    :description="authTransitionDescription"
    :stage-text="authTransitionText"
  />

  <RouterView
    v-else-if="isPublicRoute"
  />

  <AppLayout
    v-else
  />

</template>
