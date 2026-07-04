// ファイルサイズ・ファイル名まわりの純粋関数。

// サーバーのリクエスト上限は64MiB。Connect JSONではbytesがbase64で約4/3倍に
// 膨張するため、元ファイルの実効上限はその3/4とする(FR-1)。
export const MAX_UPLOAD_BYTES = (64 * 1024 * 1024 * 3) / 4

export function isTooLarge(file: File): boolean {
  return file.size > MAX_UPLOAD_BYTES
}

// 拡張子でHEIC/HEIFかどうかを判定する(選択後の軽い検証)。
export function isHeicFile(file: File): boolean {
  return /\.(heic|heif)$/i.test(file.name)
}

// 表示用にバイト数を人間可読な単位へ変換する。
export function formatBytes(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

// 元ファイル名の拡張子を出力形式のものへ置き換える。
export function replaceExtension(name: string, format: string): string {
  const base = name.replace(/\.[^.]+$/, '')
  return `${base}.${format}`
}

// 同名ファイルが複数あるとき "a.jpg" → "a-2.jpg" のように一意化する。
export function uniqueNames(names: string[]): string[] {
  const seen = new Map<string, number>()
  return names.map((name) => {
    const count = (seen.get(name) ?? 0) + 1
    seen.set(name, count)
    if (count === 1) return name
    const dot = name.lastIndexOf('.')
    return dot === -1 ? `${name}-${count}` : `${name.slice(0, dot)}-${count}${name.slice(dot)}`
  })
}

const MIME_BY_FORMAT: Record<string, string> = {
  jpg: 'image/jpeg',
  png: 'image/png',
  webp: 'image/webp',
  tiff: 'image/tiff',
  bmp: 'image/bmp',
  gif: 'image/gif',
}

export function mimeOf(format: string): string {
  return MIME_BY_FORMAT[format] ?? 'application/octet-stream'
}
