import { renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FAKE_FORMATS, fakeClient } from '../test/fakeClient'
import { useFormats } from './useFormats'

describe('useFormats', () => {
  it('loads formats from the server', async () => {
    const { result } = renderHook(() => useFormats(fakeClient()))
    expect(result.current.loading).toBe(true)

    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.formats).toEqual(FAKE_FORMATS)
    expect(result.current.error).toBeNull()
  })

  it('exposes an error message when the RPC fails', async () => {
    const client = fakeClient({
      listFormats: async () => {
        throw new Error('unavailable')
      },
    })
    const { result } = renderHook(() => useFormats(client))

    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.formats).toEqual([])
    expect(result.current.error).toContain('unavailable')
  })
})
