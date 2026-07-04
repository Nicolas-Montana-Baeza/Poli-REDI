import {
  createRouter,
  createWebHistory
} from 'vue-router'

import { isAuthenticated, login } from '../auth/authService'

import DashboardView from '../views/DashboardView.vue'
import AvailabilityView from '../views/AvailabilityView.vue'
import ReservationsView from '../views/ReservationsView.vue'
import HistoryView from '../views/HistoryView.vue'
import ResourcesView from '../views/ResourcesView.vue'
import AdminView from '../views/AdminView.vue'
import UsersView from '../views/UsersView.vue'
import SettingsView from '../views/SettingsView.vue'
import ReportsView from '../views/ReportsView.vue'
import AuthCallbackView from '../views/AuthCallbackView.vue'
import { useAuthStore } from '../stores/auth'

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
    path: '/history',
    component: HistoryView,
    meta: { requiresAuth: true }
  },
  {
    path: '/resources',
    component: ResourcesView,
    meta: { requiresAuth: true }
  },
  {
    path: '/admin',
    component: AdminView,
    meta: { requiresAuth: true, requiresAdmin: true }
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
    path: '/reports',
    component: ReportsView,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/auth/callback',
    component: AuthCallbackView,
    meta: { public: true }
  }
]

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
      await login(to.fullPath)
      return false
    }

    const authStore = useAuthStore()
    const user = authStore.user || await authStore.loadAuthUser()

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
