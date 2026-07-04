// 変換対象ファイルと変換結果の状態モデル。

export type ConversionStatus = 'waiting' | 'converting' | 'done' | 'error'

// 1形式分の変換結果。urlは揮発性で、不要になったらrevokeObjectURLで解放する。
export interface ConvertedFile {
  format: string
  filename: string
  data: Uint8Array
  blob: Blob
  url: string
}

export interface FileEntry {
  id: string
  file: File
  // サーバーの実効上限(base64膨張込み)を超えている
  tooLarge: boolean
  // .heic / .heif 以外のファイルが選ばれた
  unsupported: boolean
  status: ConversionStatus
  results: ConvertedFile[]
  error?: string
}

// 変換リクエストを送れるエントリかどうか(FR-1の事前検証を通過している)。
export function isConvertible(entry: FileEntry): boolean {
  return !entry.unsupported && !entry.tooLarge
}
