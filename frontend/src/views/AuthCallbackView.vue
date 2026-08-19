<script setup>
import {
  computed,
  onMounted,
  ref
} from 'vue'
import { useRouter } from 'vue-router'
import {
  getSafeRedirectPath,
  initializeAuth
} from '../auth/authService'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const stage = ref('identity')
const fatalError = ref(null)

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
  fatalError.value = null
  stage.value = 'identity'

  try {
    // ----------------------------------------------------------------------
    // MICROSOFT ENTRA ID
    // ----------------------------------------------------------------------
    //
    // Primero dejamos que MSAL procese el redirect y establezca la cuenta
    // autenticada. Todavía no abandonamos esta pantalla porque Poli-REDI
    // necesita reconstruir también su usuario local.
    await initializeAuth()

    stage.value = 'profile'

    // ----------------------------------------------------------------------
    // USUARIO LOCAL POLI-REDI
    // ----------------------------------------------------------------------
    //
    // Entra ID demuestra quién es el usuario. /api/me recupera después los
    // datos propios de Poli-REDI: administrador, bloqueo, RUT y demás reglas
    // de autorización mantenidas por la aplicación.
    const user = await authStore.loadAuthUser()

    if (!user) {
      if (authStore.errorStatus === 403) {
        await router.replace('/blocked')
        return
      }

      throw new Error(
        authStore.error ||
        'No se pudo cargar tu cuenta institucional.'
      )
    }

    stage.value = 'permissions'

    const redirectPath = getSafeRedirectPath(
      sessionStorage.getItem('redirectAfterLogin')
    )

    sessionStorage.removeItem('redirectAfterLogin')

    // Una pequeña espera evita un cambio visual demasiado brusco cuando
    // Microsoft y el backend responden casi instantáneamente.
    await new Promise((resolve) => {
      window.setTimeout(resolve, 250)
    })

    await router.replace(redirectPath)
  } catch (error) {
    fatalError.value =
      error?.message ||
      'No fue posible completar el inicio de sesión.'
  }
}

const goToLogin = async () => {
  await router.replace('/login')
}

onMounted(continueLogin)
</script>

<template>
  <main class="auth-loading-page">
    <section class="loading-card">
      <div class="brand">
        <div class="brand-mark">
          P
        </div>

        <div>
          <strong>
            POLI REDI
          </strong>

          <span>
            Sistema de Reservas
          </span>
        </div>
      </div>

      <template v-if="!fatalError">
        <div
          class="spinner"
          aria-hidden="true"
        />

        <div class="content">
          <p class="eyebrow">
            Acceso institucional
          </p>

          <h1>
            Preparando tu cuenta
          </h1>

          <p class="status">
            {{ stageText }}
          </p>

          <p class="description">
            Estamos cargando tu información institucional
            y configurando tu acceso.
          </p>
        </div>

        <div class="progress">
          <span />
        </div>

        <p class="footer-note">
          Universidad Central de Chile
        </p>
      </template>

      <template v-else>
        <div class="error-icon">
          !
        </div>

        <div class="content">
          <p class="eyebrow">
            Acceso institucional
          </p>

          <h1>
            No pudimos cargar tu cuenta
          </h1>

          <p class="description error-message">
            {{ fatalError }}
          </p>
        </div>

        <div class="actions">
          <button
            type="button"
            class="primary-button"
            @click="continueLogin"
          >
            Reintentar
          </button>

          <button
            type="button"
            class="secondary-button"
            @click="goToLogin"
          >
            Volver al inicio
          </button>
        </div>
      </template>
    </section>
  </main>
</template>

<style scoped>
.auth-loading-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  box-sizing: border-box;
  padding: 2rem;
  background:
    radial-gradient(
      circle at top,
      rgba(37, 99, 235, 0.12),
      transparent 42%
    ),
    var(--app-bg, #f7f8fb);
}

.loading-card {
  width: min(100%, 440px);
  box-sizing: border-box;
  padding: 2rem;
  border: 1px solid rgba(127, 127, 127, 0.18);
  border-radius: 1.4rem;
  background: var(--card-bg, #ffffff);
  box-shadow:
    0 24px 60px rgba(15, 23, 42, 0.08);
  text-align: center;
}

.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.7rem;
  margin-bottom: 2.5rem;
}

.brand > div:last-child {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.brand strong {
  letter-spacing: 0.08em;
}

.brand span {
  margin-top: 0.1rem;
  font-size: 0.72rem;
  opacity: 0.55;
}

.brand-mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 0.8rem;
  background: #2563eb;
  color: white;
  font-weight: 800;
  font-size: 1.15rem;
}

.spinner {
  width: 48px;
  height: 48px;
  box-sizing: border-box;
  margin: 0 auto 1.5rem;
  border: 4px solid rgba(37, 99, 235, 0.14);
  border-top-color: #2563eb;
  border-radius: 50%;
  animation: spin 0.85s linear infinite;
}

.content h1 {
  margin: 0.35rem 0 0.65rem;
  font-size: 1.65rem;
}

.eyebrow {
  margin: 0;
  color: #2563eb;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.09em;
  text-transform: uppercase;
}

.status {
  margin: 0;
  font-weight: 650;
}

.description {
  margin: 0.65rem auto 0;
  max-width: 330px;
  line-height: 1.55;
  opacity: 0.62;
}

.progress {
  height: 4px;
  margin: 1.8rem 0 1rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.1);
}

.progress span {
  display: block;
  width: 45%;
  height: 100%;
  border-radius: inherit;
  background: #2563eb;
  animation: progress 1.4s ease-in-out infinite;
}

.footer-note {
  margin: 0;
  font-size: 0.78rem;
  opacity: 0.45;
}

.error-icon {
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  margin: 0 auto 1.5rem;
  border-radius: 50%;
  background: rgba(220, 38, 38, 0.12);
  color: #dc2626;
  font-size: 1.4rem;
  font-weight: 800;
}

.error-message {
  color: #b91c1c;
  opacity: 1;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin-top: 1.6rem;
}

.primary-button,
.secondary-button {
  width: 100%;
  border-radius: 0.75rem;
  padding: 0.8rem 1rem;
  cursor: pointer;
  font: inherit;
  font-weight: 700;
}

.primary-button {
  border: 0;
  background: #2563eb;
  color: white;
}

.secondary-button {
  border: 1px solid rgba(127, 127, 127, 0.25);
  background: transparent;
  color: inherit;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes progress {
  0% {
    transform: translateX(-110%);
  }

  50% {
    transform: translateX(120%);
  }

  100% {
    transform: translateX(260%);
  }
}

@media (max-width: 520px) {
  .auth-loading-page {
    padding: 1rem;
  }

  .loading-card {
    padding: 1.5rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .spinner,
  .progress span {
    animation-duration: 2.5s;
  }
}
</style>
