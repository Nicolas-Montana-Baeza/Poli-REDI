import {
  msalInstance,
  loginRequest,
  apiTokenRequest
} from './msalConfig'

let initPromise = null

export async function initializeAuth() {
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

export async function logout() {
  await initializeAuth()
  clearMsalCache()

  return msalInstance.logoutRedirect({
    postLogoutRedirectUri: 'http://localhost:5173'
  })
}

export async function getCurrentAccount() {
  await initializeAuth()
  return msalInstance.getActiveAccount()
}

export async function isAuthenticated() {
  const account = await getCurrentAccount()
  return !!account
}

export async function getAccessToken() {
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
