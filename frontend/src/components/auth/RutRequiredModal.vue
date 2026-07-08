<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { IdCard } from 'lucide-vue-next'

import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { formatRutInput, isValidRut, normalizeRut } from '@/utils/validators'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const authStore = useAuthStore()
const notificationsStore = useNotificationsStore()
const router = useRouter()

const rutValue = ref('')
const rutError = ref('')
const saving = ref(false)

const handleRutInput = () => {
  rutValue.value = formatRutInput(rutValue.value)
  rutError.value = ''
}

const handleSubmit = async () => {
  rutError.value = ''

  const normalized = normalizeRut(rutValue.value)

  if (!normalized) {
    rutError.value = 'Ingresa tu RUT para continuar.'
    return
  }

  if (!isValidRut(normalized)) {
    rutError.value = 'El RUT ingresado no es válido.'
    return
  }

  saving.value = true

  try {
    await authStore.updateRut(normalized)
    rutValue.value = authStore.user?.rut || normalized
  } catch (error) {
    rutError.value = error.message
  } finally {
    saving.value = false
  }
}

const handleLogout = async () => {
  await authStore.logoutUser()
  notificationsStore.clearNotifications()
  await router.replace('/login')
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      rutValue.value = formatRutInput(authStore.user?.rut || rutValue.value)
      rutError.value = ''
    }
  }
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="rut-overlay"
    >
      <section
        class="rut-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="rut-required-title"
      >
        <div class="icon-wrap">
          <IdCard :size="26" />
        </div>

        <header>
          <h2 id="rut-required-title">
            Registra tu RUT
          </h2>

          <p>
            Necesitamos este dato para crear y asociar tus reservas.
          </p>
        </header>

        <form
          class="rut-form"
          @submit.prevent="handleSubmit"
        >
          <label for="requiredRut">
            RUT
          </label>

          <input
            id="requiredRut"
            v-model="rutValue"
            type="text"
            inputmode="text"
            autocomplete="off"
            placeholder="12345678-5"
            :disabled="saving"
            @input="handleRutInput"
          />

          <p
            v-if="rutError"
            class="rut-error"
          >
            {{ rutError }}
          </p>

          <button
            class="primary-action"
            type="submit"
            :disabled="saving"
          >
            {{ saving ? 'Guardando...' : 'Guardar y continuar' }}
          </button>
        </form>

        <button
          class="secondary-action"
          type="button"
          :disabled="saving"
          @click="handleLogout"
        >
          Cerrar sesión
        </button>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.rut-overlay {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: grid;
  place-items: center;
  padding: 20px;
  background: rgba(15, 23, 42, 0.58);
  backdrop-filter: blur(8px);
}

.rut-modal {
  width: min(100%, 430px);
  border-radius: 24px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.24);
  padding: 28px;
}

.icon-wrap {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  border-radius: 18px;
  background: #e0ecff;
  color: #2457d6;
  margin-bottom: 18px;
}

header h2 {
  margin: 0;
  color: #0f172a;
  font-size: 1.55rem;
}

header p {
  margin: 8px 0 22px;
  color: #64748b;
  line-height: 1.45;
}

.rut-form {
  display: grid;
  gap: 10px;
}

.rut-form label {
  color: #334155;
  font-weight: 800;
}

.rut-form input {
  width: 100%;
  min-height: 50px;
  border: 1px solid #cbd5e1;
  border-radius: 14px;
  padding: 0 14px;
  color: #0f172a;
  font: inherit;
  font-weight: 700;
  outline: none;
}

.rut-form input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

.rut-error {
  margin: 0;
  border-radius: 12px;
  background: #fee2e2;
  color: #b91c1c;
  padding: 10px 12px;
  font-size: 0.9rem;
  font-weight: 800;
}

.primary-action,
.secondary-action {
  width: 100%;
  min-height: 50px;
  border: 0;
  border-radius: 14px;
  font: inherit;
  font-weight: 900;
  cursor: pointer;
}

.primary-action {
  margin-top: 8px;
  background: #2563eb;
  color: #ffffff;
}

.secondary-action {
  margin-top: 12px;
  background: #f1f5f9;
  color: #334155;
}

.primary-action:disabled,
.secondary-action:disabled,
.rut-form input:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

@media (max-width: 520px) {
  .rut-overlay {
    align-items: end;
    padding: 12px;
  }

  .rut-modal {
    border-radius: 22px;
    padding: 22px;
  }
}
</style>
