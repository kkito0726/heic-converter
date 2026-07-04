import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ConvertedFile } from '../types'
import { canShareFiles, isShareSupported, shareFiles, toShareFiles } from './share'

function stubShareAPI(share: (data?: ShareData) => Promise<void>) {
  vi.stubGlobal('navigator', {
    ...navigator,
    share,
    canShare: () => true,
  })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('isShareSupported / canShareFiles', () => {
  it('is false when the browser lacks the Web Share API (jsdom default)', () => {
    expect(isShareSupported()).toBe(false)
    expect(canShareFiles([new File(['x'], 'a.jpg')])).toBe(false)
  })

  it('is true when navigator.share and canShare exist', () => {
    stubShareAPI(async () => {})
    expect(isShareSupported()).toBe(true)
    expect(canShareFiles([new File(['x'], 'a.jpg')])).toBe(true)
  })

  it('is false for an empty file list', () => {
    stubShareAPI(async () => {})
    expect(canShareFiles([])).toBe(false)
  })
})

describe('shareFiles', () => {
  it('resolves true when sharing succeeds', async () => {
    const share = vi.fn(async () => {})
    stubShareAPI(share)
    const files = [new File(['x'], 'a.jpg')]
    await expect(shareFiles(files, 'title')).resolves.toBe(true)
    expect(share).toHaveBeenCalledWith({ files, title: 'title' })
  })

  it('resolves false when the user dismisses the share sheet', async () => {
    stubShareAPI(async () => {
      throw new DOMException('canceled', 'AbortError')
    })
    await expect(shareFiles([new File(['x'], 'a.jpg')])).resolves.toBe(false)
  })

  it('rethrows unexpected errors', async () => {
    stubShareAPI(async () => {
      throw new Error('boom')
    })
    await expect(shareFiles([new File(['x'], 'a.jpg')])).rejects.toThrow('boom')
  })
})

describe('toShareFiles', () => {
  it('converts results into named files with mime types', () => {
    const result: ConvertedFile = {
      format: 'jpg',
      filename: 'photo.jpg',
      data: new Uint8Array([1]),
      blob: new Blob([new Uint8Array([1])]),
      url: 'blob:mock',
    }
    const files = toShareFiles([result])
    expect(files[0].name).toBe('photo.jpg')
    expect(files[0].type).toBe('image/jpeg')
  })
})
