export const resolveMvpScope = (value) => {
  const scope = String(value || '').trim().toLowerCase()

  switch (scope) {
    case 'mvp2':
      return 'mvp2'

    case 'full':
      return 'full'

    default:
      // El alcance por defecto se mantiene deliberadamente en MVP1.
      //
      // Esto evita habilitar funcionalidades nuevas por un error de
      // configuración o por una variable de entorno desconocida.
      return 'mvp1'
  }
}

export const MVP_SCOPE = resolveMvpScope(
  import.meta.env?.VITE_MVP_SCOPE
)

export const isMvp1Scope = () => MVP_SCOPE === 'mvp1'
export const isMvp2Scope = () => MVP_SCOPE === 'mvp2'
export const isFullScope = () => MVP_SCOPE === 'full'

export const getFeaturesForScope = (scope) => {
  const resolvedScope = resolveMvpScope(scope)

  const hasMvp2 = (
    resolvedScope === 'mvp2' ||
    resolvedScope === 'full'
  )

  const full = resolvedScope === 'full'

  return {
    // ----------------------------------------------------------
    // MVP2
    // ----------------------------------------------------------
    //
    // Las reservas grupales ya fueron migradas y validadas sobre
    // PostgreSQL, por lo que pueden habilitarse tanto en MVP2 como
    // en FULL.
    groupReservations: hasMvp2,

    // La resolución administrativa de conflictos utiliza exclusivamente
    // el backend institucional PostgreSQL y pertenece al alcance MVP2.
    schedulingConflictAdministration: hasMvp2,

    // Los talleres institucionales ya fueron migrados al modelo
    // PostgreSQL de programación institucional y forman parte de MVP2.
    workshops: hasMvp2,

    // ----------------------------------------------------------
    // FULL / legacy
    // ----------------------------------------------------------
    //
    // Estas funcionalidades todavía no forman parte del alcance
    // incremental de MVP2.
    onlineAuth: full,
    notifications: full,
    resourceAdministration: full,
    policyAdministration: full,
    reports: full
  }
}

export const mvpFeatures = Object.freeze(
  getFeaturesForScope(MVP_SCOPE)
)
