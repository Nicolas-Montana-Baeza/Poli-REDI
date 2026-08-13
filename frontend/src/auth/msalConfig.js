import { PublicClientApplication } from '@azure/msal-browser'

const tenantId = import.meta.env.VITE_ENTRA_TENANT_ID
const clientId = import.meta.env.VITE_ENTRA_CLIENT_ID
const apiScope = import.meta.env.VITE_ENTRA_API_SCOPE

const appOrigin = window.location.origin

export const postLogoutRedirectUri = `${appOrigin}/login`

export const msalConfig = {
  auth: {
    clientId,
    authority: `https://login.microsoftonline.com/${tenantId}`,
    redirectUri: `${appOrigin}/auth/callback`,
    postLogoutRedirectUri,
    navigateToLoginRequestUrl: false
  },

  cache: {
    cacheLocation: 'sessionStorage',
    storeAuthStateInCookie: false
  }
}

export const loginRequest = {
  scopes: [
    'openid',
    'profile',
    'email',
    apiScope
  ]
}

export const apiTokenRequest = {
  scopes: [apiScope]
}

export const msalInstance = new PublicClientApplication(msalConfig)
