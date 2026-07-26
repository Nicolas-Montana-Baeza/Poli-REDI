<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  IdCard,
  Mail,
  Power,
  ShieldCheck,
  UserRound
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { formatRutInput, isValidRut, normalizeRut } from '@/utils/validators'

const authStore = useAuthStore()
const notificationsStore = useNotificationsStore()
const route = useRoute()
const router = useRouter()
const rutValue = ref('')
const rutError = ref('')
const rutSuccess = ref('')
const rutSaving = ref(false)

const displayName = computed(() => {
  return (
    authStore.user?.fullName ||
    authStore.account?.name ||
    authStore.user?.email ||
    authStore.account?.username ||
    'Usuario'
  )
})

const displayEmail = computed(() => {
  return (
    authStore.user?.email ||
    authStore.account?.username ||
    'No disponible'
  )
})

const roleLabel = computed(() => {
  if (authStore.user?.isAdmin) {
    return 'Administrador'
  }

  return 'Usuario'
})

const statusLabel = computed(() => {
  if (authStore.user?.isBlocked) {
    return 'Bloqueado'
  }

  return 'Activo'
})

const rutRequired = computed(() => {
  return authStore.user?.isAdmin !== true
})

const rutStatusLabel = computed(() => {
  if (authStore.user?.rut) {
    return authStore.user.rut
  }

  return rutRequired.value ? 'Pendiente' : 'No requerido para administradores.'
})

const initials = computed(() => {
  return displayName.value
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'U'
})

const handleLogout = async () => {
  await authStore.logoutUser()
  notificationsStore.clearNotifications()
  await router.replace('/login')
}

const handleRutSubmit = async () => {
  rutError.value = ''
  rutSuccess.value = ''

  const normalized = normalizeRut(rutValue.value)

  if (rutRequired.value && !normalized) {
    rutError.value = 'Ingresa tu RUT para continuar.'
    return
  }

  if (normalized && !isValidRut(normalized)) {
    rutError.value = 'El RUT ingresado no es válido.'
    return
  }

  rutSaving.value = true

  try {
    await authStore.updateRut(normalized)

    rutValue.value = authStore.user?.rut || normalized
    rutSuccess.value = 'RUT actualizado correctamente.'

    if (route.query.redirect) {
      router.replace(String(route.query.redirect))
    }
  } catch (error) {
    rutError.value = error.message
  } finally {
    rutSaving.value = false
  }
}

const handleRutInput = () => {
  rutValue.value = formatRutInput(rutValue.value)
  rutError.value = ''
  rutSuccess.value = ''
}

onMounted(() => {
  if (!authStore.user) {
    authStore.loadAuthUser()
  }
})

watch(
  () => authStore.user?.rut,
  (rut) => {
    rutValue.value = rut || ''
  },
  {
    immediate: true
  }
)
</script>

<template>
  <main class="settings-view">

    <header class="page-header">

      <h1>
        Cuenta
      </h1>

      <p>
        Datos del usuario autenticado.
      </p>

    </header>

    <div
      v-if="authStore.loading"
      aria-label="Cargando usuario"
    >
      <SkeletonLoader
        variant="resources"
        :items="3"
      />
    </div>

    <div
      v-else-if="authStore.error"
      class="state-card error"
    >
      {{ authStore.error }}
    </div>

    <section
      v-else
      class="account-layout"
    >

      <article class="profile-panel">

        <span class="avatar">
          {{ initials }}
        </span>

        <div>

          <h2>
            {{ displayName }}
          </h2>

          <p>
            {{ displayEmail }}
          </p>

        </div>

        <button
          class="logout-button"
          type="button"
          @click="handleLogout"
        >
          <Power :size="18" />
          Cerrar sesión
        </button>

      </article>

      <section class="details-grid">

        <article class="detail-card">

          <UserRound :size="22" />

          <span>
            Nombre
          </span>

          <strong>
            {{ displayName }}
          </strong>

        </article>

        <article class="detail-card">

          <Mail :size="22" />

          <span>
            Correo
          </span>

          <strong>
            {{ displayEmail }}
          </strong>

        </article>

        <article class="detail-card">

          <IdCard :size="22" />

          <span>
            RUT
          </span>

          <strong>
            {{ rutStatusLabel }}
          </strong>

          <form
            v-if="rutRequired && !authStore.user?.rut"
            class="rut-form"
            @submit.prevent="handleRutSubmit"
          >

            <input
              v-model="rutValue"
              type="text"
              inputmode="text"
              placeholder="12345678-5"
              :disabled="rutSaving"
              @input="handleRutInput"
            />

            <button
              type="submit"
              :disabled="rutSaving"
            >
              {{ rutSaving ? 'Guardando...' : 'Guardar RUT' }}
            </button>

          </form>

          <p v-else-if="authStore.user?.rut" class="hint">
            El RUT registrado es de solo lectura.
          </p>

          <p v-else class="hint">
            No requerido para administradores.
          </p>

          <p
            v-if="rutRequired && !authStore.user?.rut"
            class="hint warning"
          >
            Debes registrar tu RUT antes de crear reservas.
          </p>

          <p
            v-if="rutError"
            class="hint error"
          >
            {{ rutError }}
          </p>

          <p
            v-if="rutSuccess"
            class="hint success"
          >
            {{ rutSuccess }}
          </p>

        </article>

        <article class="detail-card">

          <ShieldCheck :size="22" />

          <span>
            Rol
          </span>

          <strong>
            {{ roleLabel }}
          </strong>

        </article>

        <article class="detail-card">

          <ShieldCheck :size="22" />

          <span>
            Estado
          </span>

          <strong>
            {{ statusLabel }}
          </strong>

        </article>

      </section>

    </section>

  </main>
</template>

<style scoped>
.settings-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  font-size: 30px;
  font-weight: 800;

  color: #0f172a;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.state-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 22px;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.account-layout {
  display: grid;

  grid-template-columns:
    minmax(240px, 320px)
    minmax(0, 1fr);

  gap: 18px;
}

.profile-panel,
.detail-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.profile-panel {
  padding: 24px;

  display: flex;
  flex-direction: column;
  align-items: flex-start;

  gap: 18px;
}

.avatar {
  width: 72px;
  height: 72px;

  border-radius: 999px;

  background:
    linear-gradient(
      135deg,
      #1e3a8a,
      #2563eb
    );

  color: white;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 24px;
  font-weight: 800;
}

.profile-panel h2 {
  margin: 0;

  font-size: 22px;
  font-weight: 800;

  color: #0f172a;
}

.profile-panel p {
  margin: 6px 0 0;

  color: #64748b;
}

.logout-button {
  width: 100%;

  background: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 14px;

  color: #b91c1c;

  cursor: pointer;

  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;

  padding: 12px 16px;

  font-weight: 800;

  transition: 0.2s;
}

.logout-button:hover {
  background: #fecaca;
}

.details-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(220px, 1fr));

  gap: 18px;
}

.detail-card {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 10px;

  color: #1d4ed8;
}

.detail-card span {
  color: #64748b;

  font-size: 13px;
  font-weight: 700;
}

.detail-card strong {
  color: #0f172a;

  font-size: 16px;
  overflow-wrap: anywhere;
}

.rut-form {
  display: flex;
  flex-direction: column;

  gap: 10px;
}

.rut-form input,
.rut-form button {
  width: 100%;

  min-height: 44px;

  border-radius: 14px;

  box-sizing: border-box;

  font: inherit;
}

.rut-form input {
  border: 1px solid #dbe2ea;

  padding: 0 14px;

  outline: none;
}

.rut-form input:focus {
  border-color: #2563eb;

  box-shadow:
    0 0 0 4px rgba(37,99,235,0.08);
}

.rut-form button {
  border: none;

  background: #1d4ed8;
  color: white;

  cursor: pointer;

  font-weight: 800;
}

.rut-form button:disabled,
.rut-form input:disabled {
  cursor: not-allowed;

  opacity: 0.65;
}

.hint {
  margin: 0;

  font-size: 13px;
  font-weight: 700;
}

.hint.warning {
  color: #92400e;
}

.hint.error {
  color: #b91c1c;
}

.hint.success {
  color: #166534;
}

@media (max-width: 768px) {
  .account-layout {
    grid-template-columns: 1fr;
  }

  .page-header h1 {
    font-size: 26px;
  }
}
</style>
