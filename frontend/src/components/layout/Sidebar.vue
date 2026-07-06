<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import SidebarItem from './SidebarItem.vue'
import { useAuthStore } from '@/stores/auth'
import {
  Home,
  Calendar,
  ClipboardList,
  History,
  LayoutGrid,
  Shield,
  Users,
  Settings,
  BarChart3,
  HelpCircle,
  LogOut,
  Menu,
  ShieldCheck,
  X
} from 'lucide-vue-next'

/* STATE */
const isOpen = ref(false)

const route = useRoute()
const authStore = useAuthStore()
/* AUTO CLOSE ON ROUTE CHANGE */
watch(
  () => route.path,

  () => {
    isOpen.value = false
  }
)

/* TOGGLE */
const toggleSidebar = () => {
  isOpen.value = !isOpen.value
}

/* CLOSE */
const closeSidebar = () => {
  isOpen.value = false
}

/* MENU */
const menu = computed(() => {
  const sections = [
    {
      section: 'MENÚ',

      items: [
        {
          label: 'Inicio',
          icon: Home,
          to: '/'
        },

        {
          label: 'Disponibilidad',
          icon: Calendar,
          to: '/availability'
        },

        {
          label: 'Mis Reservas',
          icon: ClipboardList,
          to: '/reservations'
        },

        {
          label: 'Historial',
          icon: History,
          to: '/history'
        },

        {
          label: 'Configuración',
          icon: Settings,
          to: '/settings'
        }
      ]
    }
  ]

  if (authStore.user?.isAdmin === true) {
    sections.push({
      section: 'ADMINISTRACIÓN',

      items: [
        {
          label: 'Panel Administrativo',
          icon: Shield,
          to: '/admin'
        },

        {
          label: 'Usuarios',
          icon: Users,
          to: '/users'
        },

        {
          label: 'Mis Recursos',
          icon: LayoutGrid,
          to: '/resources'
        },

        {
          label: 'Reportes',
          icon: BarChart3,
          to: '/reports'
        }
      ]
    })
  }

  return sections
})
</script>

<template>

  <!-- HAMBURGER -->
  <button
    class="hamburger"
    type="button"
    aria-label="Abrir menu"
    @click="toggleSidebar"
  >
    <Menu :size="23" />
  </button>

  <!-- OVERLAY -->
  <div
    v-if="isOpen"
    class="overlay"
    @click="closeSidebar"
  />

  <!-- SIDEBAR -->
  <aside
    class="sidebar"
    :class="{ open: isOpen }"
  >

    <!-- TOP -->
    <div class="top">

      <!-- LOGO -->
      <div class="logo">

        <div class="logo-icon">
          <ShieldCheck :size="25" />
        </div>

        <div class="logo-text">

          <h1>
            POLI REDI
          </h1>

          <span>
            Sistema de Reservas
          </span>

        </div>

      </div>

      <!-- CLOSE -->
      <button
        class="close-btn"
        type="button"
        aria-label="Cerrar menu"
        @click="closeSidebar"
      >
        <X :size="22" />
      </button>

    </div>

    <!-- NAV -->
    <div class="nav-sections">

      <div
        v-for="section in menu"
        :key="section.section"
        class="section"
      >

        <!-- TITLE -->
        <span class="section-title">
          {{ section.section }}
        </span>

        <!-- ITEMS -->
        <div class="items">

          <SidebarItem
            v-for="item in section.items"
            :key="item.label"

            :label="item.label"

            :icon="item.icon"

            :to="item.to"
          />

        </div>

      </div>

    </div>

    <!-- FOOTER -->
    <div class="footer">

      <SidebarItem
        label="Ayuda"
        :icon="HelpCircle"
        to="/"
      />

      <SidebarItem
        label="Cerrar sesión"
        :icon="LogOut"
        to="/"
      />

      <!-- SECURITY -->
      <div class="security-box">

        <div class="shield">
          <ShieldCheck :size="20" />
        </div>

        <div>

          <strong>
            Seguro y confiable
          </strong>

          <p>
            Tus datos están protegidos
            con altos estándares.
          </p>

        </div>

      </div>

    </div>

  </aside>

</template>

<style scoped>

/* HAMBURGER */
.hamburger {
  width: 44px;
  height: 44px;

  border: none;
  border-radius: var(--radius-md);

  background: var(--color-primary);

  color: var(--color-primary-contrast);

  cursor: pointer;

  display: flex;
  align-items: center;
  justify-content: center;

  font-size: 24px;

  flex-shrink: 0;

  transition: 0.2s;
}

.hamburger:hover {
  background: var(--color-primary-strong);
}

/* OVERLAY */
.overlay {
  position: fixed;
  inset: 0;

  background: rgba(15, 23, 42, 0.45);

  backdrop-filter: blur(2px);

  z-index: 998;
}

/* SIDEBAR */
.sidebar {
  position: fixed;

  top: 0;
  left: -320px;

  width: 320px;
  height: 100vh;

  background: var(--color-sidebar);

  color: white;

  padding: var(--space-6);

  box-sizing: border-box;

  display: flex;
  flex-direction: column;

  transition: left 0.3s ease;

  z-index: 999;

  overflow-y: auto;
}

.sidebar.open {
  left: 0;
}

/* TOP */
.top {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 20px;

  margin-bottom: 32px;
}

/* LOGO */
.logo {
  display: flex;
  align-items: center;

  gap: 14px;
}

.logo-icon {
  width: 52px;
  height: 52px;

  border-radius: var(--radius-lg);

  background: var(--color-primary);

  display: flex;
  align-items: center;
  justify-content: center;

  box-shadow: 0 10px 25px rgba(37, 99, 235, 0.28);
}

.logo-text h1 {
  margin: 0;

  font-size: 18px;
  font-weight: 800;

  color: white;
}

.logo-text span {
  font-size: 13px;

  color: var(--color-sidebar-muted);
}

/* CLOSE */
.close-btn {
  width: 42px;
  height: 42px;

  border: none;
  border-radius: var(--radius-md);

  background: rgba(255, 255, 255, 0.08);

  color: white;

  cursor: pointer;

  display: flex;
  align-items: center;
  justify-content: center;

  transition: 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.14);
}

/* NAV */
.nav-sections {
  flex: 1;

  display: flex;
  flex-direction: column;

  gap: 28px;
}

/* SECTION */
.section {
  display: flex;
  flex-direction: column;

  gap: 12px;
}

/* TITLE */
.section-title {
  font-size: 12px;
  font-weight: 700;

  letter-spacing: 0;

  color: var(--color-sidebar-muted);

  padding-left: 10px;
}

/* ITEMS */
.items {
  display: flex;
  flex-direction: column;

  gap: 6px;
}

/* FOOTER */
.footer {
  margin-top: 32px;

  display: flex;
  flex-direction: column;

  gap: 10px;
}

/* SECURITY */
.security-box {
  margin-top: 20px;

  padding: 18px;

  border-radius: var(--radius-lg);

  background: rgba(255, 255, 255, 0.05);

  border: 1px solid rgba(255, 255, 255, 0.06);

  display: flex;

  gap: 14px;
}

.shield {
  width: 42px;
  height: 42px;

  border-radius: var(--radius-md);

  background: rgba(37, 99, 235, 0.18);
  color: white;

  display: flex;
  align-items: center;
  justify-content: center;

  flex-shrink: 0;
}

.security-box strong {
  font-size: 14px;

  color: white;
}

.security-box p {
  margin-top: 4px;

  font-size: 12px;
  line-height: 1.5;

  color: var(--color-sidebar-muted);
}

/* SCROLL */
.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);

  border-radius: 999px;
}

/* MOBILE */
@media (max-width: 768px) {
  .sidebar {
    width: 290px;

    left: -290px;
  }
}
</style>
