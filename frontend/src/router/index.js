import {
  createRouter,
  createWebHistory
} from 'vue-router'

import { isAuthenticated } from '../auth/authService'

import DashboardView from '../views/DashboardView.vue'
import AvailabilityView from '../views/AvailabilityView.vue'
import ReservationsView from '../views/ReservationsView.vue'
import ReservationDetailView from '../views/ReservationDetailView.vue'
import HistoryView from '../views/HistoryView.vue'
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
    component: HistoryView,
    meta: { requiresAuth: true }
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

if (mvpFeatures.workshops) {
  routes.splice(fallbackIndex, 0, {
    path: '/workshops',
    component: () => import('../views/WorkshopsView.vue'),
    meta: { requiresAuth: true }
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
  if (to.meta.public) {
    return true
  }

  if (to.meta.requiresAuth) {
    const authenticated = await isAuthenticated()

    if (!authenticated) {
      return {
        path: '/login',
        query: {
          redirect: to.fullPath
        }
      }
    }

    const authStore = useAuthStore()
    const user = await authStore.loadAuthUser()

    if (!user && authStore.errorStatus === 403) {
      return { path: '/blocked' }
    }

    if (to.meta.requiresAdmin && user?.isAdmin !== true) {
      if (to.path === '/availability') {
        return true
      }

      return { path: '/availability' }
    }
  }

  return true
})

export default router
