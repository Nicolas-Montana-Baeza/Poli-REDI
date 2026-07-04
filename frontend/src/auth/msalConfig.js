import { PublicClientApplication } from '@azure/msal-browser'

const tenantId = import.meta.env.VITE_ENTRA_TENANT_ID
const clientId = import.meta.env.VITE_ENTRA_CLIENT_ID
const redirectUri = import.meta.env.VITE_ENTRA_REDIRECT_URI
const apiScope = import.meta.env.VITE_ENTRA_API_SCOPE

export const msalConfig = {
  auth: {
    clientId,
    authority: `https://login.microsoftonline.com/${tenantId}`,
    redirectUri,
    postLogoutRedirectUri: 'http://localhost:5173'
  },
  cache: {
    cacheLocation: 'localStorage',
    storeAuthStateInCookie: false
  }
}

export const loginRequest = {
  scopes: ['openid', 'profile', 'email']
}

export const apiTokenRequest = {
  scopes: [apiScope]
}

export const msalInstance = new PublicClientApplication(msalConfig)
