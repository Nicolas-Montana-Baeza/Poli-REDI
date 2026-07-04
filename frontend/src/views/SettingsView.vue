<script setup>
import { computed, onMounted } from 'vue'

import {
  Mail,
  Power,
  ShieldCheck,
  UserRound
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

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
}

onMounted(() => {
  if (!authStore.user) {
    authStore.loadAuthUser()
  }
})
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
          Cerrar sesion
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

@media (max-width: 768px) {
  .account-layout {
    grid-template-columns: 1fr;
  }

  .page-header h1 {
    font-size: 26px;
  }
}
</style>
