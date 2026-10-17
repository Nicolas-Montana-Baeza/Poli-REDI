const integer = (value, minimum = 0, maximum = Infinity) => {
  if (value === '' || value === null || value === undefined) {
    return null
  }

  const number = Number(value)

  return Number.isSafeInteger(number) &&
    number >= minimum &&
    number <= maximum
    ? number
    : null
}

export const parsePolicyDurations = (text) => {
  const parts = String(text).split(',').map((item) => item.trim())

  if (!parts.length || parts.some((part) => !/^[0-9]+$/.test(part))) {
    return null
  }

  const values = parts.map(Number)

  if (values.some((value) =>
    !Number.isSafeInteger(value) || value <= 0
  )) {
    return null
  }

  return [...new Set(values)].sort((a, b) => a - b)
}

export const validatePolicyDraft = (draft, durationsText, resources) => {
  const errors = []

  if (!draft || !Array.isArray(resources)) {
    return { payload: null, errors: ['Borrador no disponible.'] }
  }

  const fieldRules = {
    reservableWindowDays: [1, Infinity],
    requestFrequencyDays: [1, Infinity],
    confirmationDeadlineMinutes: [0, Infinity],
    minimumParticipants: [1, Infinity],
    openingMinute: [0, 1439],
    closingMinute: [1, 1439],
    slotIntervalMinutes: [1, 1440]
  }

  const payload = {}

  for (const [key, [min, max]] of Object.entries(fieldRules)) {
    payload[key] = integer(draft[key], min, max)

    if (payload[key] === null) {
      errors.push(`Valor inválido: ${key}.`)
    }
  }

  if (
    payload.openingMinute !== null &&
    payload.closingMinute !== null &&
    payload.openingMinute >= payload.closingMinute
  ) {
    errors.push('La apertura debe ser anterior al cierre.')
  }

  payload.allowedDurations = parsePolicyDurations(durationsText)

  if (!payload.allowedDurations?.length) {
    errors.push('Ingresa duraciones válidas separadas por coma.')
  }

  const ids = Array.isArray(draft.resourceIds)
    ? draft.resourceIds.map((id) => integer(id, 1))
    : []

  if (
    ids.length === 0 ||
    ids.includes(null) ||
    new Set(ids).size !== ids.length
  ) {
    errors.push('Selecciona recursos válidos, sin duplicados.')
  }

  payload.resourceIds = [...ids].sort((a, b) => a - b)

  const resourceMap = new Map(resources.map((resource) => [
    Number(resource.id),
    resource
  ]))

  for (const id of payload.resourceIds) {
    const resource = resourceMap.get(id)

    if (!resource || resource.isActive !== true) {
      errors.push(`El recurso #${id} no existe o está inactivo.`)
    }
  }

  if (!Array.isArray(draft.groupResources)) {
    errors.push('La configuración grupal es inválida.')
    payload.groupResources = []
  } else {
    payload.groupResources = []
    const seen = new Set()

    for (const rule of draft.groupResources) {
      const id = integer(rule?.resourceId, 1)
      const minimum = integer(rule?.minimumParticipants, 1)

      if (id === null || minimum === null || seen.has(id)) {
        errors.push('Los mínimos grupales deben ser positivos y únicos.')
        continue
      }

      seen.add(id)

      const resource = resourceMap.get(id)

      if (!payload.resourceIds.includes(id)) {
        errors.push(`El recurso grupal #${id} no está permitido.`)
      }

      if (
        !resource ||
        resource.isActive !== true ||
        resource.reservationMode !== 'RESERVABLE'
      ) {
        errors.push(`El recurso grupal #${id} no es reservable.`)
      } else if (
        resource.capacity === null ||
        resource.capacity === undefined ||
        minimum > Number(resource.capacity)
      ) {
        errors.push(`El mínimo grupal de ${resource.name} supera su capacidad.`)
      }

      payload.groupResources.push({
        resourceId: id,
        minimumParticipants: minimum
      })
    }

    payload.groupResources.sort((a, b) => a.resourceId - b.resourceId)
  }

  return {
    payload: errors.length === 0 ? payload : null,
    errors
  }
}

export const getPolicyChanges = (current, next, resources = []) => {
  if (!current || !next) return []

  const names = new Map(
    resources.map((resource) => [Number(resource.id), resource.name])
  )

  const descriptions = [
    ['reservableWindowDays', 'Ventana de reserva'],
    ['requestFrequencyDays', 'Frecuencia'],
    ['confirmationDeadlineMinutes', 'Plazo de confirmación'],
    ['minimumParticipants', 'Mínimo global'],
    ['openingMinute', 'Apertura'],
    ['closingMinute', 'Cierre'],
    ['slotIntervalMinutes', 'Intervalo']
  ]

  const changes = []

  const add = (label, before, after) => {
    if (before !== after) {
      changes.push({ label, before, after })
    }
  }

  for (const [key, label] of descriptions) {
    add(label, String(current[key]), String(next[key]))
  }

  const describeIds = (ids) =>
    [...(ids || [])]
      .sort((a, b) => a - b)
      .map((id) => names.get(id) || `#${id}`)
      .join(', ') || 'Ninguno'

  const describeGroups = (rules) =>
    [...(rules || [])]
      .sort((a, b) => a.resourceId - b.resourceId)
      .map((rule) =>
        `${names.get(rule.resourceId) || `#${rule.resourceId}`}: ${rule.minimumParticipants}`
      )
      .join(', ') || 'Ninguno'

  add(
    'Duraciones permitidas',
    [...(current.allowedDurations || [])].sort((a, b) => a - b).join(', '),
    [...next.allowedDurations].sort((a, b) => a - b).join(', ')
  )

  add(
    'Recursos permitidos',
    describeIds(current.resourceIds),
    describeIds(next.resourceIds)
  )

  add(
    'Confirmación grupal por recurso',
    describeGroups(current.groupResources),
    describeGroups(next.groupResources)
  )

  return changes
}
