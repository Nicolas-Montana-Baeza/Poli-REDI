
<template>
  <aside class="sidebar">
    <h1 class="logo">Poli REDI</h1>
    <nav>
      <button @click="$emit('home')">🏠 Inicio</button>
      <button @click="$emit('reserve')">📅 Reservar</button>
      <button @click="$emit('my-reservations')">📖 Mis reservas</button>
      <button @click="$emit('availability')">🕒 Disponibilidad</button>
      <button @click="$emit('history')">💬 Historial</button>
      <button @click="$emit('contact')">📞 Contacto</button>
      <button @click="$emit('go-stats')">📊 Estadísticas</button>
      <button @click="$emit('settings')">⚙️ Configuración</button>
      <button @click="$emit('help')">❓ Ayuda</button>
      <button @click="$emit('logout')">🚪 Cerrar sesión</button>
      <div class="footer">
        <p>© 2024 Poli-REDI. Todos los derechos reservados.</p>
      </div>
    </nav>
  </aside>
</template>

<script setup>
// logic later
</script>

<style scoped>
.sidebar {
  width: 240px;
  min-height: 100vh;
  background: white;
  box-shadow: 4px 0 12px rgba(0,0,0,0.08);
  padding: 24px;
}

.logo {
  margin-bottom: 30px;
  color: #4caf50;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

button {
  border: none;
  background: #f5f5f5;
  padding: 14px;
  border-radius: 12px;
  text-align: left;
  cursor: pointer;
  font-size: 1rem;
}

button:hover {
  background: #eaeaea;
}
</style>