<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

import {
  ChevronDown,
  User,
  Settings,
  Shield,
  LogOut
} from 'lucide-vue-next'

const open = ref(false)

const toggle = () => {
  open.value = !open.value
}

const close = (event) => {
  if (!event.target.closest('.user-menu')) {
    open.value = false
  }
}

onMounted(() => {
  window.addEventListener('click', close)
})

onBeforeUnmount(() => {
  window.removeEventListener('click', close)
})
</script>

<template>
  <div class="user-menu">

    <!-- Trigger -->
    <button
      class="user-trigger"
      @click.stop="toggle"
    >

      <img
        src="https://i.pravatar.cc/100"
        alt="user"
      />

      <div class="user-info">

        <strong>
          Nicolás Montaña
        </strong>

        <span>
          Estudiante
        </span>

      </div>

      <ChevronDown
        :size="16"
        class="arrow"
        :class="{ rotate: open }"
      />

    </button>

    <!-- Dropdown -->
    <transition name="fade">

      <div
        v-if="open"
        class="dropdown"
      >

        <!-- Top User -->
        <div class="profile-card">

          <img
            src="https://i.pravatar.cc/100"
            alt="profile"
          />

          <div>

            <strong>
              Nicolás Montaña
            </strong>

            <p>
              nicolas@ucen.cl
            </p>

          </div>

        </div>

        <div class="divider"></div>

        <!-- Menu Items -->
        <button class="dropdown-item">

          <User :size="16" />

          Mi perfil

        </button>

        <button class="dropdown-item">

          <Settings :size="16" />

          Configuración

        </button>

        <button class="dropdown-item">

          <Shield :size="16" />

          Privacidad

        </button>

        <div class="divider"></div>

        <button class="dropdown-item logout">

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

/* Trigger */
.user-trigger {
  background: transparent;
  border: none;

  display: flex;
  align-items: center;
  gap: 10px;

  cursor: pointer;

  padding: 4px 6px;

  border-radius: 14px;

  transition: 0.2s;
}

.user-trigger:hover {
  background: #f8fafc;
}

.user-trigger img {
  width: 42px;
  height: 42px;

  border-radius: 999px;

  object-fit: cover;
}

/* Info */
.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.user-info strong {
  font-size: 14px;
  color: #0f172a;
}

.user-info span {
  font-size: 12px;
  color: #64748b;
}

/* Arrow */
.arrow {
  transition: 0.2s;
}

.rotate {
  transform: rotate(180deg);
}

/* Dropdown */
.dropdown {
  position: absolute;

  top: 58px;
  right: 0;

  width: 260px;

  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  box-shadow:
    0 12px 32px rgba(0,0,0,0.08);

  overflow: hidden;

  z-index: 999;
}

/* Profile card */
.profile-card {
  display: flex;
  align-items: center;
  gap: 14px;

  padding: 18px;
}

.profile-card img {
  width: 54px;
  height: 54px;

  border-radius: 999px;

  object-fit: cover;
}

.profile-card strong {
  display: block;

  font-size: 15px;

  color: #0f172a;
}

.profile-card p {
  margin: 4px 0 0;

  font-size: 13px;

  color: #64748b;
}

/* Divider */
.divider {
  height: 1px;

  background: #e2e8f0;
}

/* Items */
.dropdown-item {
  width: 100%;

  background: transparent;
  border: none;

  display: flex;
  align-items: center;
  gap: 12px;

  padding: 14px 18px;

  font-size: 14px;

  cursor: pointer;

  transition: 0.2s;
}

.dropdown-item:hover {
  background: #f8fafc;
}

.logout {
  color: #ef4444;
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
  .user-info {
    display: none;
  }

  .dropdown {
    width: 220px;
  }
}
</style>