import {
  createRouter,
  createWebHistory
} from 'vue-router'

import { getSafeRedirectPath } from '../auth/authService'

import DashboardView from '../views/DashboardView.vue'
import AvailabilityView from '../views/AvailabilityView.vue'
import ReservationsView from '../views/ReservationsView.vue'
import ReservationDetailView from '../views/ReservationDetailView.vue'
import UsersView from '../views/UsersView.vue'
import SettingsView from '../views/SettingsView.vue'
import LoginView from '../views/LoginView.vue'
import BlockedView from '../views/BlockedView.vue'
import { useAuthStore } from '../stores/auth'
import { mvpFeatures } from '../config/appScope'

const routes = [
  {
    path: '/',
    component: DashboardView,
    meta: { requiresAuth: true }
  },
  {
    path: '/availability',
    component: AvailabilityView,
    meta: { requiresAuth: true }
  },
  {
    path: '/reservations',
    component: ReservationsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/reservations/:id',
    component: ReservationDetailView,
    meta: { requiresAuth: true }
  },
  {
    path: '/history',
    redirect: {
      path: '/reservations',
      query: {
        tab: 'history'
      }
    }
  },
  {
    path: '/users',
    component: UsersView,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/settings',
    component: SettingsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    component: LoginView,
    meta: { public: true }
  },
  {
    path: '/blocked',
    component: BlockedView,
    meta: { public: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/availability'
  }
]

const fallbackIndex = routes.length - 1

// Las reservas grupales pertenecen al alcance incremental de MVP2.
//
// La ruta se registra únicamente cuando el frontend está ejecutándose
// con MVP2 o FULL, manteniendo MVP1 sin superficies todavía no habilitadas.
if (mvpFeatures.groupReservations) {
  routes.splice(routes.length - 1, 0, {
    path: '/reservations/join',
    component: () => import('../views/JoinReservationView.vue'),
    meta: { requiresAuth: true }
  })
}

if (mvpFeatures.workshops) {
  routes.splice(fallbackIndex, 0, {
    path: '/workshops',
    component: () => import('../views/WorkshopsView.vue'),
    meta: { requiresAuth: true }
  })
}

// La administración de unidades institucionales forma parte de MVP2.
//
// Esta superficie requiere administrador en frontend y backend.
// La autorización del navegador mejora UX, pero el backend continúa siendo
// la autoridad definitiva.
// Programación institucional accesible a administradores y MANAGER.
//
// No utilizamos requiresAdmin: la autorización real depende de la unidad
// concreta y es validada nuevamente por el backend.
if (mvpFeatures.institutionalActivityProgramming) {
  routes.splice(routes.length - 1, 0, {
    path: '/institutional-activities',
    component: () => import('../views/InstitutionalActivitiesView.vue'),
    meta: {
      requiresAuth: true
    }
  })
}

if (mvpFeatures.institutionalUnitAdministration) {
  routes.splice(routes.length - 1, 0, {
    path: '/admin/institutional-units',
    component: () => import('../views/InstitutionalUnitsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true
    }
  })
}

// La resolución de conflictos es una superficie administrativa MVP2.
// El backend sigue siendo la autoridad sobre la validez del plan final.
if (mvpFeatures.schedulingConflictAdministration) {
  routes.splice(routes.length - 1, 0, {
    path: '/admin/scheduling-conflicts',
    component: () => import('../views/SchedulingConflictsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true
    }
  })
}


if (mvpFeatures.resourceAdministration) {
  routes.splice(routes.length - 1, 0,
    {
      path: '/resources',
      component: () => import('../views/ResourcesView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin',
      component: () => import('../views/AdminView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    }
  )
}

if (mvpFeatures.reports) {
  routes.splice(routes.length - 1, 0, {
    path: '/reports',
    component: () => import('../views/ReportsView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  })
}

if (mvpFeatures.onlineAuth) {
  routes.splice(routes.length - 1, 0, {
    path: '/auth/callback',
    component: () => import('../views/AuthCallbackView.vue'),
    meta: { public: true }
  })
}

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  // El callback de Microsoft controla su propia pantalla
  // intermedia mientras MSAL procesa el redirect.
  if (
    to.path === '/auth/callback' ||
    to.path === '/blocked'
  ) {
    return true
  }

  // Login nunca debe renderizarse hasta saber con certeza
  // si ya existe una sesión.
  if (to.path === '/login') {
    // Si venimos explícitamente de un cierre de sesión,
    // no intentamos reconstruir la cuenta otra vez.
    //
    // Esto también cubre el nuevo arranque de Vue después
    // de logoutRedirect de Microsoft.
    if (authStore.loggingOut) {
      authStore.finishLogout()
      return true
    }

    const user =
      authStore.initialized
        ? authStore.user
        : await authStore.initializeSession()

    if (authStore.account && user) {
      return getSafeRedirectPath(
        to.query.redirect
      )
    }

    return true
  }

  if (to.meta.public) {
    return true
  }

  if (to.meta.requiresAuth) {
    const user =
      authStore.initialized
        ? authStore.user
        : await authStore.initializeSession()

    // Solo enviamos al login después de haber resuelto
    // definitivamente que no existe una cuenta.
    if (!authStore.account) {
      return {
        path: '/login',
        query: {
          redirect: to.fullPath
        }
      }
    }

    if (
      !user &&
      authStore.errorStatus === 403
    ) {
      return {
        path: '/blocked'
      }
    }

    if (!user) {
      return {
        path: '/login',
        query: {
          redirect: to.fullPath
        }
      }
    }

    if (
      to.meta.requiresAdmin &&
      user.isAdmin !== true
    ) {
      return {
        path: '/availability'
      }
    }
  }

  return true
})

export default router
