<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ParticipantsProgress from '@/components/ui/ParticipantsProgress.vue'
import { reservationsService } from '@/services/reservations.service'
const route=useRoute(),router=useRouter(),code=ref(String(route.params.code||'')),progress=ref(null),error=ref(''),busy=ref(false)
const messageFor=e=>({404:'El código no existe o ya no está disponible.',409:'Ya tienes una reserva activa en ese horario y no puedes confirmar esta participación.',410:'El plazo de confirmación ya venció.',403:'Debes tener una cuenta activa y RUT registrado.'}[e?.response?.status]||e?.response?.data?.error||'No se pudo completar la operación.')
const load=async()=>{error.value='';if(!code.value.trim())return;busy.value=true;try{progress.value=await reservationsService.getGroupProgress(code.value.trim());router.replace(`/join/${encodeURIComponent(code.value.trim())}`)}catch(e){progress.value=null;error.value=messageFor(e)}finally{busy.value=false}}
const change=async confirm=>{busy.value=true;error.value='';try{progress.value=confirm?await reservationsService.confirmGroup(code.value):await reservationsService.withdrawGroup(code.value)}catch(e){error.value=messageFor(e)}finally{busy.value=false}}
onMounted(()=>{if(code.value)load()})
</script>
<template>
  <main class="join-page">
    <header><h1>Unirse a una reserva grupal</h1><p>Ingresa el código compartido por quien creó la reserva.</p></header>
    <form @submit.prevent="load"><label for="join-code">Código de invitación</label><div><input id="join-code" v-model.trim="code" required autocomplete="off"><button :disabled="busy">Consultar</button></div></form>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <ParticipantsProgress v-if="progress" :progress="progress" :busy="busy" @confirm="change(true)" @withdraw="change(false)" />
  </main>
</template>
<style scoped>
.join-page{max-width:760px;margin:0 auto;padding:clamp(1rem,4vw,2rem);display:grid;gap:1.25rem}form{display:grid;gap:.5rem}form div{display:flex;gap:.5rem;flex-wrap:wrap}input{flex:1;min-width:14rem;min-height:44px;padding:.7rem}button{min-height:44px;padding:.7rem 1rem}.error{padding:1rem;border-radius:.75rem;background:#fee2e2;color:#991b1b}
</style>
