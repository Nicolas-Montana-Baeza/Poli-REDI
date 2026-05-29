import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { initializeAuth } from './auth/authService'

import './style.css'

async function bootstrap() {
  await initializeAuth()

  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)
  app.use(router)
  app.mount('#app')
}

bootstrap()