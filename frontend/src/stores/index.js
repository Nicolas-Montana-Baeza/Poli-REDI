import {
  createRouter,
  createWebHistory
} from 'vue-router'

import DashboardView from
'../views/DashboardView.vue'

import AvailabilityView from
'../views/AvailabilityView.vue'

import ReservationsView from
'../views/ReservationsView.vue'

import HistoryView from
'../views/HistoryView.vue'

import ResourcesView from
'../views/ResourcesView.vue'

import AdminView from
'../views/AdminView.vue'

import UsersView from
'../views/UsersView.vue'

import SettingsView from
'../views/SettingsView.vue'

import ReportsView from
'../views/ReportsView.vue'

const routes = [
  {
    path: '/',
    component: DashboardView
  },

  {
    path: '/availability',
    component: AvailabilityView
  },

  {
    path: '/reservations',
    component: ReservationsView
  },

  {
    path: '/history',
    component: HistoryView
  },

  {
    path: '/resources',
    component: ResourcesView
  },

  {
    path: '/admin',
    component: AdminView
  },

  {
    path: '/users',
    component: UsersView
  },

  {
    path: '/settings',
    component: SettingsView
  },

  {
    path: '/reports',
    component: ReportsView
  }
]

export default createRouter({
  history: createWebHistory(),

  routes
})
