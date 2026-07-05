<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from './components/layout/AppLayout.vue'

const route = useRoute()

const isPublicRoute = computed(() => route.meta.public === true)
</script>

<template>
  <RouterView v-if="isPublicRoute" />
  <AppLayout v-else />
</template>
