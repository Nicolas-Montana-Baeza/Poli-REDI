<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import ParticipantsProgress from '@/components/ui/ParticipantsProgress.vue'
import PrimaryButton from '@/components/ui/PrimaryButton.vue'
import { reservationsService } from '@/services/reservations.service'

const route = useRoute()
const code = ref(String(route.params.code || ''))
const activeCode = ref('')
const progress = ref(null)
const error = ref('')
const success = ref('')
const busy = ref(false)
const operation = ref('')
const codeInput = ref(null)
const errorElement = ref(null)
const liveAnnouncement = ref('')
let mounted = true
let requestGeneration = 0
const describedBy = computed(() => error.value ? 'join-code-help join-code-error' : 'join-code-help')

const messageFor = errorValue => {
  const status = errorValue?.response?.status
  if (!errorValue?.response) return 'No pudimos conectar con el servidor. Revisa tu conexión e intenta nuevamente.'
  if (status === 404) return 'El código no existe o ya no está disponible.'
  if (status === 403) return 'Debes tener una cuenta activa y RUT registrado para participar.'
  if (status === 410) return 'El plazo de confirmación ya venció.'
  if (status === 409) return 'No se pudo cambiar tu participación por el estado actual de la reserva.'
  return 'No se pudo completar la operación. Intenta nuevamente.'
}
const focusError = async () => {
  await nextTick()
  errorElement.value?.focus()
}
const progressAnnouncement = value => {
  const count = Number(value?.participantCount) || 0
  const participant = count === 1 ? 'participante confirmado' : 'participantes confirmados'
  return `Reserva encontrada. ${count} ${participant}; mínimo ${Number(value?.minimumParticipants) || 0}; objetivo ${Number(value?.targetParticipants) || 0}; capacidad ${Number(value?.capacity) || 0}.`
}
const validate = token => {
  if (!token) return 'Ingresa un código de invitación.'
  if (token.length < 8) return 'El código debe tener al menos 8 caracteres.'
  if (token.length > 512) return 'El código es demasiado largo.'
  return ''
}
const clearResultState = () => {
  requestGeneration += 1
  activeCode.value = ''
  progress.value = null
  error.value = ''
  success.value = ''
  liveAnnouncement.value = ''
  busy.value = false
  operation.value = ''
}
const load = async () => {
  const token = code.value.trim()
  const generation = ++requestGeneration
  error.value = validate(token)
  success.value = ''
  if (error.value) {
    progress.value = null
    await nextTick()
    codeInput.value?.focus()
    return
  }
  busy.value = true
  operation.value = 'consultar'
  try {
    const response = await reservationsService.getGroupProgress(token)
    if (!mounted || generation !== requestGeneration || code.value.trim() !== token) return
    activeCode.value = token
    progress.value = response
    liveAnnouncement.value = progressAnnouncement(response)
  } catch (errorValue) {
    if (!mounted || generation !== requestGeneration || code.value.trim() !== token) return
    activeCode.value = ''
    progress.value = null
    error.value = messageFor(errorValue)
    await focusError()
  } finally {
    if (mounted && generation === requestGeneration) {
      busy.value = false
      operation.value = ''
    }
  }
}
const change = async confirm => {
  if (!activeCode.value) return
  const requestedCode = activeCode.value
  const generation = ++requestGeneration
  busy.value = true
  operation.value = confirm ? 'confirmar' : 'retirar'
  error.value = ''
  success.value = ''
  try {
    const response = confirm
      ? await reservationsService.confirmGroup(requestedCode)
      : await reservationsService.withdrawGroup(requestedCode)
    if (!mounted || generation !== requestGeneration || activeCode.value !== requestedCode) return
    progress.value = response
    success.value = confirm ? 'Participación confirmada.' : 'Participación retirada.'
    liveAnnouncement.value = progressAnnouncement(response)
  } catch (errorValue) {
    if (!mounted || generation !== requestGeneration || activeCode.value !== requestedCode) return
    error.value = messageFor(errorValue)
    await focusError()
  } finally {
    if (mounted && generation === requestGeneration) {
      busy.value = false
      operation.value = ''
    }
  }
}
onMounted(() => {
  if (code.value) load()
})
watch(() => route.params.code, async routeCode => {
  if (!mounted) return
  clearResultState()
  code.value = String(routeCode || '')
  if (code.value) await load()
  else {
    await nextTick()
    codeInput.value?.focus()
  }
})
onBeforeUnmount(() => {
  mounted = false
  requestGeneration += 1
  code.value = ''
  activeCode.value = ''
  progress.value = null
  liveAnnouncement.value = ''
})
</script>

<template>
  <section class="join-page" aria-labelledby="join-title">
    <header class="app-section-header join-header">
      <h1 id="join-title">Unirse a una reserva grupal</h1>
      <p>Consulta la invitación y confirma tu participación antes del plazo indicado.</p>
    </header>
    <p class="sr-only" aria-live="polite" aria-atomic="true">{{ liveAnnouncement }}</p>

    <article class="app-card join-card">
      <form novalidate :aria-busy="busy" @submit.prevent="load">
        <div class="form-field">
          <label for="join-code">Código de invitación</label>
          <p id="join-code-help" class="field-help">Pega el código exactamente como lo recibiste. Se respetan mayúsculas y minúsculas.</p>
          <div class="join-controls">
            <input
              id="join-code"
              ref="codeInput"
              v-model="code"
              type="text"
              autocomplete="off"
              autocapitalize="off"
              spellcheck="false"
              :aria-describedby="describedBy"
              :aria-invalid="Boolean(error)"
              :disabled="busy"
            >
            <PrimaryButton type="submit" :loading="operation === 'consultar'" :disabled="busy">
              {{ operation === 'consultar' ? 'Consultando…' : 'Consultar invitación' }}
            </PrimaryButton>
          </div>
        </div>
      </form>

      <p v-if="!progress && !error && !busy" class="empty-state" role="status">
        Ingresa un código para consultar el estado de la invitación.
      </p>
      <p v-if="operation" class="operation-status" role="status">
        {{ operation === 'consultar' ? 'Consultando invitación…' : operation === 'confirmar' ? 'Confirmando participación…' : 'Retirando participación…' }}
      </p>
      <p v-if="error" id="join-code-error" ref="errorElement" class="state-card error" role="alert" tabindex="-1">{{ error }}</p>
      <p v-if="success" class="state-card success" role="status">{{ success }}</p>
    </article>

    <div v-if="progress" class="progress-card">
      <ParticipantsProgress :progress="progress" :busy="busy" :announce="false" @confirm="change(true)" @withdraw="change(false)" />
    </div>
  </section>
</template>

<style scoped>
.join-page{width:100%;max-width:760px;margin:0 auto;display:grid;gap:var(--space-5)}
.join-header h1,.join-header p{margin:0}.join-header{display:grid;gap:var(--space-2)}
.join-card{display:grid;gap:var(--space-4);padding:var(--space-5);min-width:0}
.form-field,.join-controls{min-width:0}.field-help{margin:0;color:var(--color-text-muted);font-size:var(--text-help);line-height:1.45}
.join-controls{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:var(--space-3);align-items:end}
.join-controls input{width:100%;min-width:0;min-height:44px;padding:0 var(--space-3);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;background:var(--color-surface)}
.join-controls :deep(.btn-ui){min-height:44px;white-space:nowrap}
.empty-state,.operation-status{margin:0;color:var(--color-text-muted);font-weight:650}.state-card{margin:0}.progress-card{min-width:0}
.sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}
@media(max-width:600px){.join-page{gap:var(--space-4)}.join-card{padding:var(--space-4)}.join-controls{grid-template-columns:minmax(0,1fr)}.join-controls :deep(.btn-ui){width:100%}}
@media(max-width:360px){.join-card{padding:var(--space-3)}}
</style>
