<script setup>
import { useRouter } from 'vue-router'
import { ShieldX } from '@lucide/vue'

import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const handleLogout = async () => {
  await authStore.logoutUser()
  await router.replace('/login')
}
</script>

<template>
  <main class="blocked-view">

    <section class="blocked-panel">

      <span class="icon">
        <ShieldX :size="34" />
      </span>

      <h1>
        Cuenta bloqueada
      </h1>

      <p>
        Tu usuario existe en Poli-REDI, pero no tiene acceso activo al sistema.
      </p>

      <div
        v-if="authStore.error"
        class="message state-card error"
      >
        {{ authStore.error }}
      </div>

      <button
        class="app-button primary"
        type="button"
        @click="handleLogout"
      >
        Cerrar sesión
      </button>

    </section>

  </main>
</template>

<style scoped>
.blocked-view {
  min-height: 100vh;

  background: var(--color-bg);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;
}

.blocked-panel {
  width: min(100%, 440px);

  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);

  padding: 30px;

  display: flex;
  flex-direction: column;
  align-items: flex-start;

  gap: 16px;

  box-shadow: var(--shadow-modal);
}

.icon {
  width: 62px;
  height: 62px;

  border-radius: var(--radius-lg);

  background: var(--color-error-soft);
  color: var(--color-error);

  display: flex;
  align-items: center;
  justify-content: center;
}

h1,
p {
  margin: 0;
}

h1 {
  color: var(--color-text);

  font-size: 28px;
  font-weight: 900;
}

p {
  color: var(--color-text-muted);

  line-height: 1.5;
}

button {
  width: 100%;
  min-height: 48px;
}
</style>
