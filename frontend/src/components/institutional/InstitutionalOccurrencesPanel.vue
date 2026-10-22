<script setup>
import { computed, reactive, ref } from 'vue'
import { institutionalActivitiesService } from '../../services/institutionalActivities.service'

const props = defineProps({ activity: { type: Object, required: true } })
const emit = defineEmits(['changed'])
const expanded = ref(false)
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')
const exceptions = ref([])
const selected = ref(null)
const form = reactive({ reason: '', newDate: '', newStartTime: '', newEndTime: '' })

const todayKey = () => {
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'America/Santiago', year: 'numeric', month: '2-digit', day: '2-digit'
  }).formatToParts(new Date())
  const get = name => parts.find(item => item.type === name)?.value || ''
  return `${get('year')}-${get('month')}-${get('day')}`
}
const fromKey = value => new Date(`${value}T12:00:00Z`)
const keyOf = value => value.toISOString().slice(0, 10)
const exceptionsMap = computed(() => new Map(exceptions.value.map(e => [`${e.scheduleId}-${e.originalDate}`, e])))

const sessions = computed(() => {
  const result = []
  const today = todayKey()
  for (const s of props.activity.schedules || []) {
    if (s.scheduleType !== 'WEEKLY' || !s.isActive || !s.validFrom || !s.validTo) continue
    const start = fromKey(s.validFrom)
    const end = fromKey(s.validTo)
    const weekday = Number(s.dayOfWeek)
    for (let date = start; date <= end && result.length < 16; date = new Date(date.getTime() + 86400000)) {
      const isoDay = date.getUTCDay() || 7
      if (isoDay !== weekday) continue
      const originalDate = keyOf(date)
      const exception = exceptionsMap.value.get(`${s.id}-${originalDate}`)
      const effectiveDate = exception?.action === 'RESCHEDULE' ? exception.newDate : originalDate
      if (effectiveDate < today) continue
      result.push({ scheduleId: s.id, originalDate, startTime: s.startTime, endTime: s.endTime, exception })
    }
  }
  return result.sort((a, b) =>
    String(a.exception?.action === 'RESCHEDULE' ? a.exception.newDate : a.originalDate)
      .localeCompare(String(b.exception?.action === 'RESCHEDULE' ? b.exception.newDate : b.originalDate))
  ).slice(0, 16)
})

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await institutionalActivitiesService.getExceptions(props.activity.id)
    exceptions.value = Array.isArray(data) ? data : []
  } catch (e) {
    error.value = e?.response?.data?.error || 'No se pudieron consultar las excepciones.'
  } finally { loading.value = false }
}

const toggle = async () => {
  expanded.value = !expanded.value
  selected.value = null
  if (expanded.value) await load()
}
const begin = (session, action) => {
  selected.value = { scheduleId: session.scheduleId, originalDate: session.originalDate, action }
  form.reason = ''
  form.newDate = session.exception?.action === 'RESCHEDULE' ? session.exception.newDate : session.originalDate
  form.newStartTime = session.exception?.action === 'RESCHEDULE'
    ? session.exception.newStartTime : String(session.startTime).slice(0, 5)
  form.newEndTime = session.exception?.action === 'RESCHEDULE'
    ? session.exception.newEndTime : String(session.endTime).slice(0, 5)
  error.value = ''
  success.value = ''
}
const submit = async () => {
  if (!selected.value || saving.value) return
  error.value = ''
  success.value = ''
  if (form.reason.trim().length < 5) { error.value = 'Indica un motivo de al menos cinco caracteres.'; return }
  if (selected.value.action === 'RESCHEDULE' && (!form.newDate || !form.newStartTime || !form.newEndTime || form.newEndTime <= form.newStartTime)) {
    error.value = 'La nueva fecha y el horario deben ser válidos.'; return
  }
  const payload = { ...selected.value, reason: form.reason.trim() }
  if (selected.value.action === 'RESCHEDULE') {
    Object.assign(payload, {newDate: form.newDate,newStartTime:form.newStartTime,newEndTime:form.newEndTime})
  }
  saving.value = true
  try {
    await institutionalActivitiesService.patchOccurrence(props.activity.id, payload)
    const verb = payload.action === 'CANCEL' ? 'suspendida' : 'reprogramada'
    success.value = `Sesión del ${payload.originalDate} ${verb}.`
    selected.value = null
    await load()
    emit('changed')
  } catch (e) {error.value = e?.response?.data?.error || 'No se pudo gestionar la sesión.'}
  finally {saving.value = false}
}
</script>

<template>
  <section class="occurrence-panel">
    <button type="button" class="toggle" :aria-expanded="expanded" @click="toggle">
      {{ expanded ? 'Ocultar sesiones' : 'Gestionar sesiones semanales' }}
    </button>
    <div v-if="expanded" class="inner">
      <p>Suspende o reprograma una fecha sin modificar el resto del taller.</p>
      <p v-if="loading" role="status">Cargando sesiones...</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="success" class="success" role="status">{{ success }}</p>
      <p v-if="Number(activity.enrollmentCount) > 0" class="notice">
        Esta actividad tiene inscripciones vigentes: no se pueden alterar sesiones sin resolver primero sus inscripciones y avisos.
      </p>
      <p v-if="!loading && !sessions.length">No quedan sesiones futuras de esta programación.</p>
      <p v-if="sessions.length">Próximas {{ sessions.length }} sesiones (máximo 16):</p>
      <div v-for="session in sessions" :key="`${session.scheduleId}-${session.originalDate}`" class="session-row">
        <div class="session-detail">
          <strong>{{ session.originalDate }}</strong>
          <span v-if="session.exception?.action === 'CANCEL'">Suspendida</span>
          <span v-else-if="session.exception?.action === 'RESCHEDULE'">
            Reprogramada al {{ session.exception.newDate }} · {{ session.exception.newStartTime }}–{{ session.exception.newEndTime }}
          </span>
          <span v-else>{{ session.startTime }}–{{ session.endTime }}</span>
        </div>
        <div v-if="Number(activity.enrollmentCount) === 0" class="session-actions">
          <button v-if="session.exception?.action !== 'CANCEL'" type="button" :disabled="saving" @click="begin(session, 'CANCEL')">Suspender</button>
          <button type="button" :disabled="saving" @click="begin(session, 'RESCHEDULE')">
            {{ session.exception?.action === 'CANCEL' ? 'Reactivar / reprogramar' : session.exception ? 'Reprogramar otra vez' : 'Reprogramar' }}
          </button>
        </div>
      </div>
      <form v-if="selected" class="session-form" @submit.prevent="submit">
        <strong>{{ selected.action === 'CANCEL' ? 'Suspender' : 'Reprogramar' }} sesión del {{ selected.originalDate }}</strong>
        <label>Motivo <textarea v-model="form.reason" required maxlength="500" minlength="5" rows="2" /></label>
        <template v-if="selected.action === 'RESCHEDULE'">
          <label>Nueva fecha <input v-model="form.newDate" type="date" required /></label>
          <label>Inicio <input v-model="form.newStartTime" type="time" required /></label>
          <label>Término <input v-model="form.newEndTime" type="time" required /></label>
        </template>
        <div class="session-actions">
          <button type="submit" :disabled="saving">{{ saving ? 'Guardando...' : 'Confirmar cambio' }}</button>
          <button type="button" :disabled="saving" @click="selected = null">Volver</button>
        </div>
      </form>
      <div v-if="exceptions.length" class="history">
        <strong>Historial de excepciones</strong>
        <div v-for="e in exceptions" :key="e.id" class="history-entry">
          <p>{{ e.originalDate }} · Estado actual: {{ e.action === 'CANCEL' ? 'Suspendida' : 'Reprogramada' }} · {{ e.reason }}</p>
          <p v-for="revision in e.revisions || []" :key="revision.id" class="history-revision">
            {{ revision.changedAt?.slice(0, 10) }} · {{ revision.previousAction }} → {{ revision.newAction }} · {{ revision.reason }}
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.occurrence-panel { margin: .75rem 0; border: 1px solid var(--color-border); border-radius: .75rem; padding: .75rem; background: var(--color-surface-soft); }
.toggle { background: var(--color-primary-soft); color: var(--color-primary-strong); font-weight: 750; border-radius: .5rem; padding: .65rem .8rem; border: 1px solid var(--color-border); cursor: pointer; }
.inner { margin-top: .75rem; display: grid; gap: .6rem; }
.inner p { margin: 0; }
.session-row { display: flex; justify-content: space-between; align-items: center; gap: .75rem; border-bottom: 1px solid var(--color-border); padding: .65rem 0; }
.session-detail { display: grid; gap: .25rem; }
.session-detail span { font-size: .85rem; color: var(--color-text-muted); }
.session-actions { display: flex; gap: .4rem; flex-wrap: wrap; }
.session-actions button, .session-form button { padding: .45rem .65rem; border-radius: .4rem; border: 1px solid var(--color-border); cursor: pointer; background: var(--color-surface); color: var(--color-text); }
.session-form { display: grid; gap: .6rem; border: 1px solid var(--color-border); padding: .8rem; border-radius: .5rem; }
.session-form label { display: grid; gap: .25rem; font-weight: 650; }
.session-form input,.session-form textarea { padding: .55rem; border: 1px solid var(--color-border); border-radius: .4rem; background: var(--color-surface); color: var(--color-text); }
.error { color: var(--color-error); font-weight: 700; }
.success { color: var(--color-success); }
.notice { color: var(--color-warning); }
 .history-revision { padding-left: 1rem; opacity: .85; }
.history-entry { margin-top: .55rem; }
.history { font-size: .82rem; border-top: 1px solid var(--color-border); padding-top: .6rem; }
@media(max-width:650px){.session-row{align-items:stretch;flex-direction:column}}
</style>
