import { unzipSync } from 'fflate'
import { describe, expect, it } from 'vitest'
import { makeZip } from './zip'

describe('makeZip', () => {
  it('produces a zip whose entries round-trip', () => {
    const zip = makeZip([
      { name: 'a.jpg', data: new Uint8Array([1, 2, 3]) },
      { name: 'b.png', data: new Uint8Array([4, 5]) },
    ])
    const unzipped = unzipSync(zip)
    expect(Object.keys(unzipped)).toEqual(['a.jpg', 'b.png'])
    expect(Array.from(unzipped['a.jpg'])).toEqual([1, 2, 3])
    expect(Array.from(unzipped['b.png'])).toEqual([4, 5])
  })

  it('deduplicates entry names', () => {
    const zip = makeZip([
      { name: 'a.jpg', data: new Uint8Array([1]) },
      { name: 'a.jpg', data: new Uint8Array([2]) },
    ])
    const unzipped = unzipSync(zip)
    expect(Object.keys(unzipped)).toEqual(['a.jpg', 'a-2.jpg'])
  })
})
