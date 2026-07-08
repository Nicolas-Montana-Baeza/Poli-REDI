<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  LogIn,
  ShieldCheck,
  UserPlus
} from 'lucide-vue-next'

import {
  isDevAuthEnabled,
  login,
  loginLocal
} from '@/auth/authService'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const fullName = ref('Usuario Prueba')
const email = ref('nicolas@universidad.cl')
const error = ref('')
const loading = ref(false)

const redirectPath = () => {
  return route.query.redirect || '/'
}

const handleMicrosoftLogin = async () => {
  await login(redirectPath())
}

const usePreset = (preset) => {
  fullName.value = preset.fullName
  email.value = preset.email
}

const handleLocalLogin = async () => {
  error.value = ''
  loading.value = true

  try {
    loginLocal({
      email: email.value,
      fullName: fullName.value
    })

    const user = await authStore.loadAuthUser()

    if (!user) {
      if (authStore.errorStatus === 403) {
        router.replace('/blocked')
        return
      }

      throw new Error(authStore.error || 'No se pudo iniciar sesión local.')
    }

    router.replace(redirectPath())
  } catch (localError) {
    error.value = localError.message
  } finally {
    loading.value = false
  }
}

const presets = [
  {
    label: 'Estudiante',
    fullName: 'Usuario Prueba',
    email: 'nicolas@universidad.cl'
  },
  {
    label: 'Admin',
    fullName: 'Administrador General',
    email: 'admin@universidad.cl'
  },
  {
    label: 'Nuevo local',
    fullName: 'Usuario Local',
    email: 'local@universidad.cl'
  }
]
</script>

<template>
  <main class="login-view">

    <section class="login-panel">

      <div class="brand">

        <span class="brand-icon">
          <ShieldCheck :size="28" />
        </span>

        <div>

          <h1>
            Poli-REDI
          </h1>

          <p>
            Acceso temporal para pruebas del MVP.
          </p>

        </div>

      </div>

      <button
        class="microsoft-button app-button primary"
        type="button"
        @click="handleMicrosoftLogin"
      >
        <LogIn :size="19" />
        Continuar con Microsoft
      </button>

      <div
        v-if="isDevAuthEnabled()"
        class="local-box"
      >

        <div class="local-header">

          <UserPlus :size="19" />

          <div>

            <h2>
              Acceso local de prueba
            </h2>

            <p>
              Bypass temporal de Entra para probar flujos.
            </p>

          </div>

        </div>

        <div class="presets">

          <button
            v-for="preset in presets"
            :key="preset.email"
            type="button"
            class="app-badge"
            @click="usePreset(preset)"
          >
            {{ preset.label }}
          </button>

        </div>

        <label>
          Nombre
          <input
            v-model="fullName"
            type="text"
            autocomplete="name"
          />
        </label>

        <label>
          Correo
          <input
            v-model="email"
            type="email"
            autocomplete="email"
          />
        </label>

        <div
          v-if="error"
          class="error"
        >
          {{ error }}
        </div>

        <button
          class="local-button app-button"
          type="button"
          :disabled="loading"
          @click="handleLocalLogin"
        >
          {{ loading ? 'Entrando...' : 'Entrar / registrar local' }}
        </button>

        <p class="hint">
          Requiere `DEV_AUTH_ENABLED=true` en el backend.
        </p>

      </div>

    </section>

  </main>
</template>

<style scoped>
.login-view {
  min-height: 100vh;

  background: var(--color-bg);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;
}

.login-panel {
  width: min(100%, 440px);

  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);

  padding: 26px;

  box-shadow: var(--shadow-card);

  display: flex;
  flex-direction: column;

  gap: 20px;
}

.brand,
.local-header {
  display: flex;
  align-items: center;

  gap: 14px;
}

.brand-icon {
  width: 48px;
  height: 48px;

  border-radius: var(--radius-lg);

  background: var(--color-primary);
  color: white;

  display: flex;
  align-items: center;
  justify-content: center;
}

h1,
h2,
p {
  margin: 0;
}

h1 {
  color: var(--color-text);

  font-size: 27px;
  font-weight: 850;
}

h2 {
  color: var(--color-text);

  font-size: 17px;
  font-weight: 750;
}

p {
  color: var(--color-text-muted);

  font-size: 14px;
}

.microsoft-button,
.local-button {
  min-height: 48px;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.local-box {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  padding: 17px;

  display: flex;
  flex-direction: column;

  gap: 14px;
}

.local-header svg {
  color: var(--color-primary);

  flex-shrink: 0;
}

.presets {
  display: flex;
  flex-wrap: wrap;

  gap: 8px;
}

label {
  color: #334155;

  display: flex;
  flex-direction: column;

  gap: 7px;

  font-size: 14px;
  font-weight: 700;
}

input {
  width: 100%;
  height: 44px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  padding: 0 14px;

  box-sizing: border-box;

  outline: none;
}

.local-button {
  background: var(--color-primary);
  color: white;
}

.local-button:hover:not(:disabled) {
  background: var(--color-primary-strong);
}

.local-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.error {
  background: var(--color-error-soft);

  border: 1px solid var(--color-error-border);
  border-radius: var(--radius-md);

  color: var(--color-error);

  padding: 12px;

  font-size: 14px;
  font-weight: 700;
}

.hint {
  color: var(--color-warning);

  font-size: 12px;
  font-weight: 700;
}
</style>
