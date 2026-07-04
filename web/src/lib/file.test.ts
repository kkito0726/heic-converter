import { describe, expect, it } from 'vitest'
import {
  MAX_UPLOAD_BYTES,
  formatBytes,
  isHeicFile,
  isTooLarge,
  mimeOf,
  replaceExtension,
  uniqueNames,
} from './file'

// 実際に巨大なバッファを確保せず、sizeだけを偽装したFileを作る。
function fileOfSize(size: number, name = 'a.heic'): File {
  const file = new File([''], name)
  Object.defineProperty(file, 'size', { value: size })
  return file
}

describe('isTooLarge', () => {
  it('accepts a file at the limit', () => {
    expect(isTooLarge(fileOfSize(MAX_UPLOAD_BYTES))).toBe(false)
  })
  it('rejects a file just over the limit', () => {
    expect(isTooLarge(fileOfSize(MAX_UPLOAD_BYTES + 1))).toBe(true)
  })
})

describe('isHeicFile', () => {
  it.each([
    ['photo.heic', true],
    ['photo.HEIC', true],
    ['photo.heif', true],
    ['photo.jpg', false],
    ['heic', false],
  ])('%s -> %s', (name, want) => {
    expect(isHeicFile(new File([''], name))).toBe(want)
  })
})

describe('formatBytes', () => {
  it.each([
    [500, '500 B'],
    [2048, '2.0 KB'],
    [5 * 1024 * 1024, '5.0 MB'],
  ])('%d -> %s', (size, want) => {
    expect(formatBytes(size)).toBe(want)
  })
})

describe('replaceExtension', () => {
  it('replaces the extension with the format', () => {
    expect(replaceExtension('IMG_001.heic', 'jpg')).toBe('IMG_001.jpg')
  })
  it('keeps dots inside the base name', () => {
    expect(replaceExtension('my.photo.heic', 'png')).toBe('my.photo.png')
  })
})

describe('uniqueNames', () => {
  it('keeps unique names as-is', () => {
    expect(uniqueNames(['a.jpg', 'b.jpg'])).toEqual(['a.jpg', 'b.jpg'])
  })
  it('suffixes duplicated names before the extension', () => {
    expect(uniqueNames(['a.jpg', 'a.jpg', 'a.jpg'])).toEqual(['a.jpg', 'a-2.jpg', 'a-3.jpg'])
  })
})

describe('mimeOf', () => {
  it('maps known formats', () => {
    expect(mimeOf('jpg')).toBe('image/jpeg')
    expect(mimeOf('png')).toBe('image/png')
  })
  it('falls back to octet-stream', () => {
    expect(mimeOf('avif')).toBe('application/octet-stream')
  })
})
