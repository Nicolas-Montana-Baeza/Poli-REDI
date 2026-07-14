export const BUSINESS_TIME_ZONE =
  import.meta.env?.VITE_APP_TIMEZONE || 'America/Santiago'

const businessPartsFormatter = new Intl.DateTimeFormat('en-CA', {
  timeZone: BUSINESS_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hourCycle: 'h23'
})

const wallDateTimePattern =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?(?:\.(\d+))?$/

const getParts = (date) => {
  const parts = {}

  businessPartsFormatter.formatToParts(date).forEach((part) => {
    if (part.type !== 'literal') {
      parts[part.type] = Number(part.value)
    }
  })

  return parts
}

const getOffsetMilliseconds = (date) => {
  const parts = getParts(date)
  const wallAsUtc = Date.UTC(
    parts.year,
    parts.month - 1,
    parts.day,
    parts.hour,
    parts.minute,
    parts.second
  )
  const instantWithoutMilliseconds =
    Math.floor(date.getTime() / 1000) * 1000

  return wallAsUtc - instantWithoutMilliseconds
}

const wallPartsToDate = ({
  year,
  month,
  day,
  hour,
  minute,
  second,
  millisecond
}) => {
  const wallAsUtc = Date.UTC(
    year,
    month - 1,
    day,
    hour,
    minute,
    second,
    millisecond
  )

  let offset = getOffsetMilliseconds(new Date(wallAsUtc))
  let result = new Date(wallAsUtc - offset)
  const correctedOffset = getOffsetMilliseconds(result)

  if (correctedOffset !== offset) {
    offset = correctedOffset
    result = new Date(wallAsUtc - offset)
  }

  return result
}

export const parseBusinessDateTime = (value) => {
  if (!value) {
    return null
  }

  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : new Date(value.getTime())
  }

  const text = String(value).trim()
  const wallMatch = text.match(wallDateTimePattern)

  if (wallMatch) {
    const [, year, month, day, hour, minute, second = '0', fraction = '0'] =
      wallMatch

    return wallPartsToDate({
      year: Number(year),
      month: Number(month),
      day: Number(day),
      hour: Number(hour),
      minute: Number(minute),
      second: Number(second),
      millisecond: Number(fraction.slice(0, 3).padEnd(3, '0'))
    })
  }

  const parsed = new Date(text)

  return Number.isNaN(parsed.getTime()) ? null : parsed
}

export const getBusinessDateKey = (value = new Date()) => {
  const date = parseBusinessDateTime(value)
  if (!date) {
    return ''
  }

  const parts = getParts(date)

  return [
    parts.year,
    String(parts.month).padStart(2, '0'),
    String(parts.day).padStart(2, '0')
  ].join('-')
}

export const getBusinessMinutes = (value) => {
  const date = parseBusinessDateTime(value)
  if (!date) {
    return null
  }

  const parts = getParts(date)

  return parts.hour * 60 + parts.minute
}

export const formatBusinessTime = (value) => {
  const minutes = getBusinessMinutes(value)
  if (minutes === null) {
    return ''
  }

  const hour = Math.floor(minutes / 60)
  const minute = minutes % 60

  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

export const formatBusinessDate = (value, options) => {
  const date = parseBusinessDateTime(value)
  if (!date) {
    return ''
  }

  return date.toLocaleDateString('es-CL', {
    ...options,
    timeZone: BUSINESS_TIME_ZONE
  })
}
