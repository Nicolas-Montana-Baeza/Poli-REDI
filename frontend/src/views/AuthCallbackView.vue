<script setup>
import {
  computed,
  onMounted,
  ref
} from 'vue'

import { useRouter } from 'vue-router'

import AuthLoadingScreen from '@/components/auth/AuthLoadingScreen.vue'

import {
  getSafeRedirectPath,
  initializeAuth
} from '@/auth/authService'

import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const stage = ref('identity')
const fatalError = ref('')

const stageText = computed(() => {
  switch (stage.value) {
    case 'identity':
      return 'Validando tu identidad institucional'

    case 'profile':
      return 'Cargando tu cuenta en Poli-REDI'

    case 'permissions':
      return 'Preparando tus permisos y preferencias'

    default:
      return 'Preparando Poli-REDI'
  }
})

const continueLogin = async () => {
  fatalError.value = ''
  stage.value = 'identity'

  try {
    // Microsoft procesa el redirect y recupera
    // la cuenta institucional.
    await initializeAuth()

    stage.value = 'profile'

    // Poli-REDI recupera el perfil local y permisos.
    const user =
      await authStore.loadAuthUser()

    if (!user) {
      if (
        authStore.errorStatus === 403
      ) {
        await router.replace('/blocked')
        return
      }

      throw new Error(
        authStore.error ||
        'No se pudo cargar tu cuenta institucional.'
      )
    }

    stage.value = 'permissions'

    const redirectPath =
      getSafeRedirectPath(
        sessionStorage.getItem(
          'redirectAfterLogin'
        )
      )

    sessionStorage.removeItem(
      'redirectAfterLogin'
    )

    // Evita un cambio visual demasiado brusco
    // cuando todo responde instantáneamente.
    await new Promise((resolve) => {
      window.setTimeout(
        resolve,
        250
      )
    })

    await router.replace(
      redirectPath
    )
  } catch (error) {
    fatalError.value =
      error?.message ||
      'No fue posible completar el inicio de sesión.'
  }
}

const goToLogin = async () => {
  await router.replace('/login')
}

onMounted(
  continueLogin
)
</script>

<template>

  <AuthLoadingScreen
    :stage-text="stageText"
    :error-message="fatalError"
    show-actions
    @retry="continueLogin"
    @back="goToLogin"
  />

</template>
