import { defineStore } from 'pinia'

import { usersService } from '@/services/userService'

export const useUsersStore = defineStore('users', {
  state: () => ({
    users: [],
    loading: false,
    error: null
  }),

  actions: {
    async fetchUsers() {
      this.loading = true
      this.error = null

      try {
        this.users = await usersService.getAll()
      } catch (error) {
        this.users = []
        this.error =
          error.response?.data?.error ||
          'No se pudieron cargar los usuarios'
      } finally {
        this.loading = false
      }
    }
  }
})
