export const resolveMvpScope = (value) => {
  return String(value || '').trim().toLowerCase() === 'full'
    ? 'full'
    : 'mvp1'
}

export const MVP_SCOPE = resolveMvpScope(
  import.meta.env?.VITE_MVP_SCOPE
)

export const isMvp1Scope = () => MVP_SCOPE === 'mvp1'

export const getFeaturesForScope = (scope) => {
  const full = resolveMvpScope(scope) === 'full'
  return {
    onlineAuth: full,
    notifications: full,
    workshops: full,
    groupReservations: full,
    resourceAdministration: full,
    policyAdministration: full,
    reports: full
  }
}

export const mvpFeatures = Object.freeze(
  getFeaturesForScope(MVP_SCOPE)
)
