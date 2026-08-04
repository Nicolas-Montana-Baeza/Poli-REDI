<script setup>
import { computed } from 'vue'

import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const props = defineProps({
  workshop: { type: Object, default: null },
  loading: { type: Boolean, default: false }
})
const emit = defineEmits(['confirm', 'cancel'])

const title = computed(() => props.workshop
  ? `¿Desinscribirte de ${props.workshop.title}?`
  : 'Desinscribirte del taller'
)
const message = computed(() => props.workshop
  ? `Tu cupo en ${props.workshop.title} quedará disponible para otra persona. Podrás volver a inscribirte solo si el taller sigue activo, tiene cupos y no se cruza con otro taller.`
  : ''
)
</script>

<template>
  <ConfirmModal
    :show="Boolean(workshop)"
    :title="title"
    :message="message"
    confirm-text="Sí, desinscribirme"
    cancel-text="Mantener inscripción"
    variant="danger"
    destructive
    :loading="loading"
    @confirm="emit('confirm')"
    @cancel="emit('cancel')"
  />
</template>
