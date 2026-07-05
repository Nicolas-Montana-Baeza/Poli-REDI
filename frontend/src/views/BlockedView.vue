<script setup>
import { useRouter } from 'vue-router'
import { ShieldX } from 'lucide-vue-next'

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
        class="message"
      >
        {{ authStore.error }}
      </div>

      <button
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

  background: #f8fafc;

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;
}

.blocked-panel {
  width: min(100%, 440px);

  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 20px;

  padding: 30px;

  display: flex;
  flex-direction: column;
  align-items: flex-start;

  gap: 16px;

  box-shadow:
    0 20px 48px rgba(15,23,42,0.12);
}

.icon {
  width: 62px;
  height: 62px;

  border-radius: 18px;

  background: #fee2e2;
  color: #b91c1c;

  display: flex;
  align-items: center;
  justify-content: center;
}

h1,
p {
  margin: 0;
}

h1 {
  color: #0f172a;

  font-size: 28px;
  font-weight: 900;
}

p {
  color: #475569;

  line-height: 1.5;
}

.message {
  width: 100%;

  background: #fef2f2;

  border: 1px solid #fecaca;
  border-radius: 14px;

  color: #b91c1c;

  padding: 12px;

  box-sizing: border-box;

  font-size: 14px;
  font-weight: 800;
}

button {
  width: 100%;
  min-height: 48px;

  border: none;
  border-radius: 14px;

  background: #0f172a;
  color: white;

  cursor: pointer;

  font-weight: 800;
}

button:hover {
  background: #1d4ed8;
}
</style>
