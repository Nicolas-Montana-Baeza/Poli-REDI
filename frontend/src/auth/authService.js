import {
  msalInstance,
  loginRequest,
  apiTokenRequest
} from './msalConfig'

let initPromise = null

export async function initializeAuth() {
  if (!initPromise) {
    initPromise = msalInstance.initialize()
      .then(() => msalInstance.handleRedirectPromise())
      .then((response) => {
        if (response?.account) {
          msalInstance.setActiveAccount(response.account)
          return response.account
        }

        const accounts = msalInstance.getAllAccounts()

        if (accounts.length > 0) {
          msalInstance.setActiveAccount(accounts[0])
          return accounts[0]
        }

        return null
      })
      .catch((error) => {
        console.error('Error inicializando MSAL:', error)
        throw error
      })
  }

  return initPromise
}

export async function login(redirectPath = '/') {
  await initializeAuth()

  sessionStorage.setItem('redirectAfterLogin', redirectPath)

  return msalInstance.loginRedirect(loginRequest)
}

export async function logout() {
  await initializeAuth()

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