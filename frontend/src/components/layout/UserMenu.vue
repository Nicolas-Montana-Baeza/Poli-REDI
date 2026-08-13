<script setup>
import {
  computed,
  onBeforeUnmount,
  onMounted,
  ref
} from 'vue'
import { useRouter } from 'vue-router'

import {
  ChevronDown,
  LogOut,
  Settings
} from '@lucide/vue'

import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'

const authStore = useAuthStore()
const notificationsStore = useNotificationsStore()
const router = useRouter()
const open = ref(false)

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
    ''
  )
})

const roleLabel = computed(() => {
  if (authStore.user?.isAdmin) {
    return 'Administrador'
  }

  return 'Usuario'
})

const initials = computed(() => {
  return displayName.value
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map(part => part[0])
    .join('')
    .toUpperCase() || 'U'
})

const toggle = () => {
  open.value = !open.value
}

const close = (event) => {
  if (!event.target.closest('.user-menu')) {
    open.value = false
  }
}

const handleLogout = async () => {
  open.value = false
  await authStore.logoutUser()
  notificationsStore.clearNotifications()
  await router.replace('/login')
}

const goToSettings = async () => {
  open.value = false
  await router.push('/settings')
}

onMounted(() => {
  window.addEventListener('click', close)

  if (!authStore.user) {
    authStore.loadAuthUser()
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('click', close)
})
</script>

<template>
  <div class="user-menu">

    <button
      class="user-trigger"
      type="button"
      @click.stop="toggle"
    >

      <span class="avatar">
        {{ initials }}
      </span>

      <div class="user-info">

        <strong>
          {{ displayName }}
        </strong>

        <span>
          {{ roleLabel }}
        </span>

      </div>

      <ChevronDown
        :size="16"
        class="arrow"
        :class="{ rotate: open }"
      />

    </button>

    <transition name="fade">

      <div
        v-if="open"
        class="dropdown"
      >

        <div class="profile-card">

          <span class="avatar large">
            {{ initials }}
          </span>

          <div>

            <strong>
              {{ displayName }}
            </strong>

            <p>
              {{ displayEmail }}
            </p>

          </div>

        </div>

        <div class="divider"></div>

        <button
          class="dropdown-item"
          type="button"
          @click="goToSettings"
        >
          <Settings :size="16" />
          Configuración
        </button>

        <div class="divider"></div>

        <button
          class="dropdown-item logout"
          type="button"
          @click="handleLogout"
        >
          <LogOut :size="16" />
          Cerrar sesión
        </button>

      </div>

    </transition>

  </div>
</template>

<style scoped>
.user-menu {
  position: relative;
}

.user-trigger {
  background: transparent;
  border: none;

  display: flex;
  align-items: center;
  gap: 10px;

  cursor: pointer;

  padding: 4px 6px;

  border-radius: var(--radius-md);

  transition: 0.2s;
}

.user-trigger:hover {
  background: var(--color-surface-muted);
}

.avatar {
  width: 42px;
  height: 42px;

  border-radius: 999px;

  background: var(--color-primary);

  color: white;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 14px;
  font-weight: 800;

  flex-shrink: 0;
}

.avatar.large {
  width: 54px;
  height: 54px;

  font-size: 18px;
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.user-info strong {
  font-size: 14px;
  color: var(--color-text);
}

.user-info span {
  font-size: 12px;
  color: var(--color-text-muted);
}

.arrow {
  transition: 0.2s;
}

.rotate {
  transform: rotate(180deg);
}

.dropdown {
  position: absolute;

  top: 58px;
  right: 0;

  width: 260px;

  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  box-shadow: var(--shadow-card);

  overflow: hidden;

  z-index: 999;
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 14px;

  padding: 18px;
}

.profile-card strong {
  display: block;

  font-size: 15px;

  color: var(--color-text);
}

.profile-card p {
  margin: 4px 0 0;

  font-size: 13px;

  color: var(--color-text-muted);
}

.divider {
  height: 1px;

  background: var(--color-border-soft);
}

.dropdown-item {
  width: 100%;

  background: transparent;
  border: none;

  display: flex;
  align-items: center;
  gap: 12px;

  padding: 14px 18px;

  color: var(--color-text);
  font-size: 14px;

  cursor: pointer;

  transition: 0.2s;
}

.dropdown-item:hover {
  background: var(--color-surface-muted);
}

.logout {
  color: var(--color-error-strong);
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@media (max-width: 768px) {
  .user-info {
    display: none;
  }

  .dropdown {
    width: 220px;
  }
}
</style>
