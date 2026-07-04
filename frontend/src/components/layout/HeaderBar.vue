<script setup>
import { computed, onMounted } from 'vue'

import Sidebar from './Sidebar.vue'
import NotificationBell from './NotificationBell.vue'
import UserMenu from './UserMenu.vue'

import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const firstName = computed(() => {
  const fullName =
    authStore.user?.fullName ||
    authStore.account?.name ||
    'Usuario'

  return fullName.split(' ')[0]
})

onMounted(() => {
  if (!authStore.user) {
    authStore.loadAuthUser()
  }
})
</script>

<template>
  <header class="header">

    <!-- LEFT -->
    <div class="left">

      <!-- Sidebar -->
      <div class="sidebar-container">
        <Sidebar />
      </div>

      <!-- Greeting -->
      <div class="greeting">

        <h1>
          Hola, {{ firstName }}
        </h1>

        <p>
          Que instalacion deseas reservar hoy?
        </p>

      </div>

    </div>

    <!-- RIGHT -->
    <div class="right">

      <!-- Notifications -->
      <NotificationBell />

      <!-- User -->
      <UserMenu />

    </div>

  </header>
</template>

<style scoped>
.header {
  height: 72px;

  background: white;

  border-bottom: 1px solid #e5e7eb;

  display: flex;
  align-items: center;
  justify-content: space-between;

  padding: 0 24px;

  position: sticky;
  top: 0;

  z-index: 100;
}

/* LEFT */
.left {
  display: flex;
  align-items: center;

  gap: 14px;

  min-width: 0;
}

/* Sidebar wrapper */
.sidebar-container {
  display: flex;
  align-items: center;

  flex-shrink: 0;
}

/* Greeting */
.greeting {
  display: flex;
  flex-direction: column;

  min-width: 0;
}

.greeting h1 {
  margin: 0;

  font-size: 20px;
  font-weight: 700;

  color: #0f172a;

  white-space: nowrap;
}

.greeting p {
  margin: 2px 0 0;

  font-size: 14px;

  color: #64748b;

  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* RIGHT */
.right {
  display: flex;
  align-items: center;

  gap: 16px;

  flex-shrink: 0;
}

/* Mobile */
@media (max-width: 768px) {
  .header {
    padding: 0 16px;
  }

  .left {
    gap: 10px;
  }

  .greeting h1 {
    font-size: 18px;
  }

  .greeting p {
    display: none;
  }
}
</style>
