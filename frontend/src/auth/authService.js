import { isMvp1Scope } from '@/config/appScope'

const DEV_ACCOUNT_KEY = 'poli_redi_dev_account'
const AUTH_PUBLIC_PATHS = [
  '/login',
  '/auth/callback',
  '/blocked'
]
let initPromise = null
let msalModulePromise = null

const getMsalModule = () => {
  if (isMvp1Scope()) {
    return Promise.resolve(null)
  }

  if (!msalModulePromise) {
    msalModulePromise = import('./msalConfig')
  }

  return msalModulePromise
}

export function getSafeRedirectPath(value) {
  const candidate = String(value || '').trim()

  if (!candidate || !candidate.startsWith('/') || candidate.startsWith('//')) {
    return '/'
  }

  try {
    const parsed = new URL(candidate, 'https://poliredi.local')

    if (parsed.origin !== 'https://poliredi.local') {
      return '/'
    }

    const isAuthPublicPath = AUTH_PUBLIC_PATHS.some((path) => {
      return parsed.pathname === path || parsed.pathname.startsWith(`${path}/`)
    })

    if (isAuthPublicPath) {
      return '/'
    }

    return `${parsed.pathname}${parsed.search}${parsed.hash}`
  } catch {
    return '/'
  }
}

export function isDevAuthEnabled() {
  return isMvp1Scope() ||
    import.meta.env.DEV ||
    import.meta.env.VITE_DEV_AUTH_ENABLED === 'true'
}

export function getDevAccount() {
  if (!isDevAuthEnabled()) {
    return null
  }

  const rawAccount = localStorage.getItem(DEV_ACCOUNT_KEY)

  if (!rawAccount) {
    return null
  }

  try {
    return JSON.parse(rawAccount)
  } catch {
    localStorage.removeItem(DEV_ACCOUNT_KEY)
    return null
  }
}

export function getDevAuthHeaders() {
  const account = getDevAccount()

  if (!account) {
    return {}
  }

  const headers = {
    'X-Dev-Auth-Email': account.username,
    'X-Dev-Auth-Name': account.name
  }

  if (account.resetRutOnNextLoad) {
    headers['X-Dev-Reset-Rut'] = 'true'
  }

  return headers
}

export async function initializeAuth() {
  const devAccount = getDevAccount()

  if (devAccount || isMvp1Scope()) {
    return devAccount
  }

  if (!initPromise) {
    initPromise = getMsalModule().then(({ msalInstance }) => msalInstance.initialize()
      .then(async () => {
        try {
          const response = await msalInstance.handleRedirectPromise()

          if (response?.account) {
            msalInstance.setActiveAccount(response.account)
            return response.account
          }
        } catch (error) {
          if (error.errorCode === 'no_token_request_cache_error') {
            clearMsalCache()
            return null
          }

          return null
        }

        const accounts = msalInstance.getAllAccounts()

        if (accounts.length > 0) {
          msalInstance.setActiveAccount(accounts[0])
          return accounts[0]
        }

        return null
      }))
      .catch(() => {
        return null
      })
  }

  return initPromise
}

export async function login(redirectPath = '/') {
  if (isMvp1Scope()) {
    throw new Error('El MVP1 local utiliza acceso de prueba, no Microsoft Entra.')
  }

  await initializeAuth()

  const {
    msalInstance,
    loginRequest
  } = await getMsalModule()

  const account = msalInstance.getActiveAccount()

  if (account) {
    return account
  }

  sessionStorage.setItem(
    'redirectAfterLogin',
    getSafeRedirectPath(redirectPath)
  )

  return msalInstance.loginRedirect(loginRequest)
}

export function loginLocal({ email, fullName, resetRut = false }) {
  const normalizedEmail = String(email || '').trim().toLowerCase()
  const name = String(fullName || normalizedEmail).trim()

  if (!normalizedEmail) {
    throw new Error('Ingresa un correo.')
  }

  const account = {
    local: true,
    username: normalizedEmail,
    name,
    resetRutOnNextLoad: Boolean(resetRut)
  }

  localStorage.setItem(DEV_ACCOUNT_KEY, JSON.stringify(account))

  return account
}

export function clearDevRutResetFlag() {
  const account = getDevAccount()

  if (!account?.resetRutOnNextLoad) {
    return
  }

  localStorage.setItem(DEV_ACCOUNT_KEY, JSON.stringify({
    ...account,
    resetRutOnNextLoad: false
  }))
}

export async function logout() {
  if (getDevAccount()) {
    localStorage.removeItem(DEV_ACCOUNT_KEY)
    return null
  }

  if (isMvp1Scope()) {
    return null
  }

  await initializeAuth()
  clearMsalCache()

  const {
    msalInstance,
    postLogoutRedirectUri
  } = await getMsalModule()

  return msalInstance.logoutRedirect({
    postLogoutRedirectUri
  })
}

export async function getCurrentAccount() {
  const devAccount = getDevAccount()

  if (devAccount) {
    return devAccount
  }

  if (isMvp1Scope()) {
    return null
  }

  await initializeAuth()
  const { msalInstance } = await getMsalModule()
  return msalInstance.getActiveAccount()
}

export async function isAuthenticated() {
  const account = await getCurrentAccount()
  return !!account
}

export async function getAccessToken() {
  if (getDevAccount() || isMvp1Scope()) {
    return ''
  }

  await initializeAuth()

  const {
    msalInstance,
    apiTokenRequest
  } = await getMsalModule()

  const account = msalInstance.getActiveAccount()

  if (!account) {
    throw new Error('No hay usuario autenticado')
  }

  const response = await msalInstance.acquireTokenSilent({
    ...apiTokenRequest,
    account
  })

  return response.accessToken
}

function clearMsalCache() {
  Object.keys(sessionStorage).forEach((key) => {
    if (key.includes('msal')) {
      sessionStorage.removeItem(key)
    }
  })

  Object.keys(localStorage).forEach((key) => {
    if (key.includes('msal')) {
      localStorage.removeItem(key)
    }
  })
}
