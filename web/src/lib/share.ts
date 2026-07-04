import type { ConvertedFile } from '../types'
import { mimeOf } from './file'

// Web Share API(OSの共有シート)のラッパ(FR-5)。
// クラウド保存・メール添付はここを経由してユーザーの手持ちアプリに委ねる。

// 変換結果をnavigator.shareへ渡せるFileに変換する。
export function toShareFiles(results: ConvertedFile[]): File[] {
  return results.map((r) => new File([r.blob], r.filename, { type: mimeOf(r.format) }))
}

// この環境にファイル共有APIが存在するか(ボタンの表示判定に使う)。
export function isShareSupported(): boolean {
  return (
    typeof navigator !== 'undefined' &&
    typeof navigator.share === 'function' &&
    typeof navigator.canShare === 'function'
  )
}

// このファイル群を実際に共有できるか。
export function canShareFiles(files: File[]): boolean {
  return isShareSupported() && files.length > 0 && navigator.canShare({ files })
}

// OSの共有シートを開く。ユーザーが共有をやめた場合はfalseを返し、エラーにしない。
export async function shareFiles(files: File[], title?: string): Promise<boolean> {
  try {
    await navigator.share({ files, title })
    return true
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return false
    throw error
  }
}
