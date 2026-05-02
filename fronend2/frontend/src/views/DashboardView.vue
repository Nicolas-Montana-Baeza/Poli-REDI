<script setup>
</script>

<template>
  <div>
    <h2>Dashboard</h2>

    <div class="cards">
      <div class="card">Cancha 1</div>
      <div class="card">Cancha 2</div>
      <div class="card">Cancha 3</div>
      <div class="card">Gimnasio</div>
    </div>
  </div>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin-top: 20px;
}

.card {
  background: white;
  padding: 30px;
  border-radius: 12px;
  text-align: center;
  box-shadow: 0 2px 6px rgba(0,0,0,0.1);
}
</style>