<script setup>
import { computed, onMounted, ref } from 'vue'
import { ShieldCheck, UserRound, Users, Search, History } from 'lucide-vue-next'
import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useUsersStore } from '@/stores/userStore'
import { useAuthStore } from '@/stores/auth'

const usersStore = useUsersStore()
const authStore = useAuthStore()
const search = ref('')
const filter = ref('ALL')
const editingId = ref(null)
const roleEditingId = ref(null)
const historyId = ref(null)
const reason = ref('')
const roleReason = ref('')

onMounted(() => usersStore.fetchUsers())

const admins = computed(() => usersStore.users.filter(u => u.isAdmin))
const blocked = computed(() => usersStore.users.filter(u => u.isBlocked))
const active = computed(() => usersStore.users.filter(u => !u.isBlocked))
const stats = computed(() => [
  { label: 'Usuarios', value: usersStore.users.length, icon: Users },
  { label: 'Administradores', value: admins.value.length, icon: ShieldCheck },
  { label: 'Activos', value: active.value.length, icon: UserRound },
  { label: 'Bloqueados', value: blocked.value.length, icon: UserRound }
])
const normalize = text => String(text || '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()
const visibleUsers = computed(() => usersStore.users.filter(user => {
  if (filter.value === 'BLOCKED' && !user.isBlocked) return false
  if (filter.value === 'ACTIVE' && user.isBlocked) return false
  if (filter.value === 'ADMIN' && !user.isAdmin) return false
  const q = normalize(search.value.trim())
  return !q || [user.fullName, user.email, user.rut].some(value => normalize(value).includes(q))
}))
const selectedUser = computed(() => usersStore.users.find(u => u.id === editingId.value))
const selectedRoleUser = computed(() => usersStore.users.find(u => u.id === roleEditingId.value))
const initials = user => (user.fullName || user.email || 'Usuario').split(' ')
  .filter(Boolean).slice(0, 2).map(part => part[0]).join('').toUpperCase() || 'U'
const isSelf = user => Number(authStore.user?.id) === Number(user.id)
const dateLabel = value => value ? new Date(value).toLocaleString('es-CL') : ''

const beginChange = user => {
  if (isSelf(user)) return
  editingId.value = user.id
  roleEditingId.value = null
  historyId.value = null
  reason.value = ''
  usersStore.clearMessages()
}
const cancelChange = () => {
  editingId.value = null
  reason.value = ''
}
const saveChange = async () => {
  const user = selectedUser.value
  if (!user || usersStore.updatingId !== null) return
  try {
    await usersStore.changeAccess(user.id, !user.isBlocked, reason.value.trim())
    cancelChange()
  } catch {
    // El store mantiene el error visible y conserva el formulario.
  }
}
const beginRoleChange = user => {
  if (isSelf(user) || (user.isBlocked && !user.isAdmin)) return
  roleEditingId.value = user.id
  editingId.value = null
  historyId.value = null
  roleReason.value = ''
  usersStore.clearMessages()
}
const cancelRoleChange = () => {
  roleEditingId.value = null
  roleReason.value = ''
}
const saveRoleChange = async () => {
  const user = selectedRoleUser.value
  if (!user || usersStore.updatingId !== null) return
  try {
    await usersStore.changeRole(user.id, !user.isAdmin, roleReason.value.trim())
    cancelRoleChange()
  } catch {
    // El store mantiene el motivo y presenta el error.
  }
}
const toggleHistory = async user => {
  usersStore.clearMessages()
  editingId.value = null
  roleEditingId.value = null
  if (historyId.value === user.id) {
    historyId.value = null
    return
  }
  historyId.value = user.id
  try {
    await Promise.all([
      usersStore.loadAccessHistory(user.id),
      usersStore.loadRoleHistory(user.id)
    ])
  } catch {
    // Error visible mediante el store.
  }
}
</script>

<template>
  <main class="users-view">
    <header class="page-header">
      <div>
        <h1>Gestión de usuarios</h1>
        <p>Consulta las cuentas, administra su acceso y revisa las decisiones registradas.</p>
      </div>
    </header>

    <div v-if="usersStore.loading" aria-label="Cargando usuarios">
      <SkeletonLoader variant="reservations" :items="4" />
    </div>
    <template v-else>
      <p v-if="usersStore.error" class="feedback error" role="alert">{{ usersStore.error }}</p>
      <p v-if="usersStore.actionError" class="feedback error" role="alert">{{ usersStore.actionError }}</p>
      <p v-if="usersStore.actionSuccess" class="feedback success" role="status">{{ usersStore.actionSuccess }}</p>

      <section class="stats-grid" aria-label="Indicadores de usuarios">
        <article v-for="stat in stats" :key="stat.label" class="stat-card">
          <component :is="stat.icon" :size="20" aria-hidden="true" />
          <span>{{ stat.label }}</span>
          <strong>{{ stat.value }}</strong>
        </article>
      </section>

      <div class="toolbar">
        <label class="search-box">
          <Search :size="18" aria-hidden="true" />
          <span class="sr-only">Buscar usuarios</span>
          <input v-model="search" type="search" placeholder="Buscar por nombre, correo o RUT" />
        </label>
        <label class="filter-box">
          <span class="sr-only">Filtrar usuarios</span>
          <select v-model="filter">
            <option value="ALL">Todos los usuarios</option>
            <option value="ACTIVE">Activos</option>
            <option value="BLOCKED">Bloqueados</option>
            <option value="ADMIN">Administradores</option>
          </select>
        </label>
      </div>

      <p v-if="!visibleUsers.length" class="empty">No se encontraron usuarios con estos criterios.</p>
      <section v-else class="users-list" aria-label="Usuarios registrados">
        <article v-for="user in visibleUsers" :key="user.id" class="user-card">
          <div class="avatar" aria-hidden="true">{{ initials(user) }}</div>
          <div class="user-main">
            <h2>{{ user.fullName || 'Usuario' }}</h2>
            <p>{{ user.email }}</p>
            <small v-if="isSelf(user)">Tu cuenta</small>
          </div>
          <div class="badges">
            <span class="badge" :class="{ admin: user.isAdmin }">{{ user.isAdmin ? 'Administrador' : 'Usuario' }}</span>
            <span class="badge" :class="{ blocked: user.isBlocked }">{{ user.isBlocked ? 'Bloqueado' : 'Activo' }}</span>
          </div>
          <div class="actions">
            <button type="button" class="secondary-btn" :aria-expanded="historyId === user.id" @click="toggleHistory(user)">
              <History :size="15" aria-hidden="true" /> Historial
            </button>
            <button type="button" class="secondary-btn" :class="{ danger: !user.isBlocked }"
              :disabled="isSelf(user) || usersStore.updatingId !== null" :title="isSelf(user) ? 'No puedes cambiar el estado de tu propia cuenta' : ''"
              @click="beginChange(user)">
              {{ user.isBlocked ? 'Desbloquear' : 'Bloquear' }}
            </button>
            <button type="button" class="secondary-btn"
              :disabled="isSelf(user) || usersStore.updatingId !== null || (user.isBlocked && !user.isAdmin)"
              :title="isSelf(user) ? 'No puedes modificar tu propio rol' : (user.isBlocked && !user.isAdmin) ? 'Desbloquea esta cuenta antes de promoverla' : ''"
              @click="beginRoleChange(user)">
              {{ user.isAdmin ? 'Retirar rol admin' : 'Asignar rol admin' }}
            </button>
          </div>

          <form v-if="editingId === user.id" class="decision-form" @submit.prevent="saveChange">
            <strong>{{ user.isBlocked ? 'Desbloquear' : 'Bloquear' }} a {{ user.fullName }} </strong>
            <p>Este cambio afecta inmediatamente el acceso a la aplicación y quedará en auditoría.</p>
            <label :for="`reason-${user.id}`">Motivo de la decisión (obligatorio)</label>
            <textarea :id="`reason-${user.id}`" v-model="reason" minlength="5" maxlength="500" rows="3" required placeholder="Describe el motivo administrativo" />
            <div class="form-actions">
              <button type="submit" class="primary-btn" :disabled="reason.trim().length < 5 || usersStore.updatingId !== null">
                {{ usersStore.updatingId === user.id ? 'Guardando…' : 'Confirmar cambio' }}
              </button>
              <button type="button" class="secondary-btn" :disabled="usersStore.updatingId !== null" @click="cancelChange">Cancelar</button>
            </div>
          </form>

          <form v-if="roleEditingId === user.id" class="decision-form" @submit.prevent="saveRoleChange">
            <strong>{{ user.isAdmin ? 'Retirar rol administrador de' : 'Asignar rol administrador a' }} {{ user.fullName }}</strong>
            <p>Los privilegios administrativos cambian de inmediato. Se conservará una auditoría con responsable y motivo.</p>
            <label :for="`role-reason-${user.id}`">Motivo de la decisión (obligatorio)</label>
            <textarea :id="`role-reason-${user.id}`" v-model="roleReason" minlength="5" maxlength="500" rows="3" required placeholder="Justificación de la modificación del rol" />
            <div class="form-actions">
              <button type="submit" class="primary-btn" :disabled="roleReason.trim().length < 5 || usersStore.updatingId !== null">
                {{ usersStore.updatingId === user.id ? 'Guardando…' : 'Confirmar cambio de rol' }}
              </button>
              <button type="button" class="secondary-btn" :disabled="usersStore.updatingId !== null" @click="cancelRoleChange">Cancelar</button>
            </div>
          </form>

          <div v-if="historyId === user.id" class="history-panel">
            <strong>Historial de acceso</strong>
            <p v-if="usersStore.historyLoadingId === user.id" role="status">Cargando historial…</p>
            <template v-else-if="usersStore.historyByUser[user.id]">
              <p v-if="!usersStore.historyByUser[user.id].length">Sin cambios registrados.</p>
              <ol v-else>
                <li v-for="entry in usersStore.historyByUser[user.id]" :key="entry.id">
                  <strong>{{ entry.newIsBlocked ? 'Bloqueo' : 'Desbloqueo' }}</strong>
                  <span>{{ dateLabel(entry.changedAt) }} · {{ entry.actorName }}</span>
                  <p>{{ entry.reason }}</p>
                </li>
              </ol>
            </template>
            <strong>Historial de roles</strong>
            <p v-if="usersStore.roleHistoryLoadingId === user.id" role="status">Cargando roles…</p>
            <template v-else-if="usersStore.roleHistoryByUser[user.id]">
              <p v-if="!usersStore.roleHistoryByUser[user.id].length">Sin modificaciones de rol registradas.</p>
              <ol v-else>
                <li v-for="entry in usersStore.roleHistoryByUser[user.id]" :key="`role-${entry.id}`">
                  <strong>{{ entry.newIsAdmin ? 'Asignación de rol administrador' : 'Retiro de rol administrador' }}</strong>
                  <span>{{ dateLabel(entry.changedAt) }} · {{ entry.actorName }}</span>
                  <p>{{ entry.reason }}</p>
                </li>
              </ol>
            </template>
          </div>
        </article>
      </section>
    </template>
  </main>
</template>

<style scoped>
.users-view { display:grid;gap:1.2rem;padding:1.5rem;color:var(--color-text); }
.page-header h1 { font-size:1.55rem;font-weight:850;margin:0; }
.page-header p { color:var(--color-text-muted);margin:.35rem 0 0; }
.stats-grid { display:grid;grid-template-columns:repeat(auto-fit,minmax(145px,1fr));gap:.75rem; }
.stat-card { display:grid;gap:.3rem;border:1px solid var(--color-border);border-radius:.8rem;padding:1rem;background:var(--color-surface); }
.stat-card svg { color:var(--color-primary); }
.stat-card span { font-size:.8rem;color:var(--color-text-muted); }
.stat-card strong { font-size:1.7rem; }
.toolbar { display:flex;flex-wrap:wrap;gap:.75rem; }
.search-box,.filter-box { display:flex;align-items:center;gap:.4rem;border:1px solid var(--color-border);border-radius:.6rem;padding:.4rem .65rem;background:var(--color-surface); }
.search-box { flex:1;min-width:230px; }
input,select,textarea { color:var(--color-text);background:var(--color-surface);font:inherit; }
.search-box input { border:0;outline:0;width:100%; }
.filter-box select {border:0;min-height:2rem;}
.users-list {display:grid;gap:.75rem;}
.user-card {border:1px solid var(--color-border);border-radius:.8rem;background:var(--color-surface);padding:1rem;display:grid;grid-template-columns:auto minmax(0,1fr) auto auto;align-items:center;gap:.9rem;}
.avatar {width:46px;height:46px;border-radius:50%;display:grid;place-items:center;background:var(--color-primary-soft);color:var(--color-primary-strong);font-weight:800;}
.user-main {min-width:0;}
.user-main h2 {font-size:1rem;margin:0;overflow-wrap:anywhere;}
.user-main p {font-size:.84rem;color:var(--color-text-muted);margin:.25rem 0;overflow-wrap:anywhere;}
.user-main small {color:var(--color-text-muted);}
.badges,.actions,.form-actions {display:flex;flex-wrap:wrap;gap:.45rem;align-items:center;}
.badge {border-radius:30px;padding:.28rem .5rem;font-size:.75rem;background:var(--color-surface-soft);border:1px solid var(--color-border);}
.badge.admin {color:var(--color-primary-strong);}
.badge.blocked {color:var(--color-error);border-color:var(--color-error-border);background:var(--color-error-soft);}
button {font:inherit;cursor:pointer;}
button:disabled {opacity:.45;cursor:not-allowed;}
.secondary-btn,.primary-btn {display:inline-flex;align-items:center;gap:.3rem;padding:.52rem .7rem;border:1px solid var(--color-border);border-radius:.5rem;background:var(--color-surface);color:var(--color-text);font-size:.84rem;font-weight:650;}
.secondary-btn.danger {border-color:var(--color-error-border);color:var(--color-error);}
.primary-btn {background:var(--color-primary);color:white;border-color:var(--color-primary);}
.decision-form,.history-panel {grid-column:1/-1;display:grid;gap:.55rem;padding:1rem;border-radius:.55rem;border:1px solid var(--color-border);background:var(--color-surface-soft);}
.decision-form p,.history-panel p {margin:0;color:var(--color-text-muted);font-size:.85rem;}
.decision-form textarea {border:1px solid var(--color-border);padding:.65rem;border-radius:.4rem;max-width:100%;resize:vertical;}
.history-panel ol {display:grid;gap:.6rem;margin:.1rem 0;padding-left:1.35rem;}
.history-panel li span {display:block;color:var(--color-text-muted);font-size:.79rem;}
.history-panel li p {margin:.15rem 0;overflow-wrap:anywhere;}
.feedback {padding:.8rem 1rem;border-radius:.6rem;margin:0;}
.feedback.error {color:var(--color-error);background:var(--color-error-soft);}
.feedback.success {color:var(--color-success);background:var(--color-success-soft);}
.empty {text-align:center;color:var(--color-text-muted);padding:2rem;}
.sr-only {position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0;}
@media(max-width:950px){.user-card{grid-template-columns:auto minmax(0,1fr) auto}.actions{grid-column:2/-1}.badges{grid-column:3}}
@media(max-width:570px){.users-view{padding:.75rem}.user-card{grid-template-columns:auto 1fr}.badges,.actions{grid-column:1/-1}}
</style>
