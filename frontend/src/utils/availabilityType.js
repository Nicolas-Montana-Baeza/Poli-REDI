const TYPE_DEFINITIONS = {
  reservation: {
    key: 'reservation',
    label: 'Reserva individual',
    family: 'reservation',
    tone: 'reservation'
  },
  group: {
    key: 'group',
    label: 'Reserva grupal',
    family: 'reservation',
    tone: 'group'
  },
  openUse: {
    key: 'open-use',
    label: 'Uso libre',
    family: 'open-use',
    tone: 'open-use'
  },
  workshop: {
    key: 'workshop',
    label: 'Taller',
    family: 'workshop',
    tone: 'workshop'
  },
  class: {
    key: 'class',
    label: 'Clase',
    family: 'institutional',
    tone: 'institutional'
  },
  training: {
    key: 'training',
    label: 'Entrenamiento',
    family: 'institutional',
    tone: 'institutional'
  },
  championship: {
    key: 'championship',
    label: 'Campeonato',
    family: 'institutional',
    tone: 'institutional'
  },
  event: {
    key: 'event',
    label: 'Evento',
    family: 'institutional',
    tone: 'institutional'
  },
  institutional: {
    key: 'institutional',
    label: 'Actividad institucional',
    family: 'institutional',
    tone: 'institutional'
  }
}

const SCHEDULED_TYPE_KEYS = {
  CLASS: 'class',
  WORKSHOP: 'workshop',
  TRAINING: 'training',
  CHAMPIONSHIP: 'championship',
  EVENT: 'event',
  OTHER: 'institutional'
}

const DISPLAY_PREFIX_BY_TYPE = {
  workshop: 'Taller',
  class: 'Clase',
  training: 'Entrenamiento',
  championship: 'Campeonato',
  event: 'Evento',
  institutional: 'Actividad institucional'
}

export const AVAILABILITY_TYPE_ORDER = [
  'reservation',
  'group',
  'open-use',
  'workshop',
  'class',
  'training',
  'championship',
  'event',
  'institutional'
]

const copyDefinition = (definition) => ({ ...definition })

export const getAvailabilityType = (item = null, resource = null) => {
  if (item?.isWorkshop === true) {
    return copyDefinition(TYPE_DEFINITIONS.workshop)
  }

  if (item?.isScheduledActivity === true ||
      item?.availabilityKind === 'SCHEDULED_ACTIVITY') {
    const scheduledType = String(item.activityType || '')
      .trim()
      .toUpperCase()
    const definitionKey =
      SCHEDULED_TYPE_KEYS[scheduledType] || 'institutional'

    return copyDefinition(TYPE_DEFINITIONS[definitionKey])
  }

  if (resource?.reservationMode === 'OPEN_USE') {
    return copyDefinition(TYPE_DEFINITIONS.openUse)
  }

  if (item?.availabilityKind === 'GROUP_RESERVATION') {
    return copyDefinition(TYPE_DEFINITIONS.group)
  }

  if (item?.targetParticipants !== null &&
      item?.targetParticipants !== undefined) {
    return copyDefinition(TYPE_DEFINITIONS.group)
  }

  return copyDefinition(TYPE_DEFINITIONS.reservation)
}

const escapeRegExp = (value) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

/**
 * Returns a presentation-only title for an availability item.
 *
 * Type comes exclusively from structural fields. A canonical leading prefix
 * is removed only when it matches that structured type and is followed by a
 * colon or dash. The source item is never mutated.
 */
export const getAvailabilityDisplayTitle = (
  item = null,
  fallback = 'Reserva'
) => {
  const originalTitle = String(item?.title || '').trim()

  if (!originalTitle) {
    return fallback
  }

  const type = getAvailabilityType(item)
  const canonicalPrefix = DISPLAY_PREFIX_BY_TYPE[type.key]

  if (!canonicalPrefix) {
    return originalTitle
  }

  const prefixPattern = new RegExp(
    `^\\s*${escapeRegExp(canonicalPrefix)}\\s*(?::|-)\\s*`,
    'i'
  )
  const displayTitle = originalTitle.replace(prefixPattern, '').trim()

  return displayTitle || fallback
}
