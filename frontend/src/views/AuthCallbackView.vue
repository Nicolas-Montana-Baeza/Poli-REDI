<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { initializeAuth } from '../auth/authService'

const router = useRouter()

onMounted(async () => {
  await initializeAuth()

  const redirectPath =
    sessionStorage.getItem('redirectAfterLogin') || '/'

  sessionStorage.removeItem('redirectAfterLogin')

  router.replace(redirectPath)
})
</script>

<template>
  <main style="padding: 32px">
    <h1>Iniciando sesión...</h1>
    <p>Validando acceso institucional.</p>
  </main>
</template>