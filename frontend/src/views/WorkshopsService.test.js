import { beforeEach, describe, expect, it, vi } from 'vitest'

const api = vi.hoisted(() => ({
  delete: vi.fn()
}))

vi.mock('@/services/api', () => ({ default: api }))

import { workshopsService } from '@/services/workshops.service'

describe('workshopsService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('desinscribe al usuario autenticado mediante el endpoint del taller', async () => {
    const updatedWorkshop = {
      id: 7,
      isEnrolled: false,
      enrolledCount: 11
    }
    api.delete.mockResolvedValueOnce({ data: updatedWorkshop })

    await expect(workshopsService.withdraw(7))
      .resolves.toEqual(updatedWorkshop)

    expect(api.delete).toHaveBeenCalledOnce()
    expect(api.delete).toHaveBeenCalledWith(
      '/workshops/7/enrollment'
    )
  })
})
