import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { fakeClient, hangingConvert, heicFile } from '../test/fakeClient'
import { MAX_UPLOAD_BYTES } from '../lib/file'
import { useConvert } from './useConvert'

function oversizedFile(name = 'big.heic'): File {
  const file = new File([''], name)
  Object.defineProperty(file, 'size', { value: MAX_UPLOAD_BYTES + 1 })
  return file
}

describe('addFiles', () => {
  it('validates files on add (FR-1)', () => {
    const { result } = renderHook(() => useConvert(fakeClient()))

    act(() => {
      result.current.addFiles([heicFile(), new File([''], 'note.txt'), oversizedFile()])
    })

    const [ok, unsupported, tooLarge] = result.current.entries
    expect(ok.status).toBe('waiting')
    expect(unsupported.status).toBe('error')
    expect(unsupported.error).toContain('Not a HEIC/HEIF file')
    expect(tooLarge.status).toBe('error')
    expect(tooLarge.error).toContain('upload limit')
  })
})

describe('convert', () => {
  it('converts every file into every selected format', async () => {
    const { result } = renderHook(() => useConvert(fakeClient()))
    act(() => {
      result.current.addFiles([heicFile('a.heic'), heicFile('b.heic')])
    })

    await act(() => result.current.convert(['jpg', 'png'], 90))

    for (const entry of result.current.entries) {
      expect(entry.status).toBe('done')
      expect(entry.results.map((r) => r.format)).toEqual(['jpg', 'png'])
    }
    expect(result.current.entries[0].results[0].filename).toBe('a.jpg')
    expect(result.current.entries[0].results[0].url).toMatch(/^blob:/)
    expect(result.current.converting).toBe(false)
  })

  it('is fail-soft: one failure does not stop the others (FR-3)', async () => {
    const client = fakeClient()
    const okConvert = client.convert
    const failing = fakeClient({
      convert: async (req, options) => {
        // 2ファイル目(内容 "fail")だけ失敗させる
        if (new TextDecoder().decode(req.image).includes('fail')) {
          throw new Error('decode error')
        }
        return okConvert(req, options)
      },
    })
    const { result } = renderHook(() => useConvert(failing))
    act(() => {
      result.current.addFiles([heicFile('ok.heic', 'good'), heicFile('ng.heic', 'fail')])
    })

    await act(() => result.current.convert(['jpg'], 90))

    const [ok, ng] = result.current.entries
    expect(ok.status).toBe('done')
    expect(ng.status).toBe('error')
    expect(ng.error).toContain('decode error')
  })

  it('marks in-flight files as canceled when cancel() is called', async () => {
    const client = fakeClient({ convert: hangingConvert() })
    const { result } = renderHook(() => useConvert(client))
    act(() => {
      result.current.addFiles([heicFile()])
    })

    let done: Promise<void>
    act(() => {
      done = result.current.convert(['jpg'], 90)
    })
    await waitFor(() => expect(result.current.entries[0].status).toBe('converting'))

    act(() => result.current.cancel())
    await act(() => done)

    expect(result.current.entries[0].status).toBe('error')
    expect(result.current.entries[0].error).toBe('Canceled')
    expect(result.current.converting).toBe(false)
  })

  it('skips invalid entries', async () => {
    const convert = vi.fn(fakeClient().convert)
    const { result } = renderHook(() => useConvert(fakeClient({ convert })))
    act(() => {
      result.current.addFiles([new File([''], 'note.txt')])
    })

    await act(() => result.current.convert(['jpg'], 90))

    expect(convert).not.toHaveBeenCalled()
  })
})

describe('removeEntry', () => {
  it('revokes object URLs of the removed entry (FR-4)', async () => {
    const revoke = vi.spyOn(URL, 'revokeObjectURL')
    const { result } = renderHook(() => useConvert(fakeClient()))
    act(() => {
      result.current.addFiles([heicFile()])
    })
    await act(() => result.current.convert(['jpg'], 90))
    const url = result.current.entries[0].results[0].url

    act(() => result.current.removeEntry(result.current.entries[0].id))

    expect(result.current.entries).toHaveLength(0)
    expect(revoke).toHaveBeenCalledWith(url)
    revoke.mockRestore()
  })
})
