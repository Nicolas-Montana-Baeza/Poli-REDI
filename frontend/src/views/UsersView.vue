<script setup>
import { computed, onMounted } from 'vue'

import {
  ShieldCheck,
  UserRound,
  Users
} from 'lucide-vue-next'

import SkeletonLoader from '@/components/ui/SkeletonLoader.vue'
import { useUsersStore } from '@/stores/userStore'

const usersStore = useUsersStore()

onMounted(() => {
  usersStore.fetchUsers()
})

const admins = computed(() => {
  return usersStore.users.filter((user) => user.isAdmin)
})

const blocked = computed(() => {
  return usersStore.users.filter((user) => user.isBlocked)
})

const active = computed(() => {
  return usersStore.users.filter((user) => !user.isBlocked)
})

const stats = computed(() => [
  {
    label: 'Usuarios',
    value: usersStore.users.length,
    icon: Users
  },
  {
    label: 'Administradores',
    value: admins.value.length,
    icon: ShieldCheck
  },
  {
    label: 'Activos',
    value: active.value.length,
    icon: UserRound
  }
])

const initials = (user) => {
  return (user.fullName || user.email || 'Usuario')
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'U'
}
</script>

<template>
  <main class="users-view">

    <header class="page-header">

      <h1>
        Usuarios
      </h1>

      <p>
        Cuentas registradas en Poli-REDI.
      </p>

    </header>

    <div
      v-if="usersStore.loading"
      aria-label="Cargando usuarios"
    >
      <SkeletonLoader
        variant="reservations"
        :items="5"
      />
    </div>

    <div
      v-else-if="usersStore.error"
      class="state-card error"
    >
      {{ usersStore.error }}
    </div>

    <section
      v-else
      class="content"
    >

      <section class="stats-grid">

        <article
          v-for="stat in stats"
          :key="stat.label"
          class="stat-card"
        >

          <component
            :is="stat.icon"
            :size="25"
          />

          <span>
            {{ stat.label }}
          </span>

          <strong>
            {{ stat.value }}
          </strong>

        </article>

        <article class="stat-card danger">

          <UserRound :size="25" />

          <span>
            Bloqueados
          </span>

          <strong>
            {{ blocked.length }}
          </strong>

        </article>

      </section>

      <div
        v-if="!usersStore.users.length"
        class="state-card"
      >
        No hay usuarios registrados.
      </div>

      <section
        v-else
        class="users-list"
      >

        <article
          v-for="user in usersStore.users"
          :key="user.id"
          class="user-card"
        >

          <span class="avatar">
            {{ initials(user) }}
          </span>

          <div class="user-main">

            <h2>
              {{ user.fullName || 'Usuario' }}
            </h2>

            <p>
              {{ user.email }}
            </p>

          </div>

          <div class="badges">

            <span
              class="badge"
              :class="{ admin: user.isAdmin }"
            >
              {{ user.isAdmin ? 'Administrador' : 'Usuario' }}
            </span>

            <span
              class="badge"
              :class="{ blocked: user.isBlocked }"
            >
              {{ user.isBlocked ? 'Bloqueado' : 'Activo' }}
            </span>

          </div>

        </article>

      </section>

    </section>

  </main>
</template>

<style scoped>
.users-view {
  width: 100%;

  display: flex;
  flex-direction: column;

  gap: 24px;
}

.page-header h1 {
  margin: 0;

  color: #0f172a;

  font-size: 30px;
  font-weight: 800;
}

.page-header p {
  margin-top: 8px;

  color: #64748b;
}

.content,
.users-list {
  display: flex;
  flex-direction: column;

  gap: 14px;
}

.state-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  padding: 22px;

  color: #334155;

  font-weight: 700;
}

.state-card.error {
  background: #fee2e2;

  color: #b91c1c;

  border-color: #fecaca;
}

.stats-grid {
  display: grid;

  grid-template-columns:
    repeat(auto-fit, minmax(190px, 1fr));

  gap: 18px;
}

.stat-card,
.user-card {
  background: white;

  border: 1px solid #e2e8f0;
  border-radius: 18px;

  box-shadow:
    0 4px 12px rgba(0,0,0,0.04);
}

.stat-card {
  padding: 20px;

  display: flex;
  flex-direction: column;

  gap: 9px;
}

.stat-card svg {
  color: #1d4ed8;
}

.stat-card.danger svg {
  color: #b91c1c;
}

.stat-card span {
  color: #64748b;

  font-size: 13px;
  font-weight: 800;
}

.stat-card strong {
  color: #0f172a;

  font-size: 34px;
  font-weight: 900;
}

.user-card {
  padding: 18px;

  display: grid;

  grid-template-columns:
    auto
    minmax(0, 1fr)
    auto;

  align-items: center;
  gap: 16px;
}

.avatar {
  width: 48px;
  height: 48px;

  border-radius: 999px;

  background:
    linear-gradient(
      135deg,
      #1e3a8a,
      #2563eb
    );

  color: white;

  display: flex;
  align-items: center;
  justify-content: center;

  font-weight: 900;
}

.user-main {
  min-width: 0;
}

.user-main h2,
.user-main p {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-main h2 {
  margin: 0;

  color: #0f172a;

  font-size: 17px;
  font-weight: 800;
}

.user-main p {
  margin: 5px 0 0;

  color: #64748b;
}

.badges {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;

  gap: 8px;
}

.badge {
  border-radius: 999px;

  background: #f1f5f9;

  color: #475569;

  padding: 7px 10px;

  font-size: 12px;
  font-weight: 800;
}

.badge.admin {
  background: #eff6ff;

  color: #1d4ed8;
}

.badge.blocked {
  background: #fee2e2;

  color: #b91c1c;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 26px;
  }

  .user-card {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .badges {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }
}
</style>
