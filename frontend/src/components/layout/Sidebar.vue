<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import SidebarItem from './SidebarItem.vue'
import { useAuthStore } from '@/stores/auth'
import {
  Home,
  Calendar,
  ClipboardList,
  Users,
  Settings,
  Menu,
  ShieldCheck,
  AlertTriangle,
  Building2,
  KeyRound,
  X
} from 'lucide-vue-next'
import { mvpFeatures } from '@/config/appScope'

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
          label: 'Reservas',
          icon: ClipboardList,
          to: '/reservations'
        },

        {
          label: 'Configuración',
          icon: Settings,
          to: '/settings'
        }
      ]
    }
  ]
  // ------------------------------------------------------------
// Reservas grupales MVP2.
// ------------------------------------------------------------
//
// El acceso para unirse mediante código solo se muestra cuando el scope
// habilita reservas grupales. MVP1 permanece sin esta funcionalidad.
if (mvpFeatures.groupReservations) {
  sections[0].items.splice(3, 0, {
    label: 'Unirse con código',
    icon: KeyRound,
    to: '/reservations/join'
  })
}
  if (mvpFeatures.workshops) {
    sections[0].items.splice(-1, 0, {
      label: 'Talleres',
      icon: Calendar,
      to: '/workshops'
    })
  }

  // La programación institucional no es exclusiva del administrador:
  // los MANAGER reciben sus unidades autorizadas desde el backend.
  if (mvpFeatures.institutionalActivityProgramming) {
    sections[0].items.splice(-1, 0, {
      label: 'Programación institucional',
      icon: Building2,
      to: '/institutional-activities'
    })
  }

  if (authStore.user?.isAdmin === true) {
    sections.push({
      section: 'ADMINISTRACIÓN',

      items: [
        {
          label: 'Usuarios',
          icon: Users,
          to: '/users'
        }
      ]
    })

    if (mvpFeatures.institutionalUnitAdministration) {
      sections.at(-1).items.push({
        label: 'Unidades institucionales',
        icon: Building2,
        to: '/admin/institutional-units'
      })
    }

    if (mvpFeatures.schedulingConflictAdministration) {
      sections.at(-1).items.push({
        label: 'Conflictos de programación',
        icon: AlertTriangle,
        to: '/admin/scheduling-conflicts'
      })
    }


    if (mvpFeatures.resourceAdministration) {
      sections.at(-1).items.unshift({
        label: 'Panel Administrativo',
        icon: ShieldCheck,
        to: '/admin'
      })
      sections.at(-1).items.push({
        label: 'Recursos',
        icon: Calendar,
        to: '/resources'
      })
    }

    if (mvpFeatures.reports) {
      sections.at(-1).items.push({
        label: 'Reportes',
        icon: ClipboardList,
        to: '/reports'
      })
    }
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

  <Teleport to="body">
    <div
      v-if="isOpen"
      class="overlay"
      @click="closeSidebar"
    />
  </Teleport>

  <Teleport to="body">
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

    </aside>
  </Teleport>

</template>

<style scoped>

/* HAMBURGER */
.hamburger {
  width: 42px;
  height: 42px;

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

  background: rgba(15, 23, 42, 0.24);

  z-index: 998;
}

/* SIDEBAR */
.sidebar {
  position: fixed;

  top: 0;
  left: -320px;

  width: 304px;
  height: 100vh;

  background: var(--color-sidebar);

  color: var(--color-sidebar-text);

  padding: var(--space-5);

  box-sizing: border-box;

  display: flex;
  flex-direction: column;

  transition: left 0.3s ease;

  z-index: 999;

  overflow-y: auto;

  border-right: 1px solid var(--color-sidebar-border);
  box-shadow: 16px 0 40px rgba(15, 23, 42, 0.16);
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
  width: 48px;
  height: 48px;

  border-radius: var(--radius-lg);

  background: var(--color-primary-soft);
  color: var(--color-primary);

  display: flex;
  align-items: center;
  justify-content: center;

  box-shadow: none;
}

.logo-text h1 {
  margin: 0;

  font-size: 17px;
  font-weight: 800;

  color: var(--color-text);
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

  background: var(--color-surface-muted);

  color: var(--color-text-muted);

  cursor: pointer;

  display: flex;
  align-items: center;
  justify-content: center;

  transition: 0.2s;
}

.close-btn:hover {
  background: var(--color-primary-soft);
  color: var(--color-primary);
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

/* SCROLL */
.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-thumb {
  background: #cbd5e1;

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
