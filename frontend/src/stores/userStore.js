import { defineStore } from 'pinia'
import { usersService } from '@/services/userService'

const describeError = (error, fallback) =>
  error?.response?.data?.error || (error?.response ? fallback : 'No se pudo conectar con el backend.')

export const useUsersStore = defineStore('users', {
  state: () => ({
    users: [], loading: false, error: null,
    updatingId: null, actionError: null, actionSuccess: null,
    historyByUser: {}, historyLoadingId: null,
    roleHistoryByUser: {}, roleHistoryLoadingId: null
  }),
  actions: {
    async fetchUsers() {
      this.loading = true
      this.error = null
      try {
        this.users = await usersService.getAll()
      } catch (error) {
        this.users = []
        this.error = describeError(error, 'No se pudieron cargar los usuarios')
      } finally {
        this.loading = false
      }
    },
    async changeAccess(userId, blocked, reason) {
      if (this.updatingId !== null) return
      this.updatingId = userId
      this.actionError = null
      this.actionSuccess = null
      try {
        const result = await usersService.changeAccess(userId, { isBlocked: blocked, reason })
        this.users = this.users.map(u => u.id === result.user.id ? result.user : u)
        delete this.historyByUser[userId]
        this.actionSuccess = blocked ? 'Cuenta bloqueada correctamente.' : 'Cuenta desbloqueada correctamente.'
        return result
      } catch (error) {
        this.actionError = describeError(error, 'No se pudo actualizar el acceso.')
        throw error
      } finally {
        this.updatingId = null
      }
    },
    async loadAccessHistory(userId) {
      this.historyLoadingId = userId
      this.actionError = null
      try {
        const items = await usersService.getAccessHistory(userId)
        this.historyByUser[userId] = items
        return items
      } catch (error) {
        this.actionError = describeError(error, 'No se pudo consultar el historial.')
        throw error
      } finally {
        this.historyLoadingId = null
      }
    },
    async changeRole(userId, isAdmin, reason) {
      if (this.updatingId !== null) throw new Error('Hay otra operación en curso.')
      this.updatingId = userId
      this.actionError = null
      this.actionSuccess = null
      try {
        const result = await usersService.changeRole(userId, { isAdmin, reason })
        this.users = this.users.map(u => u.id === result.user.id ? result.user : u)
        delete this.roleHistoryByUser[userId]
        this.actionSuccess = isAdmin ? 'Rol administrador asignado.' : 'Rol administrador retirado.'
        return result
      } catch (error) {
        this.actionError = describeError(error, 'No se pudo cambiar el rol.')
        throw error
      } finally {
        this.updatingId = null
      }
    },
    async loadRoleHistory(userId) {
      this.roleHistoryLoadingId = userId
      this.actionError = null
      try {
        const items = await usersService.getRoleHistory(userId)
        this.roleHistoryByUser[userId] = items
        return items
      } catch (error) {
        this.actionError = describeError(error, 'No se pudo consultar el historial de roles.')
        throw error
      } finally {
        this.roleHistoryLoadingId = null
      }
    },
    clearMessages() {
      this.actionError = null
      this.actionSuccess = null
    }
  }
})
