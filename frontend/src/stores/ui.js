import { defineStore } from 'pinia'

export const useUiStore =
  defineStore('ui', {

    state: () => ({

      sidebarOpen: false,

      loading: false,

      selectedDate: null
    }),

    actions: {

      toggleSidebar() {
        this.sidebarOpen =
          !this.sidebarOpen
      },

      openSidebar() {
        this.sidebarOpen = true
      },

      closeSidebar() {
        this.sidebarOpen = false
      },

      setLoading(value) {
        this.loading = value
      },

      setDate(date) {
        this.selectedDate = date
      }
    }
  })