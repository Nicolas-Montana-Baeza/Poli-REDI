<script setup>
import Sidebar from './Sidebar.vue'
import HeaderBar from './HeaderBar.vue'
import DashboardView from '../../views/DashboardView.vue'
</script>

<template>
  <div class="layout">
    
    <Sidebar />

    <div class="main">
      <HeaderBar />

      <div class="content">
        <DashboardView />
      </div>
    </div>

  </div>
</template>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
}

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.content {
  padding: 20px;
  overflow-y: auto;
}
</style>