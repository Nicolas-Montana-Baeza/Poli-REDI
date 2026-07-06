import {
  msalInstance,
  loginRequest,
  apiTokenRequest,
  postLogoutRedirectUri
} from './msalConfig'

const DEV_ACCOUNT_KEY = 'poli_redi_dev_account'
let initPromise = null

export function isDevAuthEnabled() {
  return import.meta.env.DEV || import.meta.env.VITE_DEV_AUTH_ENABLED === 'true'
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
  if (getDevAccount()) {
    return getDevAccount()
  }

  if (!initPromise) {
    initPromise = msalInstance.initialize()
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
      })
      .catch(() => {
        return null
      })
  }

  return initPromise
}

export async function login(redirectPath = '/') {
  await initializeAuth()

  const account = msalInstance.getActiveAccount()

  if (account) {
    return account
  }

  sessionStorage.setItem('redirectAfterLogin', redirectPath)

  return msalInstance.loginRedirect(loginRequest)
}

export function loginLocal({ email, fullName }) {
  const normalizedEmail = String(email || '').trim().toLowerCase()
  const name = String(fullName || normalizedEmail).trim()

  if (!normalizedEmail) {
    throw new Error('Ingresa un correo.')
  }

  const account = {
    local: true,
    username: normalizedEmail,
    name,
    resetRutOnNextLoad: normalizedEmail !== 'admin@universidad.cl'
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

  await initializeAuth()
  clearMsalCache()

  return msalInstance.logoutRedirect({
    postLogoutRedirectUri
  })
}

export async function getCurrentAccount() {
  const devAccount = getDevAccount()

  if (devAccount) {
    return devAccount
  }

  await initializeAuth()
  return msalInstance.getActiveAccount()
}

export async function isAuthenticated() {
  const account = await getCurrentAccount()
  return !!account
}

export async function getAccessToken() {
  if (getDevAccount()) {
    return ''
  }

  await initializeAuth()

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
