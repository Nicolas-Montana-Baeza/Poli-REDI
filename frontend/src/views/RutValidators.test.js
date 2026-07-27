import { describe, expect, it } from 'vitest'

import { hasRut, normalizeRut } from '@/utils/validators'

describe('registered RUT semantics', () => {
  it('accepts canonical and legacy no-separator valid values', () => {
    expect(hasRut('12.345.678-5')).toBe(true)
    expect(normalizeRut('123456785')).toBe('12345678-5')
    expect(hasRut('123456785')).toBe(true)
  })

  it('treats whitespace, malformed legacy data and wrong verifier as absent', () => {
    expect(hasRut('   ')).toBe(false)
    expect(hasRut('legacy-value')).toBe(false)
    expect(hasRut('12345678-K')).toBe(false)
  })
})
