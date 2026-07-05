import axios from 'axios'
import {
  getAccessToken,
  getDevAuthHeaders
} from '../auth/authService'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000/api',
  timeout: Number(import.meta.env.VITE_API_TIMEOUT_MS || 30000),
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(async (config) => {
  Object.assign(config.headers, getDevAuthHeaders())

  try {
    const token = await getAccessToken()

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  } catch {
    return config
  }
})

export default api
