import { PublicClientApplication } from '@azure/msal-browser'

export const msalConfig = {
  auth: {
    clientId: import.meta.env.VITE_ENTRA_CLIENT_ID,
    authority: `https://login.microsoftonline.com/${import.meta.env.VITE_ENTRA_TENANT_ID}`,
    redirectUri: import.meta.env.VITE_ENTRA_REDIRECT_URI,
    postLogoutRedirectUri: 'http://localhost:5173'
  },
  cache: {
    cacheLocation: 'sessionStorage',
    storeAuthStateInCookie: false
  }
}

export const loginRequest = {
  scopes: ['openid', 'profile', 'email']
}

export const apiTokenRequest = {
  scopes: [import.meta.env.VITE_ENTRA_API_SCOPE]
}

export const msalInstance = new PublicClientApplication(msalConfig)