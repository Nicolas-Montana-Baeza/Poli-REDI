<script setup>
</script>

<template>
  <div class="sidebar">
    <h2 class="logo">PoliREDI</h2>

    <nav>
      <ul>
        <li class="active">Inicio</li>
        <li>Disponibilidad</li>
        <li>Mis reservas</li>
        <li>Administración</li>
      </ul>
    </nav>
  </div>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: #1e3a8a; /* azul */
  color: white;
  padding: 20px;
}

.logo {
  margin-bottom: 30px;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
}

li:hover {
  background: rgba(255,255,255,0.1);
}

.active {
  background: #f97316; /* naranjo */
}
</style>