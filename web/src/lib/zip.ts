import { zipSync, type Zippable } from 'fflate'
import { uniqueNames } from './file'

export interface ZipEntry {
  name: string
  data: Uint8Array
}

// 変換結果を無圧縮zipへまとめる(FR-4)。画像は再圧縮しても縮まないため
// level 0(store)で十分で、生成も速い。
export function makeZip(entries: ZipEntry[]): Uint8Array<ArrayBuffer> {
  const names = uniqueNames(entries.map((e) => e.name))
  const input: Zippable = Object.fromEntries(
    entries.map((e, i) => [names[i], [e.data, { level: 0 }] as const]),
  )
  // fflateは通常のArrayBufferで確保するため、BlobPartとして扱える型に絞る
  return zipSync(input) as Uint8Array<ArrayBuffer>
}
