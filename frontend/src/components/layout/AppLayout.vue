<script setup>
import HeaderBar from './HeaderBar.vue'
</script>

<template>
  <div class="layout">

    <HeaderBar />

    <main class="content">

      <router-view />

    </main>

  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;

  background: #f8fafc;

  display: flex;
  flex-direction: column;
}

/* CONTENT */
.content {
  flex: 1;

  padding: 24px;
}

/* MOBILE */
@media (max-width: 768px) {
  .content {
    padding: 16px;
  }
}
</style>