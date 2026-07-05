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

      throw new Error(authStore.error || 'No se pudo iniciar sesion local.')
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
        class="microsoft-button"
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
          class="local-button"
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

  background:
    linear-gradient(
      135deg,
      #eff6ff,
      #f8fafc
    );

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;
}

.login-panel {
  width: min(100%, 460px);

  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 24px;

  padding: 28px;

  box-shadow:
    0 24px 60px rgba(15,23,42,0.12);

  display: flex;
  flex-direction: column;

  gap: 22px;
}

.brand,
.local-header {
  display: flex;
  align-items: center;

  gap: 14px;
}

.brand-icon {
  width: 54px;
  height: 54px;

  border-radius: 18px;

  background: #1d4ed8;
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
  color: #0f172a;

  font-size: 28px;
  font-weight: 900;
}

h2 {
  color: #0f172a;

  font-size: 17px;
  font-weight: 800;
}

p {
  color: #64748b;

  font-size: 14px;
}

.microsoft-button,
.local-button,
.presets button {
  border: none;
  border-radius: 14px;

  cursor: pointer;

  font-weight: 800;

  transition: 0.2s;
}

.microsoft-button,
.local-button {
  min-height: 50px;

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.microsoft-button {
  background: #1d4ed8;
  color: white;
}

.microsoft-button:hover,
.local-button:hover:not(:disabled) {
  background: #f97316;
}

.local-box {
  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 18px;

  display: flex;
  flex-direction: column;

  gap: 14px;
}

.local-header svg {
  color: #1d4ed8;

  flex-shrink: 0;
}

.presets {
  display: flex;
  flex-wrap: wrap;

  gap: 8px;
}

.presets button {
  background: #eff6ff;
  color: #1d4ed8;

  padding: 8px 11px;
}

label {
  color: #334155;

  display: flex;
  flex-direction: column;

  gap: 7px;

  font-size: 14px;
  font-weight: 800;
}

input {
  width: 100%;
  height: 46px;

  border: 1px solid #dbe2ea;
  border-radius: 14px;

  padding: 0 14px;

  box-sizing: border-box;

  outline: none;
}

input:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}

.local-button {
  background: #0f172a;
  color: white;
}

.local-button:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.error {
  background: #fee2e2;

  border: 1px solid #fecaca;
  border-radius: 14px;

  color: #b91c1c;

  padding: 12px;

  font-size: 14px;
  font-weight: 800;
}

.hint {
  color: #92400e;

  font-size: 12px;
  font-weight: 700;
}
</style>
