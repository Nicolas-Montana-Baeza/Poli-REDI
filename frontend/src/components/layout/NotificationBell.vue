<script setup>
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue'

import {
  Bell,
  Calendar,
  AlertTriangle,
  CheckCircle2
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'

const authStore = useAuthStore()
const notificationsStore = useNotificationsStore()
const open = ref(false)

const notifications = computed(() => {
  return notificationsStore.notifications
})

const unreadCount = computed(() => {
  return notificationsStore.unreadCount
})

const isAuthenticated = computed(() => {
  return !!authStore.account
})

const toggle = () => {
  open.value = !open.value
}

const close = (event) => {
  if (!event.target.closest('.notification-wrapper')) {
    open.value = false
  }
}

const iconFor = (type) => {
  switch (type) {
    case 'SYSTEM':
      return AlertTriangle

    case 'RESERVATION_CONFIRMED':
    case 'RESERVATION_CREATED':
      return CheckCircle2

    default:
      return Calendar
  }
}

const formatNotificationTime = (createdAt) => {
  if (!createdAt) {
    return ''
  }

  return new Date(createdAt).toLocaleString('es-CL', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  window.addEventListener('click', close)
})

watch(
  isAuthenticated,
  (authenticated) => {
    if (authenticated) {
      notificationsStore.fetchNotifications()
    } else {
      notificationsStore.clearNotifications()
    }
  },
  {
    immediate: true
  }
)

onBeforeUnmount(() => {
  window.removeEventListener('click', close)
})
</script>

<template>
  <div class="notification-wrapper">

    <!-- Trigger -->
    <button
      class="notification-btn"
      type="button"
      aria-label="Abrir notificaciones"
      aria-haspopup="menu"
      :aria-expanded="open"
      @click.stop="toggle"
    >

      <Bell :size="18" />

      <span
        v-if="unreadCount"
        class="badge"
      >
        {{ unreadCount }}
      </span>

    </button>

    <!-- Dropdown -->
    <transition name="fade">

      <div
        v-if="open"
        class="dropdown"
      >

        <!-- Header -->
        <div class="dropdown-header">

          <h3>
            Notificaciones
          </h3>

          <span>
            {{ unreadCount }} nuevas
          </span>

        </div>

        <!-- Loading -->
        <div
          v-if="notificationsStore.loading"
          class="notification-skeleton"
          aria-label="Cargando notificaciones"
          role="status"
          aria-live="polite"
        >
          <SkeletonLoader
            variant="compact-rows"
            :items="3"
          />
        </div>

        <!-- Error -->
        <div
          v-else-if="notificationsStore.error"
          class="empty"
        >
          {{ notificationsStore.error }}
        </div>

        <!-- Empty -->
        <div
          v-else-if="!notifications.length"
          class="empty"
        >
          No hay notificaciones
        </div>

        <!-- Items -->
        <template v-else>

          <div
            v-for="notification in notifications"
            :key="notification.id"
            class="notification-item"
          >

            <!-- Icon -->
            <div class="icon">

              <component
                :is="iconFor(notification.type)"
                :size="18"
              />

            </div>

            <!-- Content -->
            <div class="content">

              <strong>
                {{ notification.title }}
              </strong>

              <p>
                {{ notification.message }}
              </p>

              <span>
                {{ formatNotificationTime(notification.createdAt) }}
              </span>

            </div>

          </div>

        </template>

        <!-- Footer -->
        <button
          class="view-all"
          type="button"
        >
          Ver todas
        </button>

      </div>

    </transition>

  </div>
</template>

<style scoped>
.notification-wrapper {
  position: relative;
}

/* Trigger */
.notification-btn {
  position: relative;

  width: 42px;
  height: 42px;

  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);

  background: var(--color-surface);

  display: flex;
  align-items: center;
  justify-content: center;

  cursor: pointer;

  transition: 0.2s;
}

.notification-btn:hover {
  background: var(--color-surface-muted);
}

/* Badge */
.badge {
  position: absolute;

  top: 6px;
  right: 6px;

  min-width: 16px;
  height: 16px;

  padding: 0 4px;

  border-radius: 999px;

  background: var(--color-error-strong);

  color: white;

  font-size: 10px;
  font-weight: 700;

  display: flex;
  align-items: center;
  justify-content: center;
}

/* Dropdown */
.dropdown {
  position: absolute;

  top: 52px;
  right: 0;

  width: 340px;
  max-height: 420px;

  overflow-y: auto;

  background: var(--color-surface);

  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);

  box-shadow: var(--shadow-card);

  z-index: 999;
}

/* Header */
.dropdown-header {
  padding: 18px;

  border-bottom: 1px solid var(--color-border);

  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dropdown-header h3 {
  margin: 0;

  font-size: 16px;
  font-weight: 700;

  color: var(--color-text);
}

.dropdown-header span {
  font-size: 12px;

  color: var(--color-text-muted);
}

/* Notification */
.notification-item {
  display: flex;
  gap: 14px;

  padding: 16px 18px;

  border-bottom: 1px solid var(--color-border-soft);

  transition: 0.2s;

  cursor: pointer;
}

.notification-item:hover {
  background: var(--color-surface-muted);
}

/* Icon */
.icon {
  width: 38px;
  height: 38px;

  border-radius: 12px;

  background: var(--color-primary-soft);

  color: var(--color-primary);

  display: flex;
  align-items: center;
  justify-content: center;

  flex-shrink: 0;
}

/* Content */
.content {
  flex: 1;
}

.content strong {
  display: block;

  font-size: 14px;
  color: var(--color-text);
}

.content p {
  margin: 4px 0;

  font-size: 13px;
  color: #475569;
}

.content span {
  font-size: 12px;
  color: var(--color-text-soft);
}

/* Empty */
.empty {
  padding: 30px;

  text-align: center;

  color: var(--color-text-soft);
}

.notification-skeleton {
  padding: 14px;
}

/* Footer */
.view-all {
  width: 100%;

  padding: 14px;

  background: var(--color-surface);
  border: none;

  font-size: 14px;
  font-weight: 600;

  color: var(--color-primary);

  cursor: pointer;
}

.view-all:hover {
  background: var(--color-surface-muted);
}

/* Animation */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* Mobile */
@media (max-width: 768px) {
  .dropdown {
    width: 300px;
    right: -40px;
  }
}
</style>
