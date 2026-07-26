export const cleanRutInput = (value) => {
  const cleaned = String(value || '')
    .toUpperCase()
    .replace(/[^0-9K]/g, '')

  const body = cleaned
    .slice(0, -1)
    .replaceAll('K', '')

  return `${body}${cleaned.slice(-1)}`
}

export const formatRutInput = (value) => {
  const cleaned = cleanRutInput(value)

  if (cleaned.length <= 1) {
    return cleaned
  }

  return `${cleaned.slice(0, -1)}-${cleaned.slice(-1)}`
}

export const normalizeRut = (value) => {
  const cleaned = cleanRutInput(value)

  if (!cleaned) {
    return ''
  }

  if (cleaned.length < 2) {
    return cleaned
  }

  return `${cleaned.slice(0, -1)}-${cleaned.slice(-1)}`
}

export const isValidRut = (value) => {
  const normalized = normalizeRut(value)
  const [number, verifier] = normalized.split('-')

  if (!number || !verifier) {
    return false
  }

  if (!/^\d{7,8}$/.test(number)) {
    return false
  }

  if (!/^[0-9K]$/.test(verifier)) {
    return false
  }

  let sum = 0
  let multiplier = 2

  for (let index = number.length - 1; index >= 0; index -= 1) {
    sum += Number(number[index]) * multiplier
    multiplier = multiplier === 7 ? 2 : multiplier + 1
  }

  const expectedValue = 11 - (sum % 11)
  const expected =
    expectedValue === 11
      ? '0'
      : expectedValue === 10
        ? 'K'
        : String(expectedValue)

  return verifier === expected
}

export const hasRut = value => normalizeRut(value) !== ''
