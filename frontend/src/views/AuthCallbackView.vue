<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  getSafeRedirectPath,
  initializeAuth
} from '../auth/authService'

const router = useRouter()

onMounted(async () => {
  try {
    await initializeAuth()

    const redirectPath = getSafeRedirectPath(
      sessionStorage.getItem('redirectAfterLogin')
    )

    sessionStorage.removeItem('redirectAfterLogin')

    router.replace(redirectPath)
  } catch {
    router.replace('/')
  }
})
</script>

<template>
  <main style="padding: 32px">
    <h1>Iniciando sesión...</h1>
    <p>Validando acceso institucional.</p>
  </main>
</template>
